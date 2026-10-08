// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

func (a *API) initChannels() {
	a.BaseRoutes.Channels.HandleFunc("", a.getChannelAndDMList).Methods("GET")
	a.BaseRoutes.Channels.HandleFunc("/all", a.getChannelList).Methods("GET")
	a.BaseRoutes.Channels.HandleFunc("", a.createChannel).Methods("POST")
	a.BaseRoutes.Channels.HandleFunc("/klipy/search", a.searchKlipyGifs).Methods("GET")
	a.BaseRoutes.Channels.HandleFunc("/klipy/categories", a.getKlipyCategories).Methods("GET")
	a.BaseRoutes.Channels.HandleFunc("/klipy/trending", a.getKlipyTrending).Methods("GET")

	a.BaseRoutes.Channels.HandleFunc("/{id}", a.getChannelByID).Methods("GET")
	a.BaseRoutes.Channels.HandleFunc("/{id}/patch", a.updateChannel).Methods("POST")
	a.BaseRoutes.Channels.HandleFunc("/{id}/join", a.joinChannel).Methods("GET")
	a.BaseRoutes.Channels.HandleFunc("/{id}/archive", a.archiveChannel).Methods("POST")
	a.BaseRoutes.Channels.HandleFunc("/{id}/unarchive", a.unarchiveChannel).Methods("POST")
	a.BaseRoutes.Channels.HandleFunc("/{id}", a.deleteChannel).Methods("DELETE")
	a.BaseRoutes.Channels.HandleFunc("/{id}/upload", a.uploadChatFile).Methods("POST")
	a.BaseRoutes.Channels.HandleFunc("/{id}/users", a.addUserToChannel).Methods("POST")
	a.BaseRoutes.Channels.HandleFunc("/{id}/users/leave", a.leaveFromChannel).Methods("DELETE")
	a.BaseRoutes.Channels.HandleFunc("/{id}/users/{user}", a.removeUserFromChannel).Methods("DELETE")
	a.BaseRoutes.Channels.HandleFunc("/{id}/users/{user}/roles", a.changeUserRole).Methods("POST")
	a.BaseRoutes.Channels.HandleFunc("/{id}/posts", a.getChannelPosts).Methods("GET")
	a.BaseRoutes.Channels.HandleFunc("/{id}/posts", a.createPost).Methods("POST")
	a.BaseRoutes.Channels.HandleFunc("/{id}/posts/read", a.readChannelPosts).Methods("POST")
	a.BaseRoutes.Channels.HandleFunc("/{id}/posts/{postID}", a.deletePost).Methods("DELETE")
	a.BaseRoutes.Channels.HandleFunc("/{id}/posts/{postID}", a.updateChannelPost).Methods("POST")
	a.BaseRoutes.Channels.HandleFunc("/{id}/posts/{post}/reactions", a.createPostReaction).Methods("POST")
	a.BaseRoutes.Channels.HandleFunc("/{id}/posts/{post}/reactions/{reaction}", a.deletePostReaction).Methods("DELETE")
	a.BaseRoutes.Channels.HandleFunc("/{id}/posts/{post}/file/{file}", a.downloadChatFile).Methods("GET")
	a.BaseRoutes.Channels.HandleFunc("/{id}/posts/{post}/forward", a.forwardPost).Methods("POST")
	a.BaseRoutes.Channels.HandleFunc("/{id}/members", a.getChannelMemberUsers).Methods("GET")
	a.BaseRoutes.Channels.HandleFunc("/{id}/members/search", a.searchChannelMembers).Methods("GET")
	a.BaseRoutes.Channels.HandleFunc("/{id}/meetings", a.getChannelMeetings).Methods("GET")
	a.BaseRoutes.Channels.HandleFunc("/{id}/meetings", a.createChannelMeeting).Methods("POST")
	a.BaseRoutes.Channels.HandleFunc("/{id}/meetings/scheduled", a.getScheduledChannelMeetings).Methods("GET")
	a.BaseRoutes.Channels.HandleFunc("/{id}/meetings/{mid}/start", a.startChannelMeeting).Methods("POST")
	a.BaseRoutes.Channels.HandleFunc("/{id}/meetings/{mid}/end", a.endChannelMeeting).Methods("POST")
	a.BaseRoutes.Channels.HandleFunc("/{id}/meetings/{mid}/cancel", a.cancelChannelMeeting).Methods("POST")
	a.BaseRoutes.Channels.HandleFunc("/{id}/meetings/{mid}", a.updateChannelMeeting).Methods("POST")
	a.BaseRoutes.Channels.HandleFunc("/{id}/call", a.startDirectCall).Methods("POST")
	a.BaseRoutes.Channels.HandleFunc("/{id}/call/answer", a.answerDirectCall).Methods("POST")
	a.BaseRoutes.Channels.HandleFunc("/{id}/call/decline", a.declineDirectCall).Methods("POST")
	a.BaseRoutes.Channels.HandleFunc("/{id}/call/cancel", a.cancelDirectCall).Methods("POST")
	a.BaseRoutes.Channels.HandleFunc("/{id}/call/end", a.endDirectCall).Methods("POST")
	a.BaseRoutes.Channels.HandleFunc("/{id}/meetings/{mid}/host", a.setChannelMeetingHost).Methods("POST")
	a.BaseRoutes.Channels.HandleFunc("/{id}/meetings/{mid}/presence", a.announceMeetingPresence).Methods("POST")
	a.BaseRoutes.Channels.HandleFunc("/{id}/meetings/{mid}/video/token", a.getMeetingVideoToken).Methods("GET")
	a.BaseRoutes.Channels.HandleFunc("/{id}/meetings/{mid}/invites", a.getMeetingGuestLinks).Methods("GET")
	a.BaseRoutes.Channels.HandleFunc("/{id}/meetings/{mid}/invite", a.inviteMeetingGuests).Methods("POST")
	a.BaseRoutes.Channels.HandleFunc("/{id}/meetings/{mid}/members", a.inviteMeetingMembers).Methods("POST")
	a.BaseRoutes.Channels.HandleFunc("/{id}/meetings/{mid}/invite/{link}", a.revokeVideoGuestLink).Methods("DELETE")
	a.BaseRoutes.Channels.HandleFunc("/{id}/meetings/{mid}/groups", a.listMeetingGroups).Methods("GET")
	a.BaseRoutes.Channels.HandleFunc("/{id}/meetings/{mid}/groups", a.addMeetingGroups).Methods("POST")
	a.BaseRoutes.Channels.HandleFunc("/{id}/groups", a.listChannelGroups).Methods("GET")
	a.BaseRoutes.Channels.HandleFunc("/{id}/groups", a.addChannelGroups).Methods("POST")
	a.BaseRoutes.Channels.HandleFunc("/{id}/groups/{group}", a.removeChannelGroup).Methods("DELETE")

	a.BaseRoutes.Channels.Use(a.RequireSession)
	a.BaseRoutes.Channels.Use(a.RequireCSRF)
}

