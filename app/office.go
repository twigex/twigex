// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

type officeDiscovery struct {
	XMLName  xml.Name        `xml:"wopi-discovery"`
	NetZones []officeNetZone `xml:"net-zone"`
}

type officeNetZone struct {
	Apps []officeApp `xml:"app"`
}

type officeApp struct {
	Name    string         `xml:"name,attr"`
	Actions []officeAction `xml:"action"`
}

type officeAction struct {
	Name   string `xml:"name,attr"`
	UrlSrc string `xml:"urlsrc,attr"`
}

type officeFileInfo struct {
	BaseFileName     string `json:"BaseFileName"`
	Size             int64  `json:"Size"`
	OwnerId          string `json:"OwnerId"`
	UserId           string `json:"UserId"`
	UserCanWrite     bool   `json:"UserCanWrite"`
	UserFriendlyName string `json:"UserFriendlyName"`
}

var euroOfficeSupportedTypes = map[string]bool{
	// Word processing
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.template": true,
	"application/vnd.ms-word.document.macroenabled.12":                        true,
	"application/vnd.ms-word.template.macroenabled.12":                        true,
	"application/vnd.oasis.opendocument.text":                                 true,
	"application/vnd.oasis.opendocument.text-template":                        true,
	"text/plain":           true,
	"application/rtf":      true,
	"text/rtf":             true,
	"text/html":            true,
	"application/epub+zip": true,
	// Spreadsheets
	"application/vnd.ms-excel": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":    true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.template": true,
	"application/vnd.ms-excel.sheet.macroenabled.12":                       true,
	"application/vnd.ms-excel.template.macroenabled.12":                    true,
	"application/vnd.oasis.opendocument.spreadsheet":                       true,
	"application/vnd.oasis.opendocument.spreadsheet-template":              true,
	"text/csv": true,
	// Presentations
	"application/vnd.ms-powerpoint":                                             true,
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
	"application/vnd.openxmlformats-officedocument.presentationml.template":     true,
	"application/vnd.openxmlformats-officedocument.presentationml.slideshow":    true,
	"application/vnd.ms-powerpoint.presentation.macroenabled.12":                true,
	"application/vnd.ms-powerpoint.template.macroenabled.12":                    true,
	"application/vnd.ms-powerpoint.slideshow.macroenabled.12":                   true,
	"application/vnd.oasis.opendocument.presentation":                           true,
	"application/vnd.oasis.opendocument.presentation-template":                  true,
}

func (a *App) OfficeSupportsFileType(fileType string) (bool, *model.AppError) {
	if *a.ConfigStore.Config.OfficeSettings.Type == model.OfficeTypeEuroOffice {
		return euroOfficeSupportedTypes[fileType], nil
	}

	discovery, appErr := a.fetchOfficeDiscovery()
	if appErr != nil {
		return false, appErr
	}

	for _, netZone := range discovery.NetZones {
		for _, app := range netZone.Apps {
			if app.Name == fileType {
				return true, nil
			}
		}
	}

	return false, nil
}

func (a *App) CreateOfficeURL(user model.User, wopiApp, fileID string) (string, *model.AppError) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  user.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(*a.ConfigStore.Config.OfficeSettings.Secret))
	if err != nil {
		tlog.Errorw("Failed to sign office token",
			"user_id", user.ID,
			"error", err,
		)
		return "", model.NewAppError("office.open_failed", http.StatusInternalServerError)
	}

	return a.buildCollaboraURL(tokenString, wopiApp, fileID)
}

func (a *App) buildCollaboraURL(tokenString, wopiApp, fileID string) (string, *model.AppError) {
	discovery, appErr := a.fetchOfficeDiscovery()
	if appErr != nil {
		return "", appErr
	}

	urlSrc := ""
outer:
	for _, netZone := range discovery.NetZones {
		for _, officeApp := range netZone.Apps {
			for _, action := range officeApp.Actions {
				urlSrc = action.UrlSrc
				break outer
			}
		}
	}

	wopiSrc := *a.ConfigStore.Config.ServerSettings.SiteURL + "/api/office/wopi/" + wopiApp + "/files/" + fileID
	client := "WOPISrc=" + url.QueryEscape(wopiSrc) +
		"&access_token=" + url.QueryEscape(tokenString) +
		"&closebutton=true"

	return urlSrc + client, nil
}

func (a *App) VerifyOfficeToken(accessToken string) (*jwt.Token, *model.AppError) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(*a.ConfigStore.Config.OfficeSettings.Secret), nil
	})
	if err != nil {
		return nil, model.NewAppError("office.token_invalid", http.StatusUnauthorized)
	}

	return token, nil
}

