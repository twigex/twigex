// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"path"
	"regexp"
	"strings"

	"github.com/twigex/twigex/crypto"
	"github.com/twigex/twigex/internal/mfa"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

var (
	reUpperLower = regexp.MustCompile("[A-Z][a-z]")
	reNumeric    = regexp.MustCompile("[0-9]")
	reSpecial    = regexp.MustCompile("[!@#$%^&*)(+=._-]")
)

func (a *App) GetCurrentUser(r *http.Request) (*model.User, *model.AppError) {
	id, err := a.Store.Auth.GetUserID(r)
	if err != nil {
		return nil, model.NewAppError("session.invalid", http.StatusUnauthorized)
	}

	user, err := a.Store.User.Get(*id)
	if err != nil {
		tlog.Errorw("Failed to retrieve current user",
			"user_id", *id,
			"error", err,
		)
		return nil, model.NewAppError("user.not_found", http.StatusInternalServerError)
	}

	if user == nil {
		return nil, model.NewAppError("user.not_found", http.StatusNotFound)
	}

	// Fold in any system roles granted through group membership so every
	// downstream permission check sees the user's full effective role set.
	a.applyGroupRoles(r.Context(), user)

	return user, nil
}

const (
	userSearchDefaultLimit = 20
	userSearchMaxLimit     = 50
)

// SearchUsers backs the user-picker typeahead: it returns a small, capped set
// of users matching the query so callers never load the full directory. Any
// authenticated user may search (same exposure as GetUsers, which the pickers
// previously used wholesale).
func (a *App) SearchUsers(ctx context.Context, query string, limit int) ([]model.User, *model.AppError) {
	if limit <= 0 {
		limit = userSearchDefaultLimit
	}

	if limit > userSearchMaxLimit {
		limit = userSearchMaxLimit
	}

	users, err := a.Store.User.Search(ctx, strings.TrimSpace(query), limit)
	if err != nil {
		tlog.Errorw("Failed to search users", "error", err)
		return nil, model.NewAppError("user.search_failed", http.StatusInternalServerError)
	}

	return users, nil
}

func (a *App) GetUsersPaged(ctx context.Context, user model.User, query string, sort model.Sort, limit, offset int, includeDeactivated bool) ([]model.User, int, *model.AppError) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	if offset < 0 {
		offset = 0
	}

	if includeDeactivated && !a.canManageUsers(user) {
		return nil, 0, model.NewAppError("permission.forbidden", http.StatusForbidden)
	}

	query = strings.TrimSpace(query)

	users, err := a.Store.User.GetAllPaged(ctx, query, sort, limit, offset, includeDeactivated)
	if err != nil {
		tlog.Errorw("Failed to retrieve users paged", "error", err)
		return nil, 0, model.NewAppError("user.retrieval_failed", http.StatusInternalServerError)
	}

	total, err := a.Store.User.Count(ctx, query, includeDeactivated)
	if err != nil {
		tlog.Errorw("Failed to count users", "error", err)
		return nil, 0, model.NewAppError("user.retrieval_failed", http.StatusInternalServerError)
	}

	return users, total, nil
}

func (a *App) canManageUsers(user model.User) bool {
	return a.SessionHasPermission(user, model.AdminPermissions.PermissionCreateUsers) ||
		a.SessionHasPermission(user, model.AdminPermissions.PermissionEditUsers) ||
		a.SessionHasPermission(user, model.AdminPermissions.PermissionDeleteUsers)
}

func (a *App) GetUsersByIDs(ids []string) ([]model.User, *model.AppError) {
	users, err := a.Store.User.GetByIDs(ids)
	if err != nil {
		tlog.Errorw("Failed to retrieve users by ids", "error", err)
		return nil, model.NewAppError("user.retrieval_failed", http.StatusInternalServerError)
	}
	// Strip sensitive fields before returning
	for i := range users {
		users[i].Password = ""
		users[i].MfaSecret = ""
		users[i].AuthData = ""
	}

	return users, nil
}

func (a *App) GetUsersByUsernames(ctx context.Context, usernames []string) ([]model.User, *model.AppError) {
	ctx, cancel := a.dbCtx(ctx)
	defer cancel()

	users, err := a.Store.User.GetByUsernames(ctx, usernames)
	if err != nil {
		tlog.Errorw("Failed to retrieve users by usernames", "error", err)
		return nil, model.NewAppError("user.retrieval_failed", http.StatusInternalServerError)
	}

	return users, nil
}