func (a *API) getChannelAndDMList(w http.ResponseWriter, r *http.Request) {
	// Return DMS and Channels for user
	urlParams := r.URL.Query().Get("all")

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	channels, appErr := a.app.GetChannelAndDMList(*user, urlParams)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, channels)
}

func (a *API) getChannelList(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	channels, appErr := a.app.GetChannelList(*user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, channels)
}

func (a *API) uploadChatFile(w http.ResponseWriter, r *http.Request) {
	channelID := mux.Vars(r)["id"]

	r.Body = http.MaxBytesReader(w, r.Body, 100<<20)
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		if err.Error() == "http: request body too large" {
			respondAppError(w, r, model.NewAppError("channel.file_too_large", http.StatusRequestEntityTooLarge))
			return
		}

		respondAppError(w, r, model.NewAppError("request.invalid", http.StatusBadRequest))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondAppError(w, r, model.NewAppError("upload.missing_name", http.StatusBadRequest))
		return
	}

	defer file.Close()

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if !a.requirePermission(w, r, *user, model.ChatPermissions.PermissionUploadChatFiles) {
		return
	}

	fileID, storageID, appErr := a.app.UploadChatFile(channelID, *user, file, header.Size)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"id": fileID, "storage_id": storageID})
}

func (a *API) createChannel(w http.ResponseWriter, r *http.Request) {
	s := struct {
		Name        string
		Type        string
		Description string
		User        string // for direct message
	}{}

	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if !a.requirePermission(w, r, *user, model.ChatPermissions.PermissionCreateChannel) {
		return
	}

	ch, appErr := a.app.CreateChannel(*user, s.Name, s.Type, s.Description, s.User)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, ch)
}