func (a *App) GetOfficeFile(fileID string, accessToken string) (*model.File, *model.AppError) {
	token, appErr := a.VerifyOfficeToken(accessToken)
	if appErr != nil {
		return nil, appErr
	}

	if shareToken, boundFile, isGuest := officeShareBinding(token); isGuest {
		return a.guestOfficeFile(shareToken, boundFile, fileID)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, model.NewAppError("office.token_invalid", http.StatusUnauthorized)
	}

	userID := claims["id"].(string)
	_, err := a.Store.User.Get(userID)
	if err != nil {
		tlog.Errorw("Failed to find user from office token",
			"user_id", userID,
			"error", err,
		)
		return nil, model.NewAppError("user.not_found", http.StatusInternalServerError)
	}

	file, err := a.Store.File.Get(fileID)
	if err != nil {
		tlog.Errorw("Failed to retrieve office file",
			"file_id", fileID,
			"error", err,
		)
		return nil, model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
	}

	if file == nil {
		return nil, model.NewAppError("file.not_found", http.StatusNotFound)
	}

	return file, nil
}

func (a *App) CheckFileInfo(fileID, accessToken, wopiApp string) (*[]byte, *model.AppError) {
	token, appErr := a.VerifyOfficeToken(accessToken)
	if appErr != nil {
		return nil, appErr
	}

	if shareToken, boundFile, gid, name, ro, isGuest := guestOfficeClaims(token); isGuest {
		return a.guestCheckFileInfo(shareToken, boundFile, fileID, gid, name, ro)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, model.NewAppError("office.token_invalid", http.StatusUnauthorized)
	}

	user, err := a.Store.User.Get(claims["id"].(string))
	if err != nil {
		tlog.Errorw("Failed to find user from office token",
			"error", err,
		)
		return nil, model.NewAppError("user.not_found", http.StatusInternalServerError)
	}

	if user == nil {
		return nil, model.NewAppError("user.not_found", http.StatusNotFound)
	}

	build := func(name string, size int64, owner string, canWrite bool) (*[]byte, *model.AppError) {
		b, err := json.Marshal(officeFileInfo{
			BaseFileName:     name,
			Size:             size,
			OwnerId:          owner,
			UserId:           user.ID,
			UserCanWrite:     canWrite,
			UserFriendlyName: user.Name + " " + user.LastName,
		})
		if err != nil {
			tlog.Errorw("Failed to marshal office file info",
				"file_id", fileID,
				"error", err,
			)
			return nil, model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
		}

		return &b, nil
	}

	switch wopiApp {
	case "chat":
		f, err := a.Store.Posts.GetAttachmentByID(fileID)
		if err != nil {
			tlog.Errorw("Failed to get chat attachment for office",
				"file_id", fileID,
				"error", err,
			)
			return nil, model.NewAppError("channel.attachment_not_found", http.StatusInternalServerError)
		}

		if f == nil {
			return nil, model.NewAppError("channel.attachment_not_found", http.StatusNotFound)
		}

		return build(f.Name, f.Size, f.UserID, false)

	case "projects":
		f, err := a.Store.Workspace.GetAttachmentByID(fileID)
		if err != nil {
			tlog.Errorw("Failed to get workspace attachment for office",
				"file_id", fileID,
				"error", err,
			)
			return nil, model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
		}

		if f == nil {
			return nil, model.NewAppError("collimato.not_found", http.StatusNotFound)
		}

		return build(f.Name, f.Size, f.UserID, false)

	default:
		file, appErr := a.HasPermission(fileID, *user)
		if appErr != nil {
			return nil, appErr
		}

		if file == nil {
			return nil, model.NewAppError("file.not_found", http.StatusNotFound)
		}

		return build(file.Name, file.Size, file.Owner, file.AccessLevel.CanEdit())
	}
}

func officeShareBinding(token *jwt.Token) (string, string, bool) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", false
	}

	shareToken, _ := claims["share"].(string)
	fileID, _ := claims["file"].(string)
	if shareToken != "" && fileID != "" {
		return shareToken, fileID, true
	}

	return "", "", false
}

// mintGuestOfficeToken issues a share-scoped office token bound to one file.
// gid is random per open so co-editors get distinct WOPI identities/cursors;
// ro=false only when the link permits editing.
func (a *App) mintGuestOfficeToken(shareToken, fileID string, canEdit bool, name string) (string, *model.AppError) {
	if name == "" {
		name = "Guest"
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"share": shareToken,
		"file":  fileID,
		"ro":    !canEdit,
		"gid":   model.NewID(),
		"name":  name,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, err := token.SignedString([]byte(*a.ConfigStore.Config.OfficeSettings.Secret))
	if err != nil {
		tlog.Errorw("Failed to sign guest office token", "error", err)
		return "", model.NewAppError("office.open_failed", http.StatusInternalServerError)
	}

	return tokenString, nil
}

// guestOfficeClaims extracts the guest-token fields. isGuest mirrors
// officeShareBinding (share+file present).
func guestOfficeClaims(token *jwt.Token) (share, file, gid, name string, ro, isGuest bool) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return
	}

	share, _ = claims["share"].(string)
	file, _ = claims["file"].(string)
	if share == "" || file == "" {
		return
	}

	isGuest = true
	gid, _ = claims["gid"].(string)
	name, _ = claims["name"].(string)
	ro, _ = claims["ro"].(bool)
	return
}

