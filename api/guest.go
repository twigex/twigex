// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/twigex/twigex/model"
)

func (a *API) initGuest() {
	a.BaseRoutes.Guest.HandleFunc("/video/{token}", a.getGuestVideoLink).Methods("GET")
	a.BaseRoutes.Guest.HandleFunc("/video/{token}/token", a.getGuestVideoToken).Methods("POST")
	a.BaseRoutes.Guest.HandleFunc("/video/{token}/photo/{identity}", a.getGuestParticipantPhoto).Methods("GET")
}

func (a *API) getGuestVideoLink(w http.ResponseWriter, r *http.Request) {
	token := mux.Vars(r)["token"]

	link, meeting, appErr := a.app.ValidateVideoGuestLink(r.Context(), token)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	channel, err := a.app.Store.Channels.Get(link.ChannelID)
	if err != nil || channel == nil {
		respondAppError(w, r, model.NewAppError("channel.not_found", http.StatusNotFound))
		return
	}

	participants := 0
	if count, appErr := a.app.GetGuestVideoParticipantCount(meeting.ID); appErr == nil {
		participants = count
	}

	respondJSON(w, http.StatusOK, struct {
		ChannelName      string `json:"channel_name"`
		MeetingTitle     string `json:"meeting_title"`
		Status           string `json:"status"`
		ScheduledAt      *int64 `json:"scheduled_at"`
		Participants     int    `json:"participants"`
		RequiresPassword bool   `json:"requires_password"`
		Host             string `json:"host"`
	}{
		ChannelName:      channel.DisplayName,
		MeetingTitle:     meeting.Title,
		Status:           meeting.Status,
		ScheduledAt:      meeting.ScheduledAt,
		Participants:     participants,
		RequiresPassword: meeting.HasGuestPassword,
		Host:             *a.app.ConfigStore.Config.ChannelSettings.Host,
	})
}

func (a *API) getGuestParticipantPhoto(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	_, meeting, appErr := a.app.ValidateVideoGuestLink(r.Context(), params["token"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	userID, appErr := a.app.ResolveMeetingParticipant(r.Context(), meeting.ID, params["identity"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr := a.app.ServeUserPhoto(userID, w, r); appErr != nil {
		respondAppError(w, r, appErr)
	}
}

func (a *API) getGuestVideoToken(w http.ResponseWriter, r *http.Request) {
	token := mux.Vars(r)["token"]

	s := struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}{}

	if !decodeBody(w, r, &s) {
		return
	}

	videoToken, appErr := a.app.GetGuestVideoToken(r.Context(), token, s.Name, s.Password)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, struct {
		Token string `json:"token"`
		Host  string `json:"host"`
	}{
		Token: videoToken,
		Host:  *a.app.ConfigStore.Config.ChannelSettings.Host,
	})
}
