// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/twigex/twigex/model"
)

func signWithSecret(t *testing.T, secret string, claims jwt.MapClaims) string {
	t.Helper()

	claims["exp"] = time.Now().Add(time.Hour).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	return signed
}

func TestVerifyOfficeToken_RejectsTokenWhenSecretIsEmpty(t *testing.T) {
	a, _ := newOfficeAccessApp()
	a.ConfigStore.Config.OfficeSettings.Secret = model.NewString("")
	token := signWithSecret(t, "", jwt.MapClaims{"id": "user-1", "app": "files", "file": "doc-1"})

	_, appErr := a.VerifyOfficeToken(token)
	requireAppErr(t, appErr, "office.token_invalid", http.StatusUnauthorized)
}

func TestVerifyOfficeToken_RejectsTokenWhenSecretIsWhitespace(t *testing.T) {
	a, _ := newOfficeAccessApp()
	a.ConfigStore.Config.OfficeSettings.Secret = model.NewString("   ")
	token := signWithSecret(t, "   ", jwt.MapClaims{"id": "user-1", "app": "files", "file": "doc-1"})

	_, appErr := a.VerifyOfficeToken(token)
	requireAppErr(t, appErr, "office.token_invalid", http.StatusUnauthorized)
}

func TestVerifyOfficeToken_AcceptsTokenSignedWithConfiguredSecret(t *testing.T) {
	a, _ := newOfficeAccessApp()

	_, appErr := a.VerifyOfficeToken(userOfficeToken(t, "files", "doc-1"))
	if appErr != nil {
		t.Fatalf("expected token to verify, got %v", appErr)
	}
}

func TestOfficeAvailable(t *testing.T) {
	cases := []struct {
		name    string
		enabled bool
		secret  string
		want    bool
	}{
		{"enabled with secret", true, officeTestSecret, true},
		{"disabled with secret", false, officeTestSecret, false},
		{"enabled without secret", true, "", false},
		{"enabled with whitespace secret", true, " \t", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a, _ := newOfficeAccessApp()
			a.ConfigStore.Config.OfficeSettings.Enable = model.NewBool(c.enabled)
			a.ConfigStore.Config.OfficeSettings.Secret = model.NewString(c.secret)

			if got := a.OfficeAvailable(); got != c.want {
				t.Errorf("want %v, got %v", c.want, got)
			}
		})
	}
}

func TestEuroOfficeCallback_RejectsUnsignedDownloadURL(t *testing.T) {
	a, _ := newOfficeAccessApp()

	fetched := false
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fetched = true
	}))
	defer target.Close()

	body := io.NopCloser(strings.NewReader(`{"status":2,"url":"` + target.URL + `"}`))

	appErr := a.EuroOfficeCallback("doc-1", "files", userOfficeToken(t, "files", "doc-1"), "", body)
	requireAppErr(t, appErr, "office.token_invalid", http.StatusUnauthorized)

	if fetched {
		t.Fatal("the unsigned callback URL was fetched")
	}
}

func TestEuroOfficeCallback_RejectsCallbackSignedWithAnotherSecret(t *testing.T) {
	a, _ := newOfficeAccessApp()
	signed := signWithSecret(t, "another-secret", jwt.MapClaims{"status": 2, "url": "http://127.0.0.1/"})
	body := io.NopCloser(strings.NewReader(`{"token":"` + signed + `"}`))

	appErr := a.EuroOfficeCallback("doc-1", "files", userOfficeToken(t, "files", "doc-1"), "", body)
	requireAppErr(t, appErr, "office.token_invalid", http.StatusUnauthorized)
}

func TestVerifyEuroOfficeCallback_ReadsBodyToken(t *testing.T) {
	a, _ := newOfficeAccessApp()
	signed := signOfficeTestToken(t, jwt.MapClaims{"status": 2, "url": "http://office/doc"})

	cb, appErr := a.verifyEuroOfficeCallback(signed, "")
	if appErr != nil {
		t.Fatalf("expected callback to verify, got %v", appErr)
	}

	if cb.Status != 2 || cb.URL != "http://office/doc" {
		t.Fatalf("unexpected callback %+v", cb)
	}
}