// validateGuestOfficeAccess enforces that a guest token may only reach the exact
// file it was minted for, and only while the underlying share still permits it.
func (a *App) validateGuestOfficeAccess(shareToken, boundFile, fileID string) (*model.File, *model.ExternalShare, *model.AppError) {
	if boundFile != fileID {
		return nil, nil, model.NewAppError("office.token_invalid", http.StatusForbidden)
	}

	link, _, appErr := a.resolvePublicShare(context.Background(), shareToken)
	if appErr != nil {
		return nil, nil, appErr
	}

	if !link.AllowView {
		return nil, nil, model.NewAppError("office.token_invalid", http.StatusForbidden)
	}

	file, err := a.Store.File.Get(fileID)
	if err != nil {
		tlog.Errorw("Failed to retrieve guest office file", "file_id", fileID, "error", err)
		return nil, nil, model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
	}

	if file == nil || file.DeletedAt != 0 {
		return nil, nil, model.NewAppError("file.not_found", http.StatusNotFound)
	}

	return file, link, nil
}

func (a *App) guestOfficeFile(shareToken, boundFile, fileID string) (*model.File, *model.AppError) {
	file, _, appErr := a.validateGuestOfficeAccess(shareToken, boundFile, fileID)
	return file, appErr
}

// guestCheckFileInfo reports write capability from the LIVE link (never the
// token alone): editing is allowed only when the token is not read-only AND the
// share still permits editing. gid gives each guest a distinct WOPI identity.
func (a *App) guestCheckFileInfo(shareToken, boundFile, fileID, gid, name string, ro bool) (*[]byte, *model.AppError) {
	file, link, appErr := a.validateGuestOfficeAccess(shareToken, boundFile, fileID)
	if appErr != nil {
		return nil, appErr
	}

	if gid == "" {
		gid = "guest"
	}

	if name == "" {
		name = "Guest"
	}

	b, err := json.Marshal(officeFileInfo{
		BaseFileName:     file.Name,
		Size:             file.Size,
		OwnerId:          file.Owner,
		UserId:           gid,
		UserCanWrite:     !ro && link.AllowEdit,
		UserFriendlyName: name,
	})
	if err != nil {
		tlog.Errorw("Failed to marshal guest office file info", "file_id", fileID, "error", err)
		return nil, model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
	}

	return &b, nil
}

// validateGuestOfficeEdit is the strict gate for guest WRITES. Beyond the read
// checks it re-confirms the LIVE link still permits editing and that the file is
// still inside the shared subtree (not moved out / trashed since the token was
// minted). Owner revocation therefore takes effect immediately.
func (a *App) validateGuestOfficeEdit(shareToken, boundFile, fileID string) (*model.File, *model.AppError) {
	if boundFile != fileID {
		return nil, model.NewAppError("office.token_invalid", http.StatusForbidden)
	}

	link, root, appErr := a.resolvePublicShare(context.Background(), shareToken)
	if appErr != nil {
		return nil, appErr
	}

	if !link.AllowView || !link.AllowEdit {
		return nil, model.NewAppError("office.readonly", http.StatusForbidden)
	}

	if fileID == root.ID {
		if root.IsFolder {
			return nil, model.NewAppError("office.token_invalid", http.StatusForbidden)
		}

		return root, nil
	}

	if !root.IsFolder {
		return nil, model.NewAppError("office.token_invalid", http.StatusForbidden)
	}

	return a.publicSubtreeFile(*root, fileID)
}

// checkOwnerSpaceForEdit guards against unauthenticated writes growing the
// owner's storage past their limit. Returns a generic error (never reveals the
// owner's storage state) when space is low.
func (a *App) checkOwnerSpaceForEdit(ownerID string, delta int64) *model.AppError {
	if delta <= 0 {
		return nil
	}

	owner, err := a.Store.User.Get(ownerID)
	if err != nil || owner == nil {
		tlog.Errorw("Failed to retrieve owner for guest office save", "owner_id", ownerID, "error", err)
		return model.NewAppError("office.save_failed", http.StatusInternalServerError)
	}

	free, err := a.isFreeSpace(*owner, delta)
	if err != nil {
		tlog.Errorw("Failed to check owner storage limit for guest office save", "owner_id", ownerID, "error", err)
		return model.NewAppError("office.save_failed", http.StatusInternalServerError)
	}

	if !free {
		tlog.Warnw("Guest office save rejected: owner storage running low", "owner_id", ownerID)
		return model.NewAppError("office.save_failed", http.StatusInsufficientStorage)
	}

	return nil
}

