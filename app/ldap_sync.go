// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"net/http"
	"time"

	"github.com/twigex/twigex/app/jobs"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

func (a *App) LDAPSyncSettings() model.LDAPSyncSettings {
	ldap := a.ConfigStore.Config.LDAPSettings

	enabled := *ldap.Enabled &&
		*ldap.SyncEnabled &&
		a.LDAPAuth != nil &&
		a.Server.License.HasLDAP()

	return model.LDAPSyncSettings{
		Enabled:              enabled,
		Interval:             time.Duration(*ldap.SyncIntervalMinutes) * time.Minute,
		MaxDeactivatePercent: *ldap.SyncMaxDeactivatePercent,
	}
}

func (a *App) RunLDAPSync(user model.User) (*model.Job, *model.AppError) {
	if user.Role != model.SystemAdminRoleId {
		return nil, model.NewAppError("session.not_authorized", http.StatusForbidden)
	}

	if a.LDAPAuth == nil {
		return nil, model.NewAppError("auth.ldap_unavailable", http.StatusNotImplemented)
	}

	if !*a.ConfigStore.Config.LDAPSettings.Enabled || !a.Server.License.HasLDAP() {
		return nil, model.NewAppError("auth.ldap_unavailable", http.StatusNotImplemented)
	}

	active, err := a.Store.Jobs.CountActiveByType(model.JobTypeLDAPSync)
	if err != nil {
		tlog.Errorw("Could not check for a running LDAP sync", "error", err)
		return nil, model.NewAppError("ldap.sync_enqueue_failed", http.StatusInternalServerError)
	}

	if active > 0 {
		return nil, model.NewAppError("ldap.sync_already_running", http.StatusConflict)
	}

	job, err := jobs.Enqueue(a.Store.Jobs, model.JobTypeLDAPSync)
	if err != nil {
		tlog.Errorw("Could not enqueue an LDAP sync", "user_id", user.ID, "error", err)
		return nil, model.NewAppError("ldap.sync_enqueue_failed", http.StatusInternalServerError)
	}

	tlog.Infow("An administrator requested an LDAP sync", "user_id", user.ID, "job_id", job.ID)

	return job, nil
}

// A sync that has not succeeded for more than twice its interval is reported as
// stale, so a schedule that silently stopped is as visible as one that failed.
func (a *App) LDAPSyncStatus() (*model.LDAPSyncStatus, error) {
	settings := a.LDAPSyncSettings()
	if !settings.Enabled {
		return nil, nil
	}

	job, err := a.Store.Jobs.GetLatestByType(model.JobTypeLDAPSync)
	if err != nil {
		return nil, err
	}

	if job == nil {
		return &model.LDAPSyncStatus{Status: model.LDAPSyncNeverRun}, nil
	}

	payload, err := job.LDAPSyncPayload()
	if err != nil {
		return nil, err
	}

	reason := ""
	if job.Error != nil {
		reason = *job.Error
	}

	status := &model.LDAPSyncStatus{
		Status:      job.Status,
		Reason:      reason,
		Deactivated: payload.Deactivated,
		Reactivated: payload.Reactivated,
		Updated:     payload.Updated,
		Failed:      payload.Failed,
		StartedAt:   job.StartedAt,
		FinishedAt:  job.FinishedAt,
	}

	if job.Status == model.JobStatusCompleted && settings.Interval > 0 {
		age := time.Since(time.Unix(job.FinishedAt, 0))
		status.Stale = age > 2*settings.Interval
	}

	return status, nil
}
