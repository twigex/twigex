// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/twigex/twigex/internal/licensing"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

func (a *App) AddLicense(user model.User, licenseBytes []byte) (*model.License, *model.AppError) {
	if !a.SessionHasPermission(user, model.FilePermissions.PermissionChangeLicense) {
		return nil, model.NewAppError("session.not_authorized", http.StatusForbidden)
	}

	if a.ConfigStore.EnvLocked.License {
		return nil, model.NewAppError("settings.env_locked", http.StatusForbidden)
	}

	licenseString := strings.TrimSpace(string(licenseBytes))

	newLicense, err := licensing.VerifyLicense(licenseString)
	if err != nil {
		tlog.Errorw("Failed to verify license",
			"error", err,
		)
		return nil, model.NewAppError("license.verify_failed", http.StatusBadRequest)
	}

	if newLicense.ExpiresAt < time.Now().Unix() {
		tlog.Warnw("Rejected an expired license",
			"expires_at", newLicense.ExpiresAt,
		)
		return nil, model.NewAppError("license.expired", http.StatusBadRequest)
	}

	// Seat checks otherwise only run on user creation, so removing the license,
	// adding users and re-applying it would land the instance over its limit.
	// Compared against the hard limit so a customer inside their grace band can
	// still apply their own renewal.
	newLicense.SetDefaults()

	active, appErr := a.ActiveUserCount(context.Background())
	if appErr != nil {
		return nil, appErr
	}

	if newLicense.TooManyUsers(active) {
		tlog.Warnw("Rejected a license covering fewer seats than the instance uses",
			"licensed_seats", *newLicense.Features.Users,
			"active_users", active,
		)
		return nil, model.NewAppError("license.seats_exceeded", http.StatusBadRequest)
	}

	if err = a.Store.License.Create(licenseString); err != nil {
		tlog.Errorw("Failed to save license",
			"error", err,
		)
		return nil, model.NewAppError("license.save_failed", http.StatusInternalServerError)
	}

	if err = a.LoadLicense(); err != nil {
		tlog.Errorw("Failed to reload licenses after saving",
			"error", err,
		)
		return nil, model.NewAppError("license.save_failed", http.StatusInternalServerError)
	}

	return newLicense, nil
}

func (a *App) RemoveActiveLicense(user model.User) *model.AppError {
	if !a.SessionHasPermission(user, model.FilePermissions.PermissionChangeLicense) {
		return model.NewAppError("session.not_authorized", http.StatusForbidden)
	}

	if a.ConfigStore.EnvLocked.License {
		return model.NewAppError("settings.env_locked", http.StatusForbidden)
	}

	if err := a.Store.License.DeleteActive(); err != nil {
		tlog.Errorw("Failed to remove active license",
			"error", err,
		)
		return model.NewAppError("license.remove_failed", http.StatusInternalServerError)
	}

	if err := a.LoadLicense(); err != nil {
		tlog.Errorw("Failed to reload licenses after removal",
			"error", err,
		)
		return model.NewAppError("license.remove_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) GetLicenseMap(licenseInfo bool) *map[string]interface{} {
	m := make(map[string]any, 0)
	m["isLicensed"] = false
	m["version"] = model.Version

	license := a.Server.License.ActiveLicense()

	if licenseInfo {
		if queued := a.Server.License.Queued(); queued != nil {
			m["renewal_starts_at"] = queued.StartsAt
		}
	}

	if license != nil {
		if licenseInfo {
			m["name"] = license.Customer.Name
			m["edition"] = license.SkuName
			m["expires_at"] = license.ExpiresAt
			m["starts_at"] = license.StartsAt
			m["issued_at"] = license.IssuedAt
		}

		m["isLicensed"] = true

		featureMap := license.ToMap()
		for k, v := range featureMap {
			m[k] = v
		}

		if licenseInfo {
			m["deployment"] = license.Deployment
			m["seat_limit"] = license.SeatLimit()
			m["seat_hard_limit"] = license.SeatHardLimit()

			// Inside the grace band the instance still works, so the admin only
			// learns they are over their licence if we say so.
			if over, appErr := a.SeatsOverSoftLimit(context.Background()); appErr == nil {
				m["seats_over_limit"] = over
			}

			if count, appErr := a.ActiveUserCount(context.Background()); appErr == nil {
				m["active_users"] = count
			}
		}
	}

	// Overrides the licensed value with what this build can actually do, so the UI
	// hides a role editor whose saves would be rejected.
	m["collimato_roles"] = a.CollimatoRoles != nil && a.Server.License.HasCollimatoRoles()

	return &m
}