func TestVerifyEuroOfficeCallback_ReadsHeaderPayload(t *testing.T) {
	a, _ := newOfficeAccessApp()
	payload := map[string]any{"status": 2, "url": "http://office/doc"}
	signed := signOfficeTestToken(t, jwt.MapClaims{"payload": payload})

	cb, appErr := a.verifyEuroOfficeCallback("", "Bearer "+signed)
	if appErr != nil {
		t.Fatalf("expected callback to verify, got %v", appErr)
	}

	if cb.Status != 2 || cb.URL != "http://office/doc" {
		t.Fatalf("unexpected callback %+v", cb)
	}
}

func TestVerifyEuroOfficeCallback_RejectsWhenSecretIsEmpty(t *testing.T) {
	a, _ := newOfficeAccessApp()
	a.ConfigStore.Config.OfficeSettings.Secret = model.NewString("")
	signed := signWithSecret(t, "", jwt.MapClaims{"status": 2, "url": "http://office/doc"})

	_, appErr := a.verifyEuroOfficeCallback(signed, "")
	requireAppErr(t, appErr, "office.token_invalid", http.StatusUnauthorized)
}

func TestOpenOfficeFile_RefusesWhenOfficeUnavailable(t *testing.T) {
	a, _ := newOfficeAccessApp()
	a.ConfigStore.Config.OfficeSettings.Secret = model.NewString("")

	_, appErr := a.OpenOfficeFile("doc-1", model.User{ID: "user-1"}, "files")
	requireAppErr(t, appErr, "office.not_available", http.StatusNotFound)
}

func TestOpenPublicOfficeFile_RefusesWhenOfficeUnavailable(t *testing.T) {
	a, _ := newOfficeAccessApp()
	a.ConfigStore.Config.OfficeSettings.Enable = model.NewBool(false)

	_, appErr := a.OpenPublicOfficeFile(context.Background(), "share-token", "", "", "Guest")
	requireAppErr(t, appErr, "office.not_available", http.StatusNotFound)
}

func TestEuroOfficeCallback_KeepsDocumentWhenDownloadFails(t *testing.T) {
	a, stores := newOfficeAccessApp()
	stores.files.shares["user-1/doc-1"] = []model.File{{AccessLevel: model.AccessEditor}}

	failing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "expired", http.StatusNotFound)
	}))
	defer failing.Close()

	signed := signOfficeTestToken(t, jwt.MapClaims{"status": 2, "url": failing.URL + "/output.docx"})
	body := io.NopCloser(strings.NewReader(`{"token":"` + signed + `"}`))

	appErr := a.EuroOfficeCallback("doc-1", "files", userOfficeToken(t, "files", "doc-1"), "", body)
	requireAppErr(t, appErr, "office.save_failed", http.StatusBadGateway)
}

func useOfficeSaveLimit(t *testing.T, limit int) {
	t.Helper()

	previous := maxOfficeSaveSize
	maxOfficeSaveSize = limit
	t.Cleanup(func() { maxOfficeSaveSize = previous })
}

func TestReadOfficeSave(t *testing.T) {
	useOfficeSaveLimit(t, 10)

	content, err := readOfficeSave(strings.NewReader("0123456789"))
	if err != nil || string(content) != "0123456789" {
		t.Fatalf("a save at the limit should be read, got %q, %v", content, err)
	}

	_, err = readOfficeSave(strings.NewReader("0123456789X"))
	if !errors.Is(err, errOfficeSaveTooLarge) {
		t.Fatalf("a save over the limit should be refused, got %v", err)
	}
}

