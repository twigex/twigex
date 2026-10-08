// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/twigex/twigex/model"
)

type Server struct {
	Router          *mux.Router
	HttpServer      *http.Server
	FileServer      http.Handler
	NotificationHub *Hub

	License *model.License
}

func NewServer() (Server, error) {
	router := mux.NewRouter()

	httpServer := &http.Server{
		Handler: router,

		ReadHeaderTimeout: 5 * time.Second,

		IdleTimeout: 120 * time.Second,

		//Max size of header request 1MB
		MaxHeaderBytes: 1 << 20,
	}

	return Server{
		Router:          router,
		HttpServer:      httpServer,
		FileServer:      http.FileServer(http.Dir("./frontend/dist")),
		NotificationHub: newHub(),
	}, nil
}

func (srv *Server) Start(port string) error {

	srv.HttpServer.Addr = port
	err := srv.HttpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (srv *Server) StartTLS(cert string, key string) error {
	err := srv.HttpServer.ListenAndServeTLS(cert, key)
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (srv *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := srv.HttpServer.Shutdown(ctx)
	if err != nil {
		return err
	}

	return nil
}