func (a *App) ServeUserPhoto(userID string, w http.ResponseWriter, r *http.Request) *model.AppError {
	user, err := a.Store.User.Get(userID)
	if err != nil {
		tlog.Errorw("Failed to retrieve user for photo", "user_id", userID, "error", err)
		return model.NewAppError("user.retrieval_failed", http.StatusInternalServerError)
	}

	if user == nil {
		return model.NewAppError("user.not_found", http.StatusNotFound)
	}

	ctx, cancel := a.dbCtx(r.Context())
	defer cancel()

	photo, err := a.Store.UserPhoto.Get(ctx, userID)
	if err != nil {
		tlog.Errorw("Failed to retrieve user photo", "user_id", userID, "error", err)
		return model.NewAppError("user.photo_serve_failed", http.StatusInternalServerError)
	}

	if photo == nil {
		return model.NewAppError("user.photo_not_found", http.StatusNotFound)
	}

	storage, err := a.Store.Storage.GetByID(photo.StorageID)
	if err != nil {
		tlog.Errorw("Failed to retrieve photo storage", "user_id", userID, "storage_id", photo.StorageID, "error", err)
		return model.NewAppError("user.photo_serve_failed", http.StatusInternalServerError)
	}

	if storage == nil {
		tlog.Errorw("Photo storage not found", "user_id", userID, "storage_id", photo.StorageID)
		return model.NewAppError("user.photo_not_found", http.StatusNotFound)
	}

	backend, exists := a.FileStorageObjects[photo.StorageID]
	if !exists || backend == nil {
		tlog.Errorw("Photo storage backend not found", "user_id", userID, "storage_id", photo.StorageID)
		return model.NewAppError("user.photo_not_found", http.StatusNotFound)
	}

	photoPath := path.Join(storage.Directory, "photos", photo.PhotoID)

	setPhotoCacheControl(w)

	// Answered before the file is opened, so revalidating an unchanged photo
	// costs the storage a metadata lookup rather than a read.
	if etag, err := backend.FileETag(photoPath); err == nil && etag != "" {
		// RFC 7232 requires the quotes. Our own check below reads a tag either
		// way, but net/http and any cache ignore an unquoted one.
		etag = `"` + strings.Trim(etag, `"`) + `"`
		w.Header().Set("ETag", etag)

		if etagMatches(r.Header.Get("If-None-Match"), etag) {
			w.WriteHeader(http.StatusNotModified)
			return nil
		}
	}

	if err := backend.ServeFile(photoPath, w, r); err != nil {
		// The error response must not inherit the photo's caching, and
		// ServeContent's own cleanup does not reach this path.
		w.Header().Del("Cache-Control")
		w.Header().Del("ETag")
		tlog.Errorw("Failed to serve user photo", "user_id", userID, "error", err)
		return model.NewAppError("user.photo_serve_failed", http.StatusInternalServerError)
	}

	return nil
}

// etagMatches applies the weak comparison If-None-Match calls for (RFC 7232
// 3.2), which ignores a W/ prefix on either side.
func etagMatches(ifNoneMatch, etag string) bool {
	if ifNoneMatch == "" {
		return false
	}

	if strings.TrimSpace(ifNoneMatch) == "*" {
		return true
	}

	want := strings.TrimPrefix(etag, "W/")
	for _, candidate := range strings.Split(ifNoneMatch, ",") {
		if strings.TrimPrefix(strings.TrimSpace(candidate), "W/") == want {
			return true
		}
	}

	return false
}

// setPhotoCacheControl keeps a member's face out of shared caches, and expires
// soon because nothing tells other viewers that a photo has changed.
func setPhotoCacheControl(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "private, max-age=300")
}