func TestOfficeSaveFile_RefusesSaveOverLimit(t *testing.T) {
	useOfficeSaveLimit(t, 10)
	a, stores := newOfficeAccessApp()
	stores.files.shares["user-1/doc-1"] = []model.File{{AccessLevel: model.AccessEditor}}

	body := io.NopCloser(strings.NewReader("this is more than ten bytes"))

	appErr := a.OfficeSaveFile("doc-1", userOfficeToken(t, "files", "doc-1"), body)
	requireAppErr(t, appErr, "office.save_failed", http.StatusRequestEntityTooLarge)
}

func TestEuroOfficeCallback_RefusesDownloadOverLimit(t *testing.T) {
	useOfficeSaveLimit(t, 10)
	a, stores := newOfficeAccessApp()
	stores.files.shares["user-1/doc-1"] = []model.File{{AccessLevel: model.AccessEditor}}

	large := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("this is more than ten bytes"))
	}))
	defer large.Close()

	signed := signOfficeTestToken(t, jwt.MapClaims{"status": 2, "url": large.URL + "/output.docx"})
	body := io.NopCloser(strings.NewReader(`{"token":"` + signed + `"}`))

	appErr := a.EuroOfficeCallback("doc-1", "files", userOfficeToken(t, "files", "doc-1"), "", body)
	requireAppErr(t, appErr, "office.save_failed", http.StatusRequestEntityTooLarge)
}

func fillOwnerStorage(a *App, stores officeTestStores) {
	owner := a.Store.User.(*fakePublicUserStore).user
	owner.StorageLimit = 1000
	stores.files.storageUsed = 1000
}

func TestOfficeSaveFile_RefusesEditorSaveWhenOwnerStorageIsFull(t *testing.T) {
	a, stores := newOfficeAccessApp()
	stores.files.shares["user-1/doc-1"] = []model.File{{AccessLevel: model.AccessEditor}}
	fillOwnerStorage(a, stores)

	body := io.NopCloser(strings.NewReader("a larger document than before"))

	appErr := a.OfficeSaveFile("doc-1", userOfficeToken(t, "files", "doc-1"), body)
	requireAppErr(t, appErr, "office.save_failed", http.StatusInsufficientStorage)
}

func TestEuroOfficeCallback_RefusesEditorSaveWhenOwnerStorageIsFull(t *testing.T) {
	a, stores := newOfficeAccessApp()
	stores.files.shares["user-1/doc-1"] = []model.File{{AccessLevel: model.AccessEditor}}
	fillOwnerStorage(a, stores)

	saved := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("a larger document than before"))
	}))
	defer saved.Close()

	signed := signOfficeTestToken(t, jwt.MapClaims{"status": 2, "url": saved.URL + "/output.docx"})
	body := io.NopCloser(strings.NewReader(`{"token":"` + signed + `"}`))

	appErr := a.EuroOfficeCallback("doc-1", "files", userOfficeToken(t, "files", "doc-1"), "", body)
	requireAppErr(t, appErr, "office.save_failed", http.StatusInsufficientStorage)
}

func TestCheckOwnerSpaceForEdit_AllowsShrinkingSaveWhenStorageIsFull(t *testing.T) {
	a, stores := newOfficeAccessApp()
	fillOwnerStorage(a, stores)

	if appErr := a.checkOwnerSpaceForEdit("owner-1", -10, false); appErr != nil {
		t.Fatalf("a save that makes the document smaller should be allowed, got %v", appErr)
	}
}

func TestCheckOwnerSpaceForEdit_HidesStorageStateFromGuests(t *testing.T) {
	a, stores := newOfficeAccessApp()
	fillOwnerStorage(a, stores)

	editorErr := a.checkOwnerSpaceForEdit("owner-1", 10, false)
	requireAppErr(t, editorErr, "office.save_failed", http.StatusInsufficientStorage)

	guestErr := a.checkOwnerSpaceForEdit("owner-1", 10, true)
	requireAppErr(t, guestErr, "office.save_failed", http.StatusInternalServerError)
}