func (a *API) addUserToChannel(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	s := struct {
		Users []string
	}{}

	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	newMember, appErr := a.app.AddUsersToChannelByID(*user, params["id"], s.Users)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, newMember)
}

func (a *API) removeUserFromChannel(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.RemoveUserFromChannel(*user, params["id"], params["user"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) leaveFromChannel(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.LeaveFromChannel(*user, params["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) changeUserRole(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	s := struct {
		Role string
	}{}

	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.ChangeUserRole(*user, params["id"], params["user"], s.Role)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) getChannelPosts(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	param := r.URL.Query().Get("param")
	postID := r.URL.Query().Get("id")

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	posts, appErr := a.app.GetChannelPosts(r.Context(), *user, params["id"], param, postID)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, posts)
}

func (a *API) createPost(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	newPost := model.NewPost{}
	if !decodeBody(w, r, &newPost) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if !a.requirePermission(w, r, *user, model.ChatPermissions.PermissionSendMessage) {
		return
	}

	post, appErr := a.app.CreatePost(r.Context(), *user, params["id"], newPost.Message, newPost.Reply, newPost.Attachments, newPost.GIF, newPost.ValidPendingPostID())
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, post)
}

func (a *API) updateChannelPost(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	s := struct {
		Message     string
		Remove      []string
		Attachments []model.PostFileMetadata
		GIF         string
	}{}

	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	post, appErr := a.app.UpdatePost(r.Context(), *user, params["id"], params["postID"], s.Message, s.Remove, s.Attachments, s.GIF)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, post)
}

func (a *API) readChannelPosts(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.ReadChannelPosts(*user, params["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) deletePost(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.DeletePost(r.Context(), *user, params["id"], params["postID"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) createPostReaction(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	s := struct {
		Reaction string
	}{}

	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	reaction, appErr := a.app.CreatePostReaction(r.Context(), *user, params["id"], params["post"], s.Reaction)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, reaction)
}

