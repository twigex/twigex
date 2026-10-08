// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/twigex/twigex/internal/parse"
	"github.com/twigex/twigex/model"
)

func (a *API) initUsers() {
	a.BaseRoutes.Users.HandleFunc("", a.getUsers).Methods("GET")
	a.BaseRoutes.Users.HandleFunc("/by-ids", a.getUsersByIDs).Methods("GET")
	a.BaseRoutes.Users.HandleFunc("/by-usernames", a.getUsersByUsernames).Methods("GET")
	a.BaseRoutes.Users.HandleFunc("/search", a.searchUsers).Methods("GET")
	a.BaseRoutes.Users.HandleFunc("/photo/{id}", a.getUserPhoto).Methods("GET")
	a.BaseRoutes.Users.HandleFunc("/create", a.createUser).Methods("POST")
	a.BaseRoutes.Users.HandleFunc("/update", a.updateUser).Methods("POST")
	a.BaseRoutes.Users.HandleFunc("/{id}/deactivate", a.deactivateUser).Methods("POST")
	a.BaseRoutes.Users.HandleFunc("/profile/update", a.updateProfile).Methods("POST")
	a.BaseRoutes.Users.HandleFunc("/{id}/password/reset", a.resetUserPassword).Methods("POST")
	a.BaseRoutes.Users.HandleFunc("/{id}/sessions/revoke", a.revokeUserSessions).Methods("POST")
	a.BaseRoutes.Users.HandleFunc("/sessions/{id}/logout", a.logOutSession).Methods("POST")
	a.BaseRoutes.Users.HandleFunc("/me", a.getMe).Methods("GET")
	a.BaseRoutes.Users.HandleFunc("/me/permissions", a.getMyPermissions).Methods("GET")
	a.BaseRoutes.Users.HandleFunc("/me/preferences", a.getMyPreferences).Methods("GET")
	a.BaseRoutes.Users.HandleFunc("/me/preferences", a.updateMyPreferences).Methods("POST")
	a.BaseRoutes.Users.HandleFunc("/me/password/update", a.updatePassword).Methods("POST")
	a.BaseRoutes.Users.HandleFunc("/me/meetings", a.getMyMeetings).Methods("GET")
	a.BaseRoutes.Users.HandleFunc("/me/meetings/{id}", a.getMyMeeting).Methods("GET")
	a.BaseRoutes.Users.HandleFunc("/me/sessions", a.getMySessions).Methods("GET")
	a.BaseRoutes.Users.HandleFunc("/me/photo", a.uploadPhoto).Methods("POST")
	a.BaseRoutes.Users.HandleFunc("/me/photo", a.deletePhoto).Methods("DELETE")
	a.BaseRoutes.Users.HandleFunc("/mfa/generate", a.generateMfa).Methods("POST")
	a.BaseRoutes.Users.HandleFunc("/mfa", a.enableMfa).Methods("POST")
	a.BaseRoutes.Users.HandleFunc("/status", a.getStatus).Methods("GET")

	a.BaseRoutes.Users.Use(a.RequireSession)
	a.BaseRoutes.Users.Use(a.RequireCSRF)
}

func (a *API) getUsers(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	limit := parse.Int(r.URL.Query().Get("limit"), 50)
	offset := parse.Int(r.URL.Query().Get("offset"), 0)
	includeDeactivated := parse.Bool(r.URL.Query().Get("include_deactivated"), false)
	query := r.URL.Query().Get("q")

	users, total, appErr := a.app.GetUsersPaged(r.Context(), *user, query, sortFromQuery(r), limit, offset, includeDeactivated)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, model.UsersPage{Items: users, Total: total})
}

func (a *API) getUsersByIDs(w http.ResponseWriter, r *http.Request) {
	raw := r.URL.Query().Get("ids")
	if raw == "" {
		respondJSON(w, http.StatusOK, []model.User{})
		return
	}

	ids := strings.Split(raw, ",")
	if len(ids) > 100 {
		ids = ids[:100]
	}

	users, appErr := a.app.GetUsersByIDs(ids)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, users)
}

