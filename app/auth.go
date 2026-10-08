// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/avct/uasurfer"
	"github.com/google/uuid"
	"github.com/twigex/twigex/crypto"
	"github.com/twigex/twigex/internal/mfa"
	"github.com/twigex/twigex/internal/validate"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

func (a *App) ValidateSession(sid string) (*model.SessionDetails, *model.AppError) {
	sd := model.SessionDetails{}

	if err := a.Store.Auth.Get(sid, &sd); err != nil {
		tlog.Errorw("Failed to retrieve session", "session_id", sid, "error", err)
		return nil, model.NewAppError("session.invalid", http.StatusUnauthorized)
	}

	if sd.IsExpired() {
		return nil, model.NewAppError("session.expired", http.StatusUnauthorized)
	}

	user, err := a.Store.User.Get(sd.UserID)
	if err != nil {
		tlog.Errorw("User lookup failed", "user_id", sd.UserID, "error", err)
		return nil, model.NewAppError("user.not_found", http.StatusInternalServerError)
	}

	if user == nil {
		tlog.Warnw("Session references non-existent user",
			"user_id", sd.UserID,
		)
		return nil, model.NewAppError("user.not_found", http.StatusUnauthorized)
	}

	if user.DeactivatedAt != 0 {
		tlog.Warnw("Session belongs to a deactivated account",
			"user_id", sd.UserID,
		)
		return nil, model.NewAppError("auth.user_deactivated", http.StatusUnauthorized)
	}

	if !sd.IsAuthorized() {
		return nil, model.NewAppError("session.not_authorized", http.StatusUnauthorized)
	}

	return &sd, nil
}

func (a *App) UpdateSession(sd *model.SessionDetails, w http.ResponseWriter) {
	a.ExtendSessionExpiry(sd, w)
	a.UpdateLastActivity(sd)
}

func (a *App) InvalidateClientSession(w http.ResponseWriter) {
	a.DeleteAuthCookies(w)
}

func (a *App) secureCookies() bool {
	return strings.HasPrefix(*a.ConfigStore.Config.ServerSettings.SiteURL, "https://")
}

func (a *App) setAuthCookie(w http.ResponseWriter, name, value string, expires time.Time, httpOnly bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Expires:  expires,
		HttpOnly: httpOnly,
		Secure:   a.secureCookies(),
		SameSite: http.SameSiteLaxMode,
	})
}

func (a *App) ValidateCredentials(c model.Credentials, clientIP string) (*model.User, *model.AppError) {
	loginID := strings.ToLower(c.LoginID)

	limiter := a.loginLimiter()

	allowed, err := limiter.Allow(clientIP, loginID)
	if err != nil {
		tlog.Errorw("Failed to check login rate limit", "error", err)
		return nil, model.NewAppError("auth.invalid_credentials", http.StatusInternalServerError)
	}

	if !allowed {
		tlog.Warnw("Login refused by rate limit", "client_ip", clientIP)
		return nil, model.NewAppError("auth.rate_limited", http.StatusTooManyRequests)
	}

	user, appErr := a.authenticate(c, loginID)
	if appErr != nil {
		return nil, appErr
	}

	if err := limiter.Reset(clientIP, loginID); err != nil {
		tlog.Warnw("Failed to clear login attempts", "error", err)
	}

	return user, nil
}

func (a *App) authenticate(c model.Credentials, loginID string) (*model.User, *model.AppError) {
	user, appErr := a.findUserByLoginID(loginID)
	if appErr != nil {
		return nil, appErr
	}

	if user != nil && user.DeactivatedAt != 0 {
		tlog.Warnw("Login refused for deactivated account",
			"user_id", user.ID,
			"auth_service", user.AuthService,
		)
		return nil, model.NewAppError("auth.user_deactivated", http.StatusUnauthorized)
	}

	if (user == nil || user.AuthService == "ldap") && *a.ConfigStore.Config.LDAPSettings.Enabled &&
		a.Server.License.HasLDAP() {
		return a.validateLDAPCredentials(c, user)
	}

	if user == nil {
		return nil, model.NewAppError("auth.user_deactivated", http.StatusUnauthorized)
	}

	return a.validatePasswordCredentials(c, user)
}