func (a *App) OpenPublicOfficeFile(ctx context.Context, shareToken, childID, accessToken, nickname string) (*model.OfficeOpenResult, *model.AppError) {
	link, file, appErr := a.resolvePublicShare(ctx, shareToken)
	if appErr != nil {
		return nil, appErr
	}

	if !a.validatePublicAccess(link, accessToken) {
		return nil, model.NewAppError("link.password_required", http.StatusUnauthorized)
	}

	if !link.AllowView {
		return nil, model.NewAppError("link.view_forbidden", http.StatusForbidden)
	}

	target := file
	if childID != "" && childID != file.ID {
		if !file.IsFolder {
			return nil, model.NewAppError("link.not_found", http.StatusNotFound)
		}

		descendant, appErr := a.publicSubtreeFile(*file, childID)
		if appErr != nil {
			return nil, appErr
		}

		target = descendant
	}

	if target.IsFolder {
		return nil, model.NewAppError("file.unknown_type", http.StatusBadRequest)
	}

	supported, appErr := a.OfficeSupportsFileType(target.Type)
	if appErr != nil {
		return nil, appErr
	}

	if !supported {
		return nil, model.NewAppError("office.not_available", http.StatusBadRequest)
	}

	canEdit := link.AllowEdit

	guestToken, appErr := a.mintGuestOfficeToken(shareToken, target.ID, canEdit, nickname)
	if appErr != nil {
		return nil, appErr
	}

	if *a.ConfigStore.Config.OfficeSettings.Type == model.OfficeTypeEuroOffice {
		return a.openGuestEuroOffice(guestToken, *target, link.AllowDownload, canEdit, nickname)
	}

	return a.openGuestCollabora(guestToken, target.ID)
}

func (a *App) openGuestCollabora(guestToken, fileID string) (*model.OfficeOpenResult, *model.AppError) {
	officeURL, appErr := a.buildCollaboraURL(guestToken, "files", fileID)
	if appErr != nil {
		return nil, appErr
	}

	return &model.OfficeOpenResult{
		Type: model.OfficeTypeCollabora,
		URL:  officeURL,
	}, nil
}

func (a *App) openGuestEuroOffice(guestToken string, file model.File, allowDownload, canEdit bool, name string) (*model.OfficeOpenResult, *model.AppError) {
	siteURL := *a.ConfigStore.Config.ServerSettings.SiteURL
	fileURL := siteURL + "/api/office/eurooffice/file/files/" + file.ID + "?access_token=" + url.QueryEscape(guestToken)
	callbackURL := siteURL + "/api/office/eurooffice/callback/files/" + file.ID + "?access_token=" + url.QueryEscape(guestToken)

	fileExt := strings.TrimPrefix(filepath.Ext(file.Name), ".")
	docKey := fmt.Sprintf("%s_%d", file.ID, file.Modified)

	mode := "view"
	if canEdit {
		mode = "edit"
	}

	if name == "" {
		name = "Guest"
	}
	// Distinct per-open id so co-editors are separate participants.
	guestID := model.NewID()

	configPayload := jwt.MapClaims{
		"document": map[string]interface{}{
			"fileType": fileExt,
			"key":      docKey,
			"title":    file.Name,
			"url":      fileURL,
			"permissions": map[string]interface{}{
				"edit":     canEdit,
				"download": allowDownload,
				"print":    allowDownload,
				"chat":     false,
			},
		},
		"documentType": euroOfficeDocumentType(file.Type),
		"editorConfig": map[string]interface{}{
			"callbackUrl": callbackURL,
			"mode":        mode,
			"user": map[string]interface{}{
				"id":   guestID,
				"name": name,
			},
		},
	}
	configToken := jwt.NewWithClaims(jwt.SigningMethodHS256, configPayload)
	configTokenString, err := configToken.SignedString([]byte(*a.ConfigStore.Config.OfficeSettings.Secret))
	if err != nil {
		tlog.Errorw("Failed to sign guest EuroOffice config token", "file_id", file.ID, "error", err)
		return nil, model.NewAppError("office.open_failed", http.StatusInternalServerError)
	}

	return &model.OfficeOpenResult{
		Type: model.OfficeTypeEuroOffice,
		Host: *a.ConfigStore.Config.OfficeSettings.Host,
		Config: &model.EuroOfficeConfig{
			Document: model.EuroOfficeDocument{
				FileType: fileExt,
				Key:      docKey,
				Title:    file.Name,
				URL:      fileURL,
				Permissions: model.EuroOfficePermissions{
					Edit:     canEdit,
					Download: allowDownload,
					Print:    allowDownload,
					Chat:     false,
				},
			},
			DocumentType: euroOfficeDocumentType(file.Type),
			EditorConfig: model.EuroOfficeEditorConfig{
				CallbackURL: callbackURL,
				Mode:        mode,
				User: model.EuroOfficeUser{
					ID:   guestID,
					Name: name,
				},
			},
			Token: configTokenString,
		},
	}, nil
}