func (a *API) deletePostReaction(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.DeletePostReaction(r.Context(), *user, params["id"], params["post"], params["reaction"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) downloadChatFile(w http.ResponseWriter, r *http.Request) {
	channelID := mux.Vars(r)["id"]
	fileID := mux.Vars(r)["file"]

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	attachment, appErr := a.app.GetChatFileAttachment(channelID, fileID, *user)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	setUploadHeaders(w, attachment.MimeType, attachment.Name)

	if attachment.StorageID != "" {
		filePath, err := a.app.BuildFilePath(attachment.StorageID, channelID+"/"+fileID, model.AppChat)
		if err != nil {
			tlog.Errorw("Failed to build chat file path",
				"file_id", fileID,
				"error", err,
			)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := a.app.FileStorageObjects[attachment.StorageID].ServeFile(filePath, w, r); err != nil {
			tlog.Errorw("Failed to serve chat file",
				"file_id", fileID,
				"error", err,
			)
		}

		return
	}

	http.ServeFile(w, r, path.Join("data/chat", channelID, fileID))
}

func (a *API) forwardPost(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	s := struct {
		Channels []string
		Users    []string
	}{}

	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.ForwardChannelPost(r.Context(), *user, params["id"], params["post"], s.Channels, s.Users)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) getChannelByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	channel, appErr := a.app.GetChannelByID(*user, params["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, channel)
}

func (a *API) updateChannel(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	s := model.ChannelPatch{}
	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	channel, appErr := a.app.UpdateChannel(*user, params["id"], s)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, channel)
}

func (a *API) joinChannel(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	channel, appErr := a.app.JoinChannel(*user, params["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, channel)
}

func (a *API) archiveChannel(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr := a.app.ArchiveChannel(r.Context(), *user, mux.Vars(r)["id"]); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, map[string]bool{"archived": true})
}

func (a *API) unarchiveChannel(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr := a.app.UnarchiveChannel(r.Context(), *user, mux.Vars(r)["id"]); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, map[string]bool{"archived": false})
}

func (a *API) deleteChannel(w http.ResponseWriter, r *http.Request) {
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr := a.app.DeleteChannel(r.Context(), *user, mux.Vars(r)["id"]); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusAccepted, map[string]bool{"deleting": true})
}

func (a *API) getChannelMemberUsers(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	members, appErr := a.app.GetChannelMemberUsers(*user, params["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, members)
}

func (a *API) searchChannelMembers(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	query := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	users, appErr := a.app.SearchChannelMemberUsers(r.Context(), *user, params["id"], query, limit)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, users)
}

func (a *API) createChannelMeeting(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	s := struct {
		ScheduledAt     int64    `json:"scheduled_at"`
		Title           string   `json:"title"`
		DurationMinutes int      `json:"duration_minutes"`
		Invitees        []string `json:"invitees"`
		GuestPassword   string   `json:"guest_password"`
		Timezone        string   `json:"timezone"`
		Groups          []string `json:"groups"`
	}{}
	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	meeting, appErr := a.app.CreateChannelMeeting(r.Context(), *user, params["id"], s.ScheduledAt, s.Title, s.DurationMinutes, s.Invitees, s.GuestPassword, s.Timezone, s.Groups)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, meeting)
}

func (a *API) getChannelMeetings(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	meetings, appErr := a.app.GetActiveChannelMeetings(r.Context(), *user, params["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, meetings)
}

func (a *API) getScheduledChannelMeetings(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	meetings, appErr := a.app.GetScheduledChannelMeetings(r.Context(), *user, params["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, meetings)
}

func (a *API) startChannelMeeting(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	meeting, appErr := a.app.StartChannelMeeting(r.Context(), *user, params["id"], params["mid"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, meeting)
}

func (a *API) endChannelMeeting(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.EndChannelMeeting(r.Context(), *user, params["id"], params["mid"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) cancelChannelMeeting(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.CancelChannelMeeting(r.Context(), *user, params["id"], params["mid"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) startDirectCall(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	result, appErr := a.app.StartDirectCall(r.Context(), *user, params["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *API) answerDirectCall(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	result, appErr := a.app.AnswerDirectCall(r.Context(), *user, params["id"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *API) declineDirectCall(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr := a.app.DeclineDirectCall(r.Context(), *user, params["id"]); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) cancelDirectCall(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr := a.app.CancelDirectCall(r.Context(), *user, params["id"]); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) endDirectCall(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr := a.app.EndDirectCall(r.Context(), *user, params["id"]); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) updateChannelMeeting(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	s := struct {
		ScheduledAt     int64    `json:"scheduled_at"`
		Title           string   `json:"title"`
		DurationMinutes int      `json:"duration_minutes"`
		Invitees        []string `json:"invitees"`
		GuestPassword   string   `json:"guest_password"`
		Timezone        string   `json:"timezone"`
		Groups          []string `json:"groups"`
	}{}
	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	meeting, appErr := a.app.UpdateChannelMeeting(r.Context(), *user, params["id"], params["mid"], s.ScheduledAt, s.Title, s.DurationMinutes, s.Invitees, s.GuestPassword, s.Timezone, s.Groups)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, meeting)
}

func (a *API) setChannelMeetingHost(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	s := struct {
		UserID string `json:"user_id"`
	}{}
	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	meeting, appErr := a.app.SetChannelMeetingHost(r.Context(), *user, params["id"], params["mid"], s.UserID)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, meeting)
}

func (a *API) announceMeetingPresence(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	if appErr := a.app.AnnounceMeetingPresence(r.Context(), *user, params["id"], params["mid"]); appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) getMeetingVideoToken(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	token, appErr := a.app.GetMeetingVideoToken(r.Context(), params["id"], params["mid"], user.ID)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, token)
}

func (a *API) inviteMeetingGuests(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	s := struct {
		Email          []string `json:"email"`
		ExpiresInHours int      `json:"expires_in_hours"`
	}{}
	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.InviteMeetingGuests(r.Context(), *user, params["id"], params["mid"], s.Email, s.ExpiresInHours)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) inviteMeetingMembers(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	s := struct {
		Invitees []string `json:"invitees"`
	}{}
	if !decodeBody(w, r, &s) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.InviteMeetingMembers(r.Context(), *user, params["id"], params["mid"], s.Invitees)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) getMeetingGuestLinks(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	links, appErr := a.app.GetActiveMeetingGuestLinks(r.Context(), *user, params["id"], params["mid"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, links)
}

func (a *API) revokeVideoGuestLink(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	appErr = a.app.RevokeVideoGuestLink(r.Context(), *user, params["id"], params["mid"], params["link"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) klipyRequest(w http.ResponseWriter, r *http.Request, endpoint string, params url.Values) bool {
	params.Set("key", *a.app.ConfigStore.Config.ChannelSettings.KlipyGIF.APIKey)

	resp, err := http.Get("https://api.klipy.com/v2/" + endpoint + "?" + params.Encode())
	if err != nil {
		tlog.Errorw("Failed to fetch from Klipy",
			"endpoint", endpoint,
			"error", err,
		)
		respondAppError(w, r, model.NewAppError("channel.gif_failed", http.StatusInternalServerError))
		return false
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		tlog.Errorw("Failed to read Klipy response",
			"endpoint", endpoint,
			"error", err,
		)
		respondAppError(w, r, model.NewAppError("channel.gif_failed", http.StatusInternalServerError))
		return false
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(body); err != nil {
		tlog.Errorw("Failed to write response", "error", err)
	}

	return true
}

func (a *API) getKlipyCategories(w http.ResponseWriter, r *http.Request) {
	if !*a.app.ConfigStore.Config.ChannelSettings.KlipyGIF.Enabled {
		respondAppError(w, r, model.NewAppError("channel.gif_disabled", http.StatusForbidden))
		return
	}

	a.klipyRequest(w, r, "categories", url.Values{})
}

func (a *API) getKlipyTrending(w http.ResponseWriter, r *http.Request) {
	if !*a.app.ConfigStore.Config.ChannelSettings.KlipyGIF.Enabled {
		respondAppError(w, r, model.NewAppError("channel.gif_disabled", http.StatusForbidden))
		return
	}

	query := r.URL.Query()
	params := url.Values{}
	params.Set("limit", "20")
	if limit := query.Get("limit"); limit != "" {
		params.Set("limit", limit)
	}

	if next := query.Get("next"); next != "" {
		params.Set("pos", next)
	}

	a.klipyRequest(w, r, "featured", params)
}

func (a *API) searchKlipyGifs(w http.ResponseWriter, r *http.Request) {
	if !*a.app.ConfigStore.Config.ChannelSettings.KlipyGIF.Enabled {
		respondAppError(w, r, model.NewAppError("channel.gif_disabled", http.StatusForbidden))
		return
	}

	query := r.URL.Query()
	params := url.Values{}
	params.Set("q", query.Get("search"))
	params.Set("limit", "20")
	if limit := query.Get("limit"); limit != "" {
		params.Set("limit", limit)
	}

	if next := query.Get("next"); next != "" {
		params.Set("pos", next)
	}

	a.klipyRequest(w, r, "search", params)
}

func (a *API) listMeetingGroups(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	groups, appErr := a.app.ListMeetingGroups(r.Context(), *user, params["id"], params["mid"])
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, groups)
}

func (a *API) addMeetingGroups(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	var req model.AddMeetingGroupsRequest
	if !decodeBody(w, r, &req) {
		return
	}

	user, appErr := a.app.GetCurrentUser(r)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	added, appErr := a.app.AddGroupsToMeeting(r.Context(), *user, params["id"], params["mid"], req)
	if appErr != nil {
		respondAppError(w, r, appErr)
		return
	}

	respondJSON(w, http.StatusOK, added)
}
