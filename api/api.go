// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/twigex/twigex/app"
	"github.com/twigex/twigex/config"
	"github.com/twigex/twigex/tlog"
)

type API struct {
	app        *app.App
	BaseRoutes *Routes
}

type Routes struct {
	Root    *mux.Router // ''
	APIRoot *mux.Router // 'api/'

	Files             *mux.Router // 'api/files'
	Office            *mux.Router // 'api/office'
	Config            *mux.Router // 'api/config'
	Collimato         *mux.Router // 'api/collimato'
	Channels          *mux.Router // 'api/channels'
	Activity          *mux.Router // 'api/activity'
	Users             *mux.Router // 'api/users'
	Notifications     *mux.Router // 'api/notifications'
	Settings          *mux.Router // 'api/settings'
	Metadata          *mux.Router // 'api/metadata'
	Search            *mux.Router // 'api/search'
	PublicLink        *mux.Router // 'api/link'
	Auth              *mux.Router // 'api/auth'
	WebSocket         *mux.Router // 'api/ws'
	PublicSharedFiles *mux.Router // 'api/public'
	License           *mux.Router // 'api/license'
	Workspaces        *mux.Router // 'api/workspaces'
	Jobs              *mux.Router // 'api/jobs'
	OIDC              *mux.Router // 'api/oidc'
	Guest             *mux.Router // 'api/guest'
	Groups            *mux.Router // 'api/groups'
	LinkPreview       *mux.Router // 'api/link-preview'
	CubeInternal      *mux.Router // 'api/internal/cube'
}

func InitAPI(a *app.App) {
	api := &API{
		app:        a,
		BaseRoutes: &Routes{},
	}

	api.BaseRoutes.Root = api.app.Server.Router
	api.BaseRoutes.Root.Use(api.SecurityHeadersMiddleware)
	api.BaseRoutes.APIRoot = api.app.Server.Router.PathPrefix("/api").Subrouter()
	api.BaseRoutes.APIRoot.Use(api.LocaleMiddleware)
	api.BaseRoutes.APIRoot.Use(api.VersionHeaderMiddleware)

	api.BaseRoutes.Files = api.BaseRoutes.APIRoot.PathPrefix("/files").Subrouter()
	api.BaseRoutes.Users = api.BaseRoutes.APIRoot.PathPrefix("/users").Subrouter()
	api.BaseRoutes.Office = api.BaseRoutes.APIRoot.PathPrefix("/office").Subrouter()
	api.BaseRoutes.Collimato = api.BaseRoutes.APIRoot.PathPrefix("/collimato").Subrouter()
	api.BaseRoutes.Config = api.BaseRoutes.APIRoot.PathPrefix("/config").Subrouter()
	api.BaseRoutes.Channels = api.BaseRoutes.APIRoot.PathPrefix("/channels").Subrouter()
	api.BaseRoutes.Activity = api.BaseRoutes.APIRoot.PathPrefix("/activity").Subrouter()
	api.BaseRoutes.Notifications = api.BaseRoutes.APIRoot.PathPrefix("/notifications").Subrouter()
	api.BaseRoutes.Settings = api.BaseRoutes.APIRoot.PathPrefix("/settings").Subrouter()
	api.BaseRoutes.Search = api.BaseRoutes.APIRoot.PathPrefix("/search").Subrouter()
	api.BaseRoutes.PublicLink = api.BaseRoutes.APIRoot.PathPrefix("/link").Subrouter()
	api.BaseRoutes.Auth = api.BaseRoutes.APIRoot.PathPrefix("/auth").Subrouter()
	api.BaseRoutes.PublicSharedFiles = api.BaseRoutes.APIRoot.PathPrefix("/public").Subrouter()
	api.BaseRoutes.License = api.BaseRoutes.APIRoot.PathPrefix("/license").Subrouter()
	api.BaseRoutes.Workspaces = api.BaseRoutes.APIRoot.PathPrefix("/workspaces").Subrouter()
	api.BaseRoutes.Jobs = api.BaseRoutes.APIRoot.PathPrefix("/jobs").Subrouter()
	api.BaseRoutes.Guest = api.BaseRoutes.APIRoot.PathPrefix("/guest").Subrouter()
	api.BaseRoutes.Groups = api.BaseRoutes.APIRoot.PathPrefix("/groups").Subrouter()
	api.BaseRoutes.LinkPreview = api.BaseRoutes.APIRoot.PathPrefix("/link-preview").Subrouter()
	api.BaseRoutes.CubeInternal = api.BaseRoutes.APIRoot.PathPrefix("/internal/cube").Subrouter()

	api.initActivity()
	api.initAuth()
	api.initFiles()
	api.initNotifications()
	api.initOffice()
	api.initPublic()
	api.initPublicShare()
	api.initSearch()
	api.initSettings()
	api.initUsers()
	api.initLicense()
	api.initCollimato()
	api.initCollimatoAdmin()
	api.initCubeInternal()
	api.initChannels()
	api.initGuest()
	api.initConfig()
	api.initProjectAdmin()
	api.initWorkspaces()
	api.initJobs()
	api.initGroups()
	api.initLinkPreview()

	if !config.IsDev {
		tlog.Info("Running in production mode")
		distDir := "frontend/dist"
		distFS := http.Dir(distDir)
		fileServer := http.FileServer(distFS)
		a.Server.Router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			f, err := distFS.Open(r.URL.Path)
			if err == nil {
				stat, err := f.Stat()
				f.Close()
				if err == nil && !stat.IsDir() {
					fileServer.ServeHTTP(w, r)
					return
				}
			}

			w.Header().Set("cache-control", "no-cache")
			http.ServeFile(w, r, distDir+"/index.html")
		})
	} else {
		tlog.Info("Running in development mode")
		// Proxy to Vite development server
		a.Server.Router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("cache-control", "no-cache")

			http.Redirect(w, r, "http://localhost:5173"+r.RequestURI, http.StatusTemporaryRedirect)
		})

	}
}