func (a *App) OfficeSaveFile(fileID string, accessToken string, body io.ReadCloser) *model.AppError {
	defer body.Close()

	token, appErr := a.VerifyOfficeToken(accessToken)
	if appErr != nil {
		return appErr
	}

	var f *model.File
	guest := false
	if shareToken, boundFile, _, _, ro, isGuest := guestOfficeClaims(token); isGuest {
		guest = true
		if ro {
			return model.NewAppError("office.readonly", http.StatusForbidden)
		}

		file, appErr := a.validateGuestOfficeEdit(shareToken, boundFile, fileID)
		if appErr != nil {
			return appErr
		}

		f = file
	} else {
		file, err := a.Store.File.Get(fileID)
		if err != nil {
			tlog.Errorw("Failed to retrieve file",
				"file_id", fileID,
				"error", err,
			)
			return model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
		}

		if file == nil {
			return model.NewAppError("file.not_found", http.StatusNotFound)
		}

		f = file
	}

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		tlog.Errorw("Failed to read office save body",
			"file_id", fileID,
			"error", err,
		)
		return model.NewAppError("office.save_failed", http.StatusInternalServerError)
	}

	if guest {
		if appErr := a.checkOwnerSpaceForEdit(f.Owner, int64(len(bodyBytes))-f.Size); appErr != nil {
			return appErr
		}
	}

	officeFilePath, err := a.BuildFilePath(f.Storage, f.ID, model.AppFiles)
	if err != nil {
		tlog.Errorw("Failed to build file path",
			"file_id", fileID,
			"storage", f.Storage,
			"error", err,
		)
		return model.NewAppError("office.save_failed", http.StatusInternalServerError)
	}

	if err = a.FileStorageObjects[f.Storage].WriteFile(context.Background(), officeFilePath, bytes.NewReader(bodyBytes), int64(len(bodyBytes))); err != nil {
		tlog.Errorw("Failed to write file to storage",
			"file_id", fileID,
			"storage", f.Storage,
			"error", err,
		)
		return model.NewAppError("office.save_failed", http.StatusInternalServerError)
	}

	size, err := a.FileStorageObjects[f.Storage].FileSize(officeFilePath)
	if err != nil {
		tlog.Errorw("Failed to retrieve file size",
			"file_id", fileID,
			"error", err,
		)
		return model.NewAppError("office.save_failed", http.StatusInternalServerError)
	}

	if err = a.Store.File.UpdateSize(size, f.ID); err != nil {
		tlog.Errorw("Failed to update file size",
			"file_id", fileID,
			"error", err,
		)
		return model.NewAppError("office.save_failed", http.StatusInternalServerError)
	}

	if guest {
		a.recordGuestOfficeEdit(f)
	}

	return nil
}

// recordGuestOfficeEdit logs a public-link edit. Attributed to the owner (the
// activity join requires a real user_id and there is no authenticated editor);
// the frontend renders it without an actor name.
func (a *App) recordGuestOfficeEdit(f *model.File) {
	a.RecordActivity(f.Owner, model.AppFiles, model.ActivityFilePublicEdit, f.Parent, f.ID, map[string]any{
		"name": f.DisplayName,
		"type": f.Type,
	})
}

func (a *App) OpenOfficeFile(fileID string, user model.User, wopiApp string) (*model.OfficeOpenResult, *model.AppError) {
	officeType := *a.ConfigStore.Config.OfficeSettings.Type

	if officeType == model.OfficeTypeEuroOffice {
		return a.openEuroOfficeFile(fileID, user, wopiApp)
	}

	return a.openCollaboraFile(fileID, user, wopiApp)
}

func (a *App) openCollaboraFile(fileID string, user model.User, wopiApp string) (*model.OfficeOpenResult, *model.AppError) {
	switch wopiApp {
	case "files":
		file, appErr := a.HasPermission(fileID, user)
		if appErr != nil {
			return nil, appErr
		}

		if file == nil {
			return nil, model.NewAppError("file.not_found", http.StatusNotFound)
		}
	case "projects":
		file, appErr := a.GetUserWorkspaceAttachment(fileID, user)
		if appErr != nil {
			return nil, appErr
		}

		if file == nil {
			return nil, model.NewAppError("file.not_found", http.StatusNotFound)
		}
	}

	officeURL, appErr := a.CreateOfficeURL(user, wopiApp, fileID)
	if appErr != nil {
		return nil, appErr
	}

	return &model.OfficeOpenResult{
		Type: model.OfficeTypeCollabora,
		URL:  officeURL,
	}, nil
}