func (a *API) getUsersByUsernames(w http.ResponseWriter, r *http.Request) {
	raw := r.URL.Query().Get("usernames")
	if raw == "" {
		respondJSON(w, http.StatusOK, []model.User{})
		return
	}

	usernames := strings.Split(raw, ",")
	if len(usernames) > 100 {
		usernames = usernames[:100]
	}

	users, appErr := a.app.GetUsersByUsernames(r.Context(), usernames)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, users)
}

func (a *API) searchUsers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	users, appErr := a.app.SearchUsers(r.Context(), query, limit)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, users)
}

func (a *API) getUserPhoto(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	if appErr := a.app.ServeUserPhoto(params["id"], w, r); appErr != nil {
		respondAppError(w, r, appErr)
	}
}

func (a *API) createUser(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	req := model.NewUser{}
	if !decodeBody(w, r, &req) {
		return
	}

	newUser, appErr := a.app.CreateUser(*user, req)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, newUser)
}

func (a *API) updateUser(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	req := model.UserPatch{}
	if !decodeBody(w, r, &req) {
		return
	}

	editUser, appErr := a.app.EditUser(*user, req)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, editUser)
}

func (a *API) deactivateUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	deactivatedUser, appErr := a.app.DeactivateUser(*user, params["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.DeleteUserSessions(*deactivatedUser)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, deactivatedUser)
}

func (a *API) updateProfile(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	u := model.UserProfileRequest{}
	if !decodeBody(w, r, &u) {
		return
	}

	appErr = a.app.UpdateProfile(*user, u)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) updatePassword(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	req := struct {
		Password    string
		OldPassword string
		Confirmed   string
	}{}
	if !decodeBody(w, r, &req) {
		return
	}

	currentSession := ""
	if sid, err := r.Cookie(model.SessionCookieToken); err == nil {
		currentSession = sid.Value
	}

	appErr = a.app.UpdatePassword(req.Password, req.Confirmed, req.OldPassword, *user, currentSession)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) resetUserPassword(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.ResetUserPassword(*user, params["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) revokeUserSessions(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.RevokeUserSessions(*user, params["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) uploadPhoto(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			respondAppError(w, r, model.NewAppError("user.photo_upload_failed", http.StatusRequestEntityTooLarge))
			return
		}

		respondAppError(w, r, model.NewAppError("user.photo_upload_failed", http.StatusBadRequest))
		return
	}

	file, _, err := r.FormFile("photo")
	if err != nil {
		respondAppError(w, r, model.NewAppError("user.photo_upload_failed", http.StatusBadRequest))
		return
	}

	defer file.Close()

	newPhoto, appErr := a.app.UploadPhoto(r.Context(), *user, file)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, newPhoto)
}

func (a *API) deletePhoto(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr = a.app.DeletePhoto(r.Context(), *user); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) getMe(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.Header().Set("Cache-Control", "private, no-cache")
	respondJSON(w, http.StatusOK, user)
}

func (a *API) getMyMeetings(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	meetings, appErr := a.app.GetUserMeetings(r.Context(), *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, meetings)
}

func (a *API) getMyMeeting(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	meeting, appErr := a.app.GetMeeting(r.Context(), *user, mux.Vars(r)["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, meeting)
}

func (a *API) getMyPermissions(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	permissions, appErr := a.app.GetMyPermissions(*user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, permissions)
}

func (a *API) getMyPreferences(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	preferences, appErr := a.app.GetMyPreferences(*user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, preferences)
}

func (a *API) updateMyPreferences(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	req := struct {
		Preferences []model.Preference
	}{}
	if !decodeBody(w, r, &req) {
		return
	}

	appErr = a.app.UpdateUserPreferences(*user, req.Preferences)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) generateMfa(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	mfa, appErr := a.app.GenerateMfa(*user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, mfa)
}

func (a *API) enableMfa(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	s := struct {
		Activate bool
		Token    string
	}{}
	if !decodeBody(w, r, &s) {
		return
	}

	if s.Activate {
		appErr = a.app.EnableMfa(*user, s.Token)
	} else {
		appErr = a.app.DisableMfa(*user)
	}

	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) getStatus(w http.ResponseWriter, r *http.Request) {
	status, appErr := a.app.GetStatus()
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, status)
}