func (a *App) findUserByLoginID(loginID string) (*model.User, *model.AppError) {
	if validate.IsEmail(loginID) {
		user, err := a.Store.User.GetByEmail(loginID)
		if err != nil {
			tlog.Errorw("Failed to retrieve user by email", "error", err)
			return nil, model.NewAppError("auth.invalid_credentials", http.StatusInternalServerError)
		}

		return user, nil
	}

	user, err := a.Store.User.GetByUsername(loginID)
	if err != nil {
		tlog.Errorw("Failed to retrieve user by username", "error", err)
		return nil, model.NewAppError("auth.invalid_credentials", http.StatusInternalServerError)
	}

	return user, nil
}

func (a *App) validateLDAPCredentials(c model.Credentials, user *model.User) (*model.User, *model.AppError) {
	appErr := verifyMFAToken(c.Token, user)
	if appErr != nil {
		return nil, appErr
	}

	if a.LDAPAuth == nil {
		return nil, model.NewAppError("auth.ldap_unavailable", http.StatusNotImplemented)
	}

	ldapUser, err := a.LDAPAuth.Authorize(c, user)
	if err != nil {
		tlog.Errorw("Failed to authorize LDAP user", "error", err)
		return nil, model.NewAppError("auth.ldap_failed", http.StatusUnauthorized)
	}

	return ldapUser, nil
}

func (a *App) validatePasswordCredentials(c model.Credentials, user *model.User) (*model.User, *model.AppError) {
	if !crypto.PasswordMatches(c.Password, user.Password) {
		return nil, model.NewAppError("auth.invalid_credentials", http.StatusUnauthorized)
	}

	appErr := verifyMFAToken(c.Token, user)
	if appErr != nil {
		return nil, appErr
	}

	return user, nil
}

func verifyMFAToken(token string, user *model.User) *model.AppError {
	// A first LDAP login has no local account yet, so there is no MFA secret to
	// check. Existing users always resolve here non-nil and are still enforced.
	if user == nil || !user.MfaActive {
		return nil
	}

	if token == "" {
		return model.NewAppErrorRaw("provide_mfa_token", http.StatusUnauthorized)
	}

	ok, err := mfa.ValidateToken(user.MfaSecret, token)
	if err != nil {
		tlog.Errorw("Failed to validate MFA token",
			"user_id", user.ID,
			"error", err,
		)
		return model.NewAppError("auth.mfa_validation_failed", http.StatusInternalServerError)
	}

	if !ok {
		return model.NewAppError("auth.mfa_token_invalid", http.StatusUnauthorized)
	}

	return nil
}

func (a *App) ClientIP(r *http.Request) string {
	s := a.ConfigStore.Config.ServerSettings
	return getIPAddress(r, s.TrustedProxyHeaders, s.TrustedProxies)
}

func (a *App) LogIn(u model.User, isOauth bool, w http.ResponseWriter, r *http.Request) *model.AppError {
	if u.DeactivatedAt != 0 {
		tlog.Warnw("Login refused for deactivated account",
			"user_id", u.ID,
			"is_oauth", isOauth,
		)
		return model.NewAppError("auth.user_deactivated", http.StatusUnauthorized)
	}

	str := r.UserAgent()
	ua := uasurfer.Parse(str)

	ip := a.ClientIP(r)

	expiry, err := a.GetSessionLength()
	if err != nil {
		tlog.Errorw("Failed to log in", "error", err)
		return model.NewAppError("auth.login_failed", http.StatusInternalServerError)
	}

	t := time.Now()
	csrfToken := model.NewID()

	sd := model.SessionDetails{
		AccessToken:  uuid.New().String(),
		UserID:       u.ID,
		IP:           ip,
		DeviceID:     getOSName(ua),
		IsOauth:      isOauth,
		Authorized:   true,
		Browser:      getBrowserName(ua),
		CSRF:         csrfToken,
		Created:      t.Unix(),
		LastActivity: t.Unix(),
		Expires:      t.AddDate(0, 0, expiry).Unix(),
	}

	if err = a.Store.Auth.Set(sd.AccessToken, time.Hour*24*time.Duration(expiry), sd); err != nil {
		tlog.Errorw("Failed to set session", "error", err)
		return model.NewAppError("auth.login_failed", http.StatusInternalServerError)
	}

	expires := time.Now().AddDate(0, 0, expiry)
	a.setAuthCookie(w, model.SessionCookieToken, sd.AccessToken, expires, true)
	a.setAuthCookie(w, model.SessionCookieCsrf, csrfToken, expires, false)
	a.setAuthCookie(w, model.SessionCookieLoggedIn, "1", expires, false)

	return nil
}