func (a *App) openEuroOfficeFile(fileID string, user model.User, wopiApp string) (*model.OfficeOpenResult, *model.AppError) {
	var fileName, fileType string
	var canWrite bool
	var modified int64

	switch wopiApp {
	case "files":
		file, appErr := a.HasPermission(fileID, user)
		if appErr != nil {
			return nil, appErr
		}

		if file == nil {
			return nil, model.NewAppError("file.not_found", http.StatusNotFound)
		}

		fileName = file.Name
		fileType = file.Type
		canWrite = file.AccessLevel.CanEdit()
		modified = file.Modified

	case "chat":
		f, err := a.Store.Posts.GetAttachmentByID(fileID)
		if err != nil || f == nil {
			return nil, model.NewAppError("channel.attachment_not_found", http.StatusNotFound)
		}

		fileName = f.Name
		fileType = f.MimeType
		canWrite = false
		modified = f.UpdatedAt

	case "projects":
		f, appErr := a.GetUserWorkspaceAttachment(fileID, user)
		if appErr != nil {
			return nil, appErr
		}

		if f == nil {
			return nil, model.NewAppError("file.not_found", http.StatusNotFound)
		}

		fileName = f.Name
		fileType = f.MimeType
		canWrite = false
		modified = f.UpdatedAt
	}

	config, appErr := a.CreateEuroOfficeConfig(user, wopiApp, fileID, modified, canWrite, fileName, fileType)
	if appErr != nil {
		return nil, appErr
	}

	return &model.OfficeOpenResult{
		Type:   model.OfficeTypeEuroOffice,
		Host:   *a.ConfigStore.Config.OfficeSettings.Host,
		Config: config,
	}, nil
}

func (a *App) CreateEuroOfficeConfig(user model.User, wopiApp, fileID string, modified int64, canWrite bool, fileName, fileType string) (*model.EuroOfficeConfig, *model.AppError) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  user.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, err := token.SignedString([]byte(*a.ConfigStore.Config.OfficeSettings.Secret))
	if err != nil {
		tlog.Errorw("Failed to sign EuroOffice token", "user_id", user.ID, "error", err)
		return nil, model.NewAppError("office.open_failed", http.StatusInternalServerError)
	}

	siteURL := *a.ConfigStore.Config.ServerSettings.SiteURL
	fileURL := siteURL + "/api/office/eurooffice/file/" + wopiApp + "/" + fileID + "?access_token=" + url.QueryEscape(tokenString)
	callbackURL := siteURL + "/api/office/eurooffice/callback/" + wopiApp + "/" + fileID + "?access_token=" + url.QueryEscape(tokenString)

	mode := "edit"
	if !canWrite {
		mode = "view"
	}

	fileExt := strings.TrimPrefix(filepath.Ext(fileName), ".")

	docKey := fmt.Sprintf("%s_%d", fileID, modified)

	permissions := map[string]interface{}{
		"edit":     canWrite,
		"download": false,
		"print":    false,
		"chat":     false,
	}

	configPayload := map[string]interface{}{
		"document": map[string]interface{}{
			"fileType":    fileExt,
			"key":         docKey,
			"title":       fileName,
			"url":         fileURL,
			"permissions": permissions,
		},
		"documentType": euroOfficeDocumentType(fileType),
		"editorConfig": map[string]interface{}{
			"callbackUrl": callbackURL,
			"mode":        mode,
			"user": map[string]interface{}{
				"id":   user.ID,
				"name": user.Name + " " + user.LastName,
			},
		},
	}

	configToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"document":     configPayload["document"],
		"documentType": configPayload["documentType"],
		"editorConfig": configPayload["editorConfig"],
	})
	configTokenString, err := configToken.SignedString([]byte(*a.ConfigStore.Config.OfficeSettings.Secret))
	if err != nil {
		tlog.Errorw("Failed to sign EuroOffice config token", "user_id", user.ID, "error", err)
		return nil, model.NewAppError("office.open_failed", http.StatusInternalServerError)
	}

	return &model.EuroOfficeConfig{
		Document: model.EuroOfficeDocument{
			FileType: fileExt,
			Key:      docKey,
			Title:    fileName,
			URL:      fileURL,
			Permissions: model.EuroOfficePermissions{
				Edit:     canWrite,
				Download: false,
				Print:    false,
				Chat:     false, // Disable chat for documents for now
			},
		},
		DocumentType: euroOfficeDocumentType(fileType),
		EditorConfig: model.EuroOfficeEditorConfig{
			CallbackURL: callbackURL,
			Mode:        mode,
			User: model.EuroOfficeUser{
				ID:   user.ID,
				Name: user.Name + " " + user.LastName,
			},
		},
		Token: configTokenString,
	}, nil
}