func (a *App) GetMyPreferences(user model.User) (*model.Preferences, *model.AppError) {
	display, err := a.Store.Preferences.GetForUser(user.ID, model.PreferencesCategoryDisplay)
	if err != nil {
		tlog.Errorw("Failed to retrieve display preferences",
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("user.preferences_failed", http.StatusInternalServerError)
	}

	notifications, err := a.Store.Preferences.GetForUser(user.ID, model.PreferencesCategoryNotifications)
	if err != nil {
		tlog.Errorw("Failed to retrieve notification preferences",
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("user.preferences_failed", http.StatusInternalServerError)
	}

	s := model.Preferences{
		DisplaySettings: display,
	}

	if len(notifications) == 0 {
		p := make([]model.Preference, 0, len(model.DefaultNotifications))
		for k, v := range model.DefaultNotifications {
			p = append(p, model.Preference{
				UserID:   user.ID,
				Category: model.PreferencesCategoryNotifications,
				Name:     k,
				Value:    v,
			})
		}

		s.Notifications = p
	} else {
		missing := generateMissingNotifications(user.ID, notifications)
		s.Notifications = append(notifications, missing...)
	}

	return &s, nil
}

func (a *App) UpdateUserPreferences(user model.User, notifications []model.Preference) *model.AppError {
	for _, v := range notifications {
		if v.UserID != user.ID {
			return model.NewAppError("session.not_authorized", http.StatusForbidden)
		}
	}

	if err := a.Store.Preferences.Update(notifications); err != nil {
		tlog.Errorw("Failed to update user preferences",
			"user_id", user.ID,
			"error", err,
		)
		return model.NewAppError("user.preferences_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) isValidPassword(p string) (bool, *model.AppError) {
	settings := a.ConfigStore.Config.PasswordSettings

	if len(p) < *settings.PasswordMinLength {
		return false, model.NewAppError("user.password_too_short", http.StatusBadRequest)
	}

	if *settings.UpperLowerCharacters {
		if !reUpperLower.MatchString(p) {
			return false, model.NewAppError("user.password_missing_case", http.StatusBadRequest)
		}
	}

	if *settings.NumericCharacters {
		if !reNumeric.MatchString(p) {
			return false, model.NewAppError("user.password_missing_number", http.StatusBadRequest)
		}
	}

	if *settings.SpecialCharacters {
		if !reSpecial.MatchString(p) {
			return false, model.NewAppError("user.password_missing_special", http.StatusBadRequest)
		}
	}

	return true, nil
}

// SeatsExhausted reports whether another user would exceed the hard limit.
// Over-limit instances (a downgrade, say) keep every existing user working; only
// adding another is refused, and only when the licence blocks rather than
// settling overage at renewal.
func (a *App) SeatsExhausted(ctx context.Context) (bool, *model.AppError) {
	license := a.Server.License
	if !license.BlocksOnSeatLimit() {
		return false, nil
	}

	limit := license.SeatHardLimit()
	if limit <= 0 {
		return false, nil
	}

	count, appErr := a.ActiveUserCount(ctx)
	if appErr != nil {
		return false, appErr
	}

	return count >= limit, nil
}

// SeatsOverSoftLimit reports whether the instance is using more seats than it
// licensed. Within the grace band this is true while SeatsExhausted is still
// false, which is the window the admin needs to be told about.
func (a *App) SeatsOverSoftLimit(ctx context.Context) (bool, *model.AppError) {
	limit := a.Server.License.SeatLimit()
	if limit <= 0 {
		return false, nil
	}

	count, appErr := a.ActiveUserCount(ctx)
	if appErr != nil {
		return false, appErr
	}

	return count > limit, nil
}

func (a *App) ActiveUserCount(ctx context.Context) (int, *model.AppError) {
	count, err := a.Store.User.CountActive(ctx)
	if err != nil {
		tlog.Errorw("Failed to count active users",
			"error", err,
		)
		return 0, model.NewAppError("user.retrieval_failed", http.StatusInternalServerError)
	}

	return count, nil
}

// requireIdentityFree names the field that clashed, rather than letting the
// unique constraint surface as a server error. selfID keeps a user their own
// values. The constraint is still the authority: two admins submitting the
// same name at once can both pass here, and the second write fails as before.
func (a *App) requireIdentityFree(username, email, selfID string) *model.AppError {
	byUsername, err := a.Store.User.GetByUsername(username)
	if err != nil {
		tlog.Errorw("Failed to check username availability", "error", err)
		return model.NewAppError("user.retrieval_failed", http.StatusInternalServerError)
	}

	if byUsername != nil && byUsername.ID != selfID {
		return model.NewAppError("user.username_taken", http.StatusConflict)
	}

	byEmail, err := a.Store.User.GetByEmail(email)
	if err != nil {
		tlog.Errorw("Failed to check email availability", "error", err)
		return model.NewAppError("user.retrieval_failed", http.StatusInternalServerError)
	}

	if byEmail != nil && byEmail.ID != selfID {
		return model.NewAppError("user.email_taken", http.StatusConflict)
	}

	return nil
}

func (a *App) CreateUser(user model.User, req model.NewUser) (*model.User, *model.AppError) {
	if !a.SessionHasPermission(user, model.AdminPermissions.PermissionCreateUsers) {
		return nil, model.NewAppError("permission.forbidden", http.StatusForbidden)
	}

	if req.Email == "" {
		return nil, model.NewAppError("user.email_required", http.StatusBadRequest)
	}

	if req.Username == "" {
		return nil, model.NewAppError("user.username_required", http.StatusBadRequest)
	}

	req.Username = model.NormalizeUsername(req.Username)
	if !model.ValidUsername(req.Username) {
		return nil, model.NewAppError("user.username_invalid", http.StatusBadRequest)
	}

	if model.ReservedUsername(req.Username) {
		return nil, model.NewAppError("user.username_reserved", http.StatusBadRequest)
	}

	if req.StorageLimit < 0 {
		return nil, model.NewAppError("user.storage_limit_invalid", http.StatusBadRequest)
	}

	if appErr := guardNewUserRole(user, req.Role); appErr != nil {
		return nil, appErr
	}

	if req.Name == "" {
		return nil, model.NewAppError("user.name_required", http.StatusBadRequest)
	}

	if req.LastName == "" {
		return nil, model.NewAppError("user.lastname_required", http.StatusBadRequest)
	}

	exhausted, appErr := a.SeatsExhausted(context.Background())
	if appErr != nil {
		return nil, appErr
	}

	if exhausted {
		return nil, model.NewAppError("user.seat_limit", http.StatusPaymentRequired)
	}

	req.Email = strings.ToLower(req.Email)

	if appErr := a.requireIdentityFree(req.Username, req.Email, ""); appErr != nil {
		return nil, appErr
	}

	sendFinishRegistration := false
	password := req.Password
	if password == "" {
		sendFinishRegistration = true
		password = crypto.RandomText()
	}

	hashedPassword, err := crypto.HashPassword(password)
	if err != nil {
		tlog.Errorw("Failed to hash password during user creation",
			"error", err,
		)
		return nil, model.NewAppError("user.create_failed", http.StatusInternalServerError)
	}

	req.Password = hashedPassword
	newUser, err := a.Store.User.Create(req)
	if err != nil {
		tlog.Errorw("Failed to insert user",
			"error", err,
		)
		return nil, model.NewAppError("user.create_failed", http.StatusInternalServerError)
	}

	if sendFinishRegistration {
		token := crypto.RandomText()
		if err = a.Store.PasswordReset.Create(*newUser, crypto.HashSHA256(token)); err != nil {
			tlog.Errorw("Failed to create password reset token",
				"user_id", newUser.ID,
				"error", err,
			)
			return nil, model.NewAppError("user.create_failed", http.StatusInternalServerError)
		}

		go a.SendFinishRegistrationEmail(*newUser, token)
	}

	return newUser, nil
}

func (a *App) EditUser(user model.User, req model.UserPatch) (*model.User, *model.AppError) {
	if !a.SessionHasPermission(user, model.AdminPermissions.PermissionEditUsers) {
		return nil, model.NewAppError("permission.forbidden", http.StatusForbidden)
	}

	if req.Email == "" {
		return nil, model.NewAppError("user.email_required", http.StatusBadRequest)
	}

	if req.Role == "" {
		return nil, model.NewAppError("user.role_required", http.StatusBadRequest)
	}

	if req.Name == "" {
		return nil, model.NewAppError("user.name_required", http.StatusBadRequest)
	}

	if req.LastName == "" {
		return nil, model.NewAppError("user.lastname_required", http.StatusBadRequest)
	}

	if req.Username == "" {
		return nil, model.NewAppError("user.username_required", http.StatusBadRequest)
	}

	req.Username = model.NormalizeUsername(req.Username)
	if !model.ValidUsername(req.Username) {
		return nil, model.NewAppError("user.username_invalid", http.StatusBadRequest)
	}

	if model.ReservedUsername(req.Username) {
		return nil, model.NewAppError("user.username_reserved", http.StatusBadRequest)
	}

	if req.StorageLimit < 0 {
		return nil, model.NewAppError("user.storage_limit_invalid", http.StatusBadRequest)
	}

	existing, err := a.Store.User.Get(req.ID)
	if err != nil {
		tlog.Errorw("Failed to retrieve user before editing",
			"user_id", req.ID,
			"error", err,
		)
		return nil, model.NewAppError("user.retrieval_failed", http.StatusInternalServerError)
	}

	if existing == nil {
		return nil, model.NewAppError("user.not_found", http.StatusNotFound)
	}

	if appErr := guardUserRoleChange(user, existing, req.Role); appErr != nil {
		return nil, appErr
	}

	if existing.AuthService == "ldap" && existing.Username != req.Username {
		return nil, model.NewAppError("user.username_managed_externally", http.StatusBadRequest)
	}

	req.Email = strings.ToLower(req.Email)
	if appErr := a.requireIdentityFree(req.Username, req.Email, req.ID); appErr != nil {
		return nil, appErr
	}

	if err := a.Store.User.Update(req); err != nil {
		tlog.Errorw("Failed to edit user",
			"user_id", req.ID,
			"error", err,
		)
		return nil, model.NewAppError("user.update_failed", http.StatusInternalServerError)
	}

	editUser, err := a.Store.User.Get(req.ID)
	if err != nil {
		tlog.Errorw("Failed to retrieve updated user",
			"user_id", req.ID,
			"error", err,
		)
		return nil, model.NewAppError("user.retrieval_failed", http.StatusInternalServerError)
	}

	if editUser == nil {
		return nil, model.NewAppError("user.not_found", http.StatusNotFound)
	}

	return editUser, nil
}

func (a *App) DeactivateUser(user model.User, id string) (*model.User, *model.AppError) {
	if !a.SessionHasPermission(user, model.AdminPermissions.PermissionDeleteUsers) {
		return nil, model.NewAppError("permission.forbidden", http.StatusForbidden)
	}

	target, err := a.Store.User.Get(id)
	if err != nil {
		tlog.Errorw("Failed to retrieve user for deactivation",
			"user_id", id,
			"error", err,
		)
		return nil, model.NewAppError("user.retrieval_failed", http.StatusInternalServerError)
	}

	if target == nil {
		return nil, model.NewAppError("user.not_found", http.StatusNotFound)
	}

	if appErr := guardSystemAdminAccount(user, target); appErr != nil {
		return nil, appErr
	}

	if target.DeactivatedAt != 0 {
		exhausted, appErr := a.SeatsExhausted(context.Background())
		if appErr != nil {
			return nil, appErr
		}

		if exhausted {
			return nil, model.NewAppError("user.seat_limit", http.StatusPaymentRequired)
		}
	}

	deactivate, err := a.Store.User.Deactivate(target.ID, target.DeactivatedAt)
	if err != nil {
		tlog.Errorw("Failed to deactivate user",
			"user_id", id,
			"error", err,
		)
		return nil, model.NewAppError("user.deactivate_failed", http.StatusInternalServerError)
	}

	// Best-effort cascade: clean up memberships and task assignments when deactivating.
	// Only runs when transitioning active → deactivated (not on re-activation).
	isDeactivating := target.DeactivatedAt == 0
	target.DeactivatedAt = deactivate
	if isDeactivating {
		a.cleanupDeactivatedUser(target.ID)
	}

	return target, nil
}

// Best-effort cleanup: failures are logged as warnings, not returned, so they
// don't block the user's deactivation.
func (a *App) cleanupDeactivatedUser(userID string) {
	ctx := context.Background()

	// Get workspace IDs before removing from groups, since the lookup uses group membership.
	workspaceIDs, err := a.Store.Workspace.GetAllIDsForUser(userID)
	if err != nil {
		tlog.Warnw("Failed to get workspace IDs for deactivated user",
			"user_id", userID,
			"error", err,
		)
		return
	}

	// Remove from all groups after workspace lookup
	if err := a.Store.Groups.RemoveAllMembershipsForUser(ctx, userID); err != nil {
		tlog.Warnw("Failed to remove group memberships for deactivated user",
			"user_id", userID,
			"error", err,
		)
	}

	for _, wsID := range workspaceIDs {
		// Clear task assignments in this workspace
		rawTables, err := a.Store.Workspace.GetAllTablesBasic(wsID)
		if err != nil {
			tlog.Warnw("Failed to get tables for workspace cleanup",
				"workspace_id", wsID,
				"user_id", userID,
				"error", err,
			)
			continue
		}

		for _, rt := range rawTables {
			a.clearAllPersonFields(rt, userID)
		}
		// System cleanup, so the membership is removed without a permission check.
		if _, err := a.Store.Workspace.DeleteMember(wsID, userID); err != nil {
			tlog.Warnw("Failed to delete workspace member during deactivation cleanup",
				"workspace_id", wsID,
				"user_id", userID,
				"error", err,
			)
		}
	}
}

func (a *App) UpdateProfile(user model.User, u model.UserProfileRequest) *model.AppError {
	if u.Name == "" {
		return model.NewAppError("user.name_required", http.StatusBadRequest)
	}

	if u.LastName == "" {
		return model.NewAppError("user.lastname_required", http.StatusBadRequest)
	}

	if user.AuthService == "" && u.Email == "" {
		return model.NewAppError("user.email_required", http.StatusBadRequest)
	}

	tz, err := json.Marshal(map[string]any{
		"automaticTimezone":    u.AutomaticTimezone,
		"manualTimezone":       u.ManualTimezone,
		"useAutomaticTimezone": u.UseAutomaticTimezone,
	})
	if err != nil {
		tlog.Errorw("Failed to marshal timezone",
			"user_id", user.ID,
			"error", err,
		)
		return model.NewAppError("user.timezone_update_failed", http.StatusInternalServerError)
	}

	if _, err = a.Store.User.UpdateTimezone(user.ID, tz); err != nil {
		tlog.Errorw("Failed to update timezone",
			"user_id", user.ID,
			"error", err,
		)
		return model.NewAppError("user.timezone_update_failed", http.StatusInternalServerError)
	}

	clock := "12h"
	if u.ClockDisplay {
		clock = "24h"
	}

	if err = a.Store.Preferences.Update([]model.Preference{
		{
			UserID:   user.ID,
			Category: model.PreferencesCategoryDisplay,
			Name:     "clock_display",
			Value:    clock,
		},
		{
			UserID:   user.ID,
			Category: model.PreferencesCategoryDisplay,
			Name:     "language",
			Value:    u.Language,
		},
	}); err != nil {
		tlog.Errorw("Failed to update display preferences",
			"user_id", user.ID,
			"error", err,
		)
		return model.NewAppError("user.preferences_failed", http.StatusInternalServerError)
	}

	if _, err = a.Store.User.UpdateName(user.ID, u.Name, u.LastName); err != nil {
		tlog.Errorw("Failed to update user name",
			"user_id", user.ID,
			"error", err,
		)
		return model.NewAppError("user.update_failed", http.StatusInternalServerError)
	}

	if user.AuthService == "" {
		if _, err = a.Store.User.UpdateEmail(user.ID, u.Email); err != nil {
			tlog.Errorw("Failed to update user email",
				"user_id", user.ID,
				"error", err,
			)
			return model.NewAppError("user.update_failed", http.StatusInternalServerError)
		}
	}

	return nil
}

func (a *App) UpdatePassword(password, confirmed, oldPassword string, user model.User, currentSession string) *model.AppError {
	if password == "" {
		return model.NewAppError("user.password_required", http.StatusBadRequest)
	}

	if oldPassword == "" {
		return model.NewAppError("user.password_required", http.StatusBadRequest)
	}

	if confirmed != password {
		return model.NewAppError("user.password_mismatch", http.StatusBadRequest)
	}

	if !crypto.PasswordMatches(oldPassword, user.Password) {
		return model.NewAppError("user.password_incorrect", http.StatusBadRequest)
	}

	if _, appErr := a.isValidPassword(password); appErr != nil {
		return appErr
	}

	hashedPassword, err := crypto.HashPassword(password)
	if err != nil {
		tlog.Errorw("Failed to hash password",
			"user_id", user.ID,
			"error", err,
		)
		return model.NewAppError("user.update_failed", http.StatusInternalServerError)
	}

	if _, err = a.Store.User.UpdatePassword(user.ID, hashedPassword); err != nil {
		tlog.Errorw("Failed to update password",
			"user_id", user.ID,
			"error", err,
		)
		return model.NewAppError("user.update_failed", http.StatusInternalServerError)
	}

	if appErr := a.DeleteOtherUserSessions(user, currentSession); appErr != nil {
		return appErr
	}

	return nil
}

func (a *App) RevokeUserSessions(user model.User, id string) *model.AppError {
	if !a.SessionHasPermission(user, model.AdminPermissions.PermissionEditUsers) {
		return model.NewAppError("permission.forbidden", http.StatusForbidden)
	}

	u, err := a.Store.User.Get(id)
	if err != nil {
		tlog.Errorw("Failed to retrieve user for session revocation",
			"user_id", id,
			"error", err,
		)
		return model.NewAppError("user.retrieval_failed", http.StatusInternalServerError)
	}

	if u == nil {
		return model.NewAppError("user.not_found", http.StatusNotFound)
	}

	if appErr := guardSystemAdminAccount(user, u); appErr != nil {
		return appErr
	}

	return a.DeleteUserSessions(*u)
}

func (a *App) ResetUserPassword(user model.User, id string) *model.AppError {
	if !a.SessionHasPermission(user, model.AdminPermissions.PermissionEditUsers) {
		return model.NewAppError("permission.forbidden", http.StatusForbidden)
	}

	u, err := a.Store.User.Get(id)
	if err != nil {
		tlog.Errorw("Failed to retrieve user for password reset",
			"user_id", id,
			"error", err,
		)
		return model.NewAppError("user.retrieval_failed", http.StatusInternalServerError)
	}

	if u == nil {
		return model.NewAppError("user.not_found", http.StatusNotFound)
	}

	if appErr := guardSystemAdminAccount(user, u); appErr != nil {
		return appErr
	}

	token := crypto.RandomText()
	if err = a.Store.PasswordReset.Create(*u, crypto.HashSHA256(token)); err != nil {
		tlog.Errorw("Failed to create password reset token",
			"user_id", id,
			"error", err,
		)
		return model.NewAppError("user.password_reset_failed", http.StatusInternalServerError)
	}

	go a.SendPasswordResetEmail(*u, token)
	return nil
}

func (a *App) UploadPhoto(ctx context.Context, user model.User, file io.Reader) (string, *model.AppError) {
	primary, err := a.Store.Storage.GetPrimary()
	if err != nil {
		tlog.Errorw("Failed to retrieve primary storage for photo upload", "user_id", user.ID, "error", err)
		return "", model.NewAppError("user.photo_upload_failed", http.StatusInternalServerError)
	}

	if primary == nil {
		tlog.Errorw("No primary storage for photo upload", "user_id", user.ID)
		return "", model.NewAppError("user.photo_upload_failed", http.StatusInternalServerError)
	}

	backend, exists := a.FileStorageObjects[primary.ID]
	if !exists {
		tlog.Errorw("Primary storage backend not found", "user_id", user.ID, "storage_id", primary.ID)
		return "", model.NewAppError("user.photo_upload_failed", http.StatusInternalServerError)
	}

	data, readErr := io.ReadAll(file)
	if readErr != nil {
		tlog.Errorw("Failed to read photo upload", "user_id", user.ID, "error", readErr)
		return "", model.NewAppError("user.photo_upload_failed", http.StatusBadRequest)
	}

	photoID := model.NewID()
	photoPath := path.Join(primary.Directory, "photos", photoID)
	if err := backend.WriteFile(ctx, photoPath, bytes.NewReader(data), int64(len(data))); err != nil {
		tlog.Errorw("Failed to write photo to storage", "user_id", user.ID, "error", err)
		return "", model.NewAppError("user.photo_upload_failed", http.StatusInternalServerError)
	}

	dbCtx, cancel := a.dbCtx(ctx)
	defer cancel()

	old, err := a.Store.UserPhoto.Replace(dbCtx, model.UserPhoto{
		UserID: user.ID, PhotoID: photoID, StorageID: primary.ID,
	})
	if err != nil {
		if removeErr := backend.RemoveFile(ctx, photoPath); removeErr != nil {
			tlog.Warnw("Failed to remove unreferenced photo", "user_id", user.ID, "error", removeErr)
		}

		tlog.Errorw("Failed to update user photo", "user_id", user.ID, "error", err)
		return "", model.NewAppError("user.photo_upload_failed", http.StatusInternalServerError)
	}

	if old != nil && old.StorageID != "" {
		storage, err := a.Store.Storage.GetByID(old.StorageID)
		if err != nil || storage == nil {
			tlog.Warnw("Failed to retrieve replaced photo storage", "user_id", user.ID, "storage_id", old.StorageID, "error", err)
		} else if oldBackend, exists := a.FileStorageObjects[old.StorageID]; !exists || oldBackend == nil {
			tlog.Warnw("Replaced photo storage backend unavailable", "user_id", user.ID, "storage_id", old.StorageID)
		} else if err := oldBackend.RemoveFile(ctx, path.Join(storage.Directory, "photos", old.PhotoID)); err != nil {
			tlog.Warnw("Failed to remove replaced photo", "user_id", user.ID, "storage_id", old.StorageID, "error", err)
		}
	}

	return photoID, nil
}

func (a *App) DeletePhoto(ctx context.Context, user model.User) *model.AppError {
	dbCtx, cancel := a.dbCtx(ctx)
	defer cancel()

	photo, err := a.Store.UserPhoto.Get(dbCtx, user.ID)
	if err != nil {
		tlog.Errorw("Failed to retrieve photo record", "user_id", user.ID, "error", err)
		return model.NewAppError("user.photo_delete_failed", http.StatusInternalServerError)
	}

	if photo == nil {
		return nil
	}

	deleted, err := a.Store.UserPhoto.Delete(dbCtx, user.ID, photo.PhotoID)
	if err != nil {
		tlog.Errorw("Failed to clear user photo", "user_id", user.ID, "error", err)
		return model.NewAppError("user.photo_delete_failed", http.StatusInternalServerError)
	}

	if !deleted {
		return model.NewAppError("user.photo_delete_failed", http.StatusConflict)
	}

	storage, err := a.Store.Storage.GetByID(photo.StorageID)
	if err != nil {
		tlog.Warnw("Failed to retrieve deleted photo storage", "user_id", user.ID, "storage_id", photo.StorageID, "photo_id", photo.PhotoID, "error", err)
		return nil
	}

	if storage == nil {
		tlog.Warnw("Deleted photo storage not found", "user_id", user.ID, "storage_id", photo.StorageID, "photo_id", photo.PhotoID)
		return nil
	}

	backend, exists := a.FileStorageObjects[photo.StorageID]
	if !exists || backend == nil {
		tlog.Warnw("Deleted photo storage backend not found", "user_id", user.ID, "storage_id", photo.StorageID, "photo_id", photo.PhotoID)
		return nil
	}

	if err := backend.RemoveFile(ctx, path.Join(storage.Directory, "photos", photo.PhotoID)); err != nil {
		tlog.Warnw("Failed to remove deleted photo from storage", "user_id", user.ID, "storage_id", photo.StorageID, "photo_id", photo.PhotoID, "error", err)
	}

	return nil
}

func (a *App) GenerateMfa(user model.User) (*model.Mfa, *model.AppError) {
	secret, qr, err := mfa.GenerateSecret(user.Email)
	if err != nil {
		tlog.Errorw("Failed to generate MFA secret",
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("user.mfa_generate_failed", http.StatusInternalServerError)
	}

	if err = a.Store.User.UpdateMfaSecret(user, secret); err != nil {
		tlog.Errorw("Failed to save MFA secret",
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("user.mfa_generate_failed", http.StatusInternalServerError)
	}

	return &model.Mfa{
		Secret: secret,
		QRCode: base64.StdEncoding.EncodeToString(qr),
	}, nil
}

func (a *App) EnableMfa(user model.User, token string) *model.AppError {
	ok, err := mfa.ValidateToken(user.MfaSecret, token)
	if err != nil {
		tlog.Errorw("Failed to validate MFA token",
			"user_id", user.ID,
			"error", err,
		)
		return model.NewAppError("user.mfa_enable_failed", http.StatusInternalServerError)
	}

	if !ok {
		return model.NewAppError("user.mfa_enable_failed", http.StatusBadRequest)
	}

	if err := a.Store.User.UpdateMfaActive(user, true); err != nil {
		tlog.Errorw("Failed to enable MFA for user",
			"user_id", user.ID,
			"error", err,
		)
		return model.NewAppError("user.mfa_enable_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) DisableMfa(user model.User) *model.AppError {
	if err := a.Store.User.UpdateMfaActive(user, false); err != nil {
		tlog.Errorw("Failed to disable MFA for user",
			"user_id", user.ID,
			"error", err,
		)
		return model.NewAppError("user.mfa_disable_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) GetStatus() ([]model.UserStatus, *model.AppError) {
	statuses, err := a.Store.User.GetStatuses()
	if err != nil {
		tlog.Errorw("Failed to retrieve user statuses",
			"error", err,
		)
		return nil, model.NewAppError("user.status_failed", http.StatusInternalServerError)
	}

	return statuses, nil
}

func (a *App) SetUserStatus(userID string, status string, userDefined bool) *model.AppError {
	userStatus, err := a.Store.User.UpdateStatus(userID, status, userDefined)
	if err != nil {
		tlog.Errorw("Failed to set user status",
			"user_id", userID,
			"status", status,
			"error", err,
		)
		return model.NewAppError("user.status_failed", http.StatusInternalServerError)
	}

	if err = a.SendStatusEvent(*userStatus); err != nil {
		tlog.Errorw("Failed to send status event",
			"user_id", userID,
			"status", status,
			"error", err,
		)
		return model.NewAppError("user.status_failed", http.StatusInternalServerError)
	}

	return nil
}