func (a *App) DeleteAuthCookies(w http.ResponseWriter) {
	expired := time.Unix(0, 0)
	a.setAuthCookie(w, model.SessionCookieCsrf, "", expired, false)
	a.setAuthCookie(w, model.SessionCookieToken, "", expired, true)
	a.setAuthCookie(w, model.SessionCookieLoggedIn, "", expired, false)
}

func (a *App) LogOut(sid string, w http.ResponseWriter) *model.AppError {
	err := a.Store.Auth.Delete(sid)
	if err != nil {
		tlog.Errorw("Failed to delete session", "session_id", sid, "error", err)
		return model.NewAppError("session.delete_failed", http.StatusInternalServerError)
	}

	a.DeleteAuthCookies(w)

	return nil
}

func (a *App) UpdateLastActivity(session *model.SessionDetails) bool {
	if session == nil || session.IsExpired() {
		return false
	}

	session.LastActivity = time.Now().Unix()

	err := a.Store.Auth.UpdateLastActivity(*session)
	if err != nil {
		tlog.Errorw("Failed to update sessions last activity", "error", err)
		return false
	}

	return true
}

func (a *App) ExtendSessionExpiry(session *model.SessionDetails, w http.ResponseWriter) bool {
	if session == nil || session.IsExpired() {
		return false
	}

	day := 24 * 60 * 60
	currentTime := time.Now().Unix()
	sessionLength, err := a.GetSessionLength()
	if err != nil {
		tlog.Errorw("Failed to get session length", "error", err)
		return false
	}

	elapsed := currentTime - (time.Unix(session.Expires, 0).Unix() - int64((24*60*60)*sessionLength)) + 60*60
	if elapsed < int64(day) {
		return false
	}

	session.Expires += elapsed

	if err := a.Store.Auth.UpdateSession(*session, sessionLength); err != nil {
		tlog.Errorw("Failed to update session",
			"user_id", session.UserID,
			"error", err,
		)
		return false
	}

	expires := time.Now().AddDate(0, 0, sessionLength)
	a.setAuthCookie(w, model.SessionCookieToken, session.AccessToken, expires, true)
	a.setAuthCookie(w, model.SessionCookieCsrf, session.CSRF, expires, false)

	return true
}

func (a *App) GetSessionLength() (int, error) {
	e := a.ConfigStore.Config.ServerSettings.SessionLengthInDays

	expiry, err := strconv.Atoi(*e)
	if err != nil {
		return 0, err
	}

	return expiry, nil
}

func (a *App) GetOIDCByID(id string) (*model.OIDCProvider, *model.AppError) {
	provider, err := a.Store.OIDC.GetByID(id)
	if err != nil {
		tlog.Errorw("Failed to get OIDC by id", "id", id, "error", err)
		return nil, model.NewAppError("auth.oauth_unavailable", http.StatusInternalServerError)
	}

	if !provider.Enabled {
		return nil, model.NewAppError("auth.oauth_provider_disabled", http.StatusForbidden)
	}

	return provider, nil
}

func (a *App) OIDCAuthCodeURL(ctx context.Context, providerID, state, nonce string) (string, *model.AppError) {
	if a.OIDCAuth == nil {
		return "", model.NewAppError("auth.oauth_unavailable", http.StatusNotImplemented)
	}

	return a.OIDCAuth.AuthCodeURL(ctx, providerID, state, nonce)
}

func (a *App) ExchangeOauth2Token(ctx context.Context, providerID, code, nonce string) (*model.User, *model.AppError) {
	if a.OIDCAuth == nil {
		return nil, model.NewAppError("auth.oauth_unavailable", http.StatusNotImplemented)
	}

	return a.OIDCAuth.ExchangeToken(ctx, providerID, code, nonce)
}