func (a *App) EuroOfficeCallback(fileID, wopiApp, accessToken string, body io.ReadCloser) *model.AppError {
	defer body.Close()

	token, appErr := a.VerifyOfficeToken(accessToken)
	if appErr != nil {
		return appErr
	}

	shareToken, boundFile, _, _, ro, isGuest := guestOfficeClaims(token)
	if isGuest && ro {
		return nil
	}

	var cb struct {
		Status int    `json:"status"`
		URL    string `json:"url"`
	}
	if err := json.NewDecoder(body).Decode(&cb); err != nil {
		return model.NewAppError("office.save_failed", http.StatusBadRequest)
	}

	// Status 2 means EuroOffice is ready to save the document
	if cb.Status != 2 {
		return nil
	}

	if wopiApp != "files" {
		return nil
	}

	var f *model.File
	if isGuest {
		file, appErr := a.validateGuestOfficeEdit(shareToken, boundFile, fileID)
		if appErr != nil {
			return appErr
		}

		f = file
	} else {
		file, err := a.Store.File.Get(fileID)
		if err != nil {
			tlog.Errorw("Failed to retrieve file for EuroOffice save", "file_id", fileID, "error", err)
			return model.NewAppError("file.retrieval_failed", http.StatusInternalServerError)
		}

		if file == nil {
			return model.NewAppError("file.not_found", http.StatusNotFound)
		}

		f = file
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(cb.URL)
	if err != nil {
		tlog.Errorw("Failed to download file from EuroOffice", "file_id", fileID, "error", err)
		return model.NewAppError("office.save_failed", http.StatusInternalServerError)
	}

	defer resp.Body.Close()

	fileBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		tlog.Errorw("Failed to read EuroOffice file content", "file_id", fileID, "error", err)
		return model.NewAppError("office.save_failed", http.StatusInternalServerError)
	}

	if isGuest {
		if appErr := a.checkOwnerSpaceForEdit(f.Owner, int64(len(fileBytes))-f.Size); appErr != nil {
			return appErr
		}
	}

	officeFilePath, err := a.BuildFilePath(f.Storage, f.ID, model.AppFiles)
	if err != nil {
		tlog.Errorw("Failed to build file path for EuroOffice save", "file_id", fileID, "error", err)
		return model.NewAppError("office.save_failed", http.StatusInternalServerError)
	}

	if err = a.FileStorageObjects[f.Storage].WriteFile(context.Background(), officeFilePath, bytes.NewReader(fileBytes), int64(len(fileBytes))); err != nil {
		tlog.Errorw("Failed to write EuroOffice file to storage", "file_id", fileID, "error", err)
		return model.NewAppError("office.save_failed", http.StatusInternalServerError)
	}

	size, err := a.FileStorageObjects[f.Storage].FileSize(officeFilePath)
	if err != nil {
		tlog.Errorw("Failed to retrieve file size after EuroOffice save", "file_id", fileID, "error", err)
		return model.NewAppError("office.save_failed", http.StatusInternalServerError)
	}

	if err = a.Store.File.UpdateSize(size, f.ID); err != nil {
		tlog.Errorw("Failed to update file size after EuroOffice save", "file_id", fileID, "error", err)
		return model.NewAppError("office.save_failed", http.StatusInternalServerError)
	}

	if isGuest {
		a.recordGuestOfficeEdit(f)
	}

	return nil
}

func euroOfficeDocumentType(mimeType string) string {
	lower := strings.ToLower(mimeType)
	switch {
	case strings.Contains(lower, "spreadsheet") ||
		strings.Contains(lower, "excel") ||
		strings.HasSuffix(lower, "/csv") ||
		strings.Contains(lower, "opendocument.spreadsheet"):
		return "cell"
	case strings.Contains(lower, "presentation") ||
		strings.Contains(lower, "powerpoint") ||
		strings.Contains(lower, "opendocument.presentation"):
		return "slide"
	default:
		return "word"
	}
}

