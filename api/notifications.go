// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/twigex/twigex/tlog"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     sameOrigin,
}

// sameOrigin compares whole hosts: a substring test would also accept
// "<our host>.example.com", letting any page open an authenticated socket.
func sameOrigin(r *http.Request) bool {
	origin, err := url.Parse(r.Header.Get("Origin"))
	if err != nil || origin.Host == "" {
		return false
	}

	if strings.EqualFold(origin.Host, r.Host) {
		return true
	}

	return origin.Host == "localhost:5173" && isLoopback(r.Host)
}

func isLoopback(host string) bool {
	name, _, err := net.SplitHostPort(host)
	if err != nil {
		name = host
	}

	if name == "localhost" {
		return true
	}

	ip := net.ParseIP(name)
	return ip != nil && ip.IsLoopback()
}

func (a *API) initNotifications() {
	a.BaseRoutes.Notifications.HandleFunc("", a.getUserNotifications)
	a.BaseRoutes.Notifications.HandleFunc("/{id}/resolve", a.reolveNotification).Methods("GET")
	a.BaseRoutes.Notifications.HandleFunc("/read", a.readNotification).Methods("PUT")
	a.BaseRoutes.Notifications.HandleFunc("/{id}/delete", a.deleteNotification).Methods("POST")
	a.BaseRoutes.Notifications.HandleFunc("/ws", a.wsNotifications)

	a.BaseRoutes.Notifications.Use(a.RequireSession)
	a.BaseRoutes.Notifications.Use(a.RequireCSRF)
}

func (a *API) wsNotifications(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			tlog.Errorw("Failed to upgrade websocket connection",
				"error", err,
			)
		}

		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.NewWebsocketConnection(conn, *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}
}

func (a *API) getUserNotifications(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		http.Error(w, appErr.Message, appErr.Status)
		return
	}

	notifications, appErr := a.app.GetUserNotifications(user.ID)
	if appErr != nil {
		http.Error(w, appErr.Message, appErr.Status)
		return
	}

	respondJSON(w, http.StatusOK, notifications)
}

func (a *API) reolveNotification(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		http.Error(w, appErr.Message, appErr.Status)
		return
	}

	data, appErr := a.app.ResolveNotification(params["id"], *user)
	if appErr != nil {
		http.Error(w, appErr.Message, appErr.Status)
		return
	}

	respondJSON(w, http.StatusOK, data)
}

func (a *API) readNotification(w http.ResponseWriter, r *http.Request) {
	s := struct {
		Read bool
		ID   []string
	}{}

	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		http.Error(w, appErr.Message, appErr.Status)
		return
	}

	appErr = a.app.ReadNotification(s.Read, s.ID, user.ID)
	if appErr != nil {
		http.Error(w, appErr.Message, appErr.Status)
		return
	}
}

func (a *API) deleteNotification(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		http.Error(w, appErr.Message, appErr.Status)
		return
	}

	appErr = a.app.DeleteNotification(params["id"], user.ID)
	if appErr != nil {
		http.Error(w, appErr.Message, appErr.Status)
		return
	}
}