func (a *App) ServeOfficeChatAttachment(id string, w io.Writer) *model.AppError {
	file, err := a.Store.Posts.GetAttachmentByID(id)
	if err != nil {
		tlog.Errorw("Failed to get chat attachment",
			"attachment_id", id,
			"error", err,
		)
		return model.NewAppError("channel.attachment_not_found", http.StatusInternalServerError)
	}

	if file == nil {
		return model.NewAppError("channel.attachment_not_found", http.StatusNotFound)
	}

	if file.StorageID != "" {
		filePath, err := a.BuildFilePath(file.StorageID, file.ChannelID+"/"+file.ID, model.AppChat)
		if err != nil {
			tlog.Errorw("Failed to build chat attachment path",
				"attachment_id", id,
				"error", err,
			)
			return model.NewAppError("office.open_failed", http.StatusInternalServerError)
		}

		backend, exists := a.FileStorageObjects[file.StorageID]
		if !exists {
			return model.NewAppError("storage.not_available", http.StatusInternalServerError)
		}

		reader, err := backend.ReadFile(context.Background(), filePath)
		if err != nil {
			tlog.Errorw("Failed to read chat attachment from storage",
				"attachment_id", id,
				"error", err,
			)
			return model.NewAppError("office.open_failed", http.StatusInternalServerError)
		}

		defer reader.Close()

		if _, err = io.Copy(w, reader); err != nil {
			tlog.Errorw("Failed to stream chat attachment",
				"attachment_id", id,
				"error", err,
			)
			return model.NewAppError("office.open_failed", http.StatusInternalServerError)
		}

		return nil
	}

	filePath := path.Join("data", "chat", file.ChannelID, file.ID)
	f, err := os.Open(filePath)
	if err != nil {
		tlog.Errorw("Failed to open chat attachment",
			"attachment_id", id,
			"path", filePath,
			"error", err,
		)
		return model.NewAppError("office.open_failed", http.StatusInternalServerError)
	}

	defer f.Close()

	if _, err = io.Copy(w, f); err != nil {
		tlog.Errorw("Failed to stream chat attachment",
			"attachment_id", id,
			"error", err,
		)
		return model.NewAppError("office.open_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) ServeOfficeWorkspaceAttachment(id string, w io.Writer) *model.AppError {
	attachment, err := a.Store.Workspace.GetAttachmentByID(id)
	if err != nil {
		tlog.Errorw("Failed to get workspace attachment",
			"attachment_id", id,
			"error", err,
		)
		return model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
	}

	if attachment == nil {
		return model.NewAppError("collimato.not_found", http.StatusNotFound)
	}

	fileExt := filepath.Ext(attachment.Name)

	if attachment.StorageID != "" {
		subPath := attachment.WorkspaceID + "/" + attachment.TableID + "/" + attachment.TaskID + "/" + attachment.ID + fileExt
		filePath, pathErr := a.BuildFilePath(attachment.StorageID, subPath, model.AppProjects)
		if pathErr != nil {
			tlog.Errorw("Failed to build project file path", "attachment_id", id, "error", pathErr)
			return model.NewAppError("office.open_failed", http.StatusInternalServerError)
		}

		backend, exists := a.FileStorageObjects[attachment.StorageID]
		if !exists {
			tlog.Errorw("Storage backend not available", "storage_id", attachment.StorageID)
			return model.NewAppError("storage.not_available", http.StatusInternalServerError)
		}

		rc, readErr := backend.ReadFile(context.Background(), filePath)
		if readErr != nil {
			tlog.Errorw("Failed to open workspace attachment from storage",
				"attachment_id", id,
				"path", filePath,
				"error", readErr,
			)
			return model.NewAppError("office.open_failed", http.StatusInternalServerError)
		}

		defer rc.Close()
		if _, err = io.Copy(w, rc); err != nil {
			tlog.Errorw("Failed to stream workspace attachment", "attachment_id", id, "error", err)
			return model.NewAppError("office.open_failed", http.StatusInternalServerError)
		}

		return nil
	}

	filePath := path.Join("data", "projects",
		attachment.WorkspaceID,
		attachment.TableID,
		attachment.TaskID,
		attachment.ID+fileExt)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return model.NewAppError("collimato.not_found", http.StatusNotFound)
	}

	f, err := os.Open(filePath)
	if err != nil {
		tlog.Errorw("Failed to open workspace attachment",
			"attachment_id", id,
			"path", filePath,
			"error", err,
		)
		return model.NewAppError("office.open_failed", http.StatusInternalServerError)
	}

	defer f.Close()

	if _, err = io.Copy(w, f); err != nil {
		tlog.Errorw("Failed to stream workspace attachment",
			"attachment_id", id,
			"error", err,
		)
		return model.NewAppError("office.open_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) fetchOfficeDiscovery() (*officeDiscovery, *model.AppError) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest(http.MethodGet, *a.ConfigStore.Config.OfficeSettings.Host+"/hosting/discovery", nil)
	if err != nil {
		tlog.Errorw("Failed to create office discovery request",
			"error", err,
		)
		return nil, model.NewAppError("office.not_available", http.StatusInternalServerError)
	}

	resp, err := client.Do(req)
	if err != nil {
		tlog.Errorw("Failed to fetch office discovery",
			"error", err,
		)
		return nil, model.NewAppError("office.not_available", http.StatusInternalServerError)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		tlog.Errorw("Unexpected status from office discovery",
			"status", resp.StatusCode,
		)
		return nil, model.NewAppError("office.not_available", http.StatusInternalServerError)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		tlog.Errorw("Failed to read office discovery response",
			"error", err,
		)
		return nil, model.NewAppError("office.not_available", http.StatusInternalServerError)
	}

	var discovery officeDiscovery
	if err = xml.Unmarshal(body, &discovery); err != nil {
		tlog.Errorw("Failed to parse office discovery XML",
			"error", err,
		)
		return nil, model.NewAppError("office.not_available", http.StatusInternalServerError)
	}

	return &discovery, nil
}
