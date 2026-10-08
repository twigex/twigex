// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

// Tests for the untyped claim and for retention, which keeps the newest
// unsuccessful run of a type so an overnight refusal survives cleanup.
// Runs against real MySQL (see testhelper_test.go).

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/twigex/twigex/model"
)

var jobSeedSeq int

func seedJob(t *testing.T, repo *jobsRepository, jobType, status, userID string, createdAt, finishedAt int64) string {
	t.Helper()

	jobSeedSeq++
	id := fmt.Sprintf("job-%06d", jobSeedSeq)

	job, err := repo.Create(model.Job{
		ID:         id,
		Type:       jobType,
		Status:     status,
		UserID:     userID,
		Payload:    json.RawMessage("{}"),
		CreatedAt:  createdAt,
		FinishedAt: finishedAt,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	return job.ID
}

func TestClaimNextPending_ClaimsAnyTypeAndMarksRunning(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "jobs")

	repo := &jobsRepository{Db: db}
	first := seedJob(t, repo, model.JobTypeLDAPSync, model.JobStatusPending, "", 100, 0)
	seedJob(t, repo, model.JobTypeMoveFiles, model.JobStatusPending, "user-1", 200, 0)

	claimed, err := repo.ClaimNextPending()
	if err != nil {
		t.Fatalf("ClaimNextPending: %v", err)
	}
	if claimed == nil {
		t.Fatal("ClaimNextPending returned nothing with two pending jobs")
	}

	if claimed.ID != first {
		t.Errorf("claimed %s, want the oldest pending job %s", claimed.ID, first)
	}
	if claimed.Status != model.JobStatusRunning {
		t.Errorf("claimed job status is %q, want %q", claimed.Status, model.JobStatusRunning)
	}

	stored, err := repo.Get(claimed.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if stored.Status != model.JobStatusRunning {
		t.Errorf("stored status is %q, want the claim to have persisted %q", stored.Status, model.JobStatusRunning)
	}
}

// A system job stores NULL in user_id, which the scan has to survive.
func TestClaimNextPending_ReadsSystemJobWithNoUser(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "jobs")

	repo := &jobsRepository{Db: db}
	id := seedJob(t, repo, model.JobTypeLDAPSync, model.JobStatusPending, "", 100, 0)

	var userID *string
	if err := db.QueryRow("SELECT user_id FROM jobs WHERE id = ?", id).Scan(&userID); err != nil {
		t.Fatalf("reading user_id: %v", err)
	}
	if userID != nil {
		t.Errorf("user_id is %q, want NULL for a job with no user", *userID)
	}

	claimed, err := repo.ClaimNextPending()
	if err != nil {
		t.Fatalf("ClaimNextPending: %v", err)
	}
	if claimed == nil || claimed.UserID != "" {
		t.Fatalf("a system job did not read back with an empty user, got %+v", claimed)
	}
}

func TestClaimNextPending_IgnoresRunningAndFinished(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "jobs")

	repo := &jobsRepository{Db: db}
	seedJob(t, repo, model.JobTypeLDAPSync, model.JobStatusRunning, "", 100, 0)
	seedJob(t, repo, model.JobTypeMoveFiles, model.JobStatusCompleted, "user-1", 200, 300)

	claimed, err := repo.ClaimNextPending()
	if err != nil {
		t.Fatalf("ClaimNextPending: %v", err)
	}
	if claimed != nil {
		t.Errorf("claimed %s (%s) with nothing pending", claimed.ID, claimed.Status)
	}
}

func TestCountActiveJobsByType(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "jobs")

	repo := &jobsRepository{Db: db}
	seedJob(t, repo, model.JobTypeLDAPSync, model.JobStatusPending, "", 100, 0)
	seedJob(t, repo, model.JobTypeLDAPSync, model.JobStatusRunning, "", 200, 0)
	seedJob(t, repo, model.JobTypeLDAPSync, model.JobStatusCompleted, "", 300, 400)
	seedJob(t, repo, model.JobTypeMoveFiles, model.JobStatusPending, "user-1", 500, 0)

	count, err := repo.CountActiveByType(model.JobTypeLDAPSync)
	if err != nil {
		t.Fatalf("CountActiveByType: %v", err)
	}
	if count != 2 {
		t.Errorf("counted %d active sync jobs, want 2 (the pending and the running one)", count)
	}
}

func TestGetLatestJobByType(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "jobs")

	repo := &jobsRepository{Db: db}
	seedJob(t, repo, model.JobTypeLDAPSync, model.JobStatusCompleted, "", 100, 150)
	newest := seedJob(t, repo, model.JobTypeLDAPSync, model.JobStatusRefused, "", 200, 250)
	seedJob(t, repo, model.JobTypeMoveFiles, model.JobStatusCompleted, "user-1", 300, 350)

	latest, err := repo.GetLatestByType(model.JobTypeLDAPSync)
	if err != nil {
		t.Fatalf("GetLatestByType: %v", err)
	}
	if latest == nil || latest.ID != newest {
		t.Fatalf("got %+v, want the newest sync job %s", latest, newest)
	}
}

func TestGetLatestJobByType_NoRows(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "jobs")

	repo := &jobsRepository{Db: db}

	latest, err := repo.GetLatestByType(model.JobTypeLDAPSync)
	if err != nil {
		t.Fatalf("GetLatestByType: %v", err)
	}
	if latest != nil {
		t.Errorf("got %+v, want nil when the type has never run", latest)
	}
}

// The founding requirement: a refusal at 3am must still be readable later.
func TestCleanupOldJobs_KeepsNewestUnsuccessfulRunPerType(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "jobs")

	repo := &jobsRepository{Db: db}
	oldSuccess := seedJob(t, repo, model.JobTypeLDAPSync, model.JobStatusCompleted, "", 100, 100)
	refusal := seedJob(t, repo, model.JobTypeLDAPSync, model.JobStatusRefused, "", 200, 200)
	moveFailure := seedJob(t, repo, model.JobTypeMoveFiles, model.JobStatusFailed, "user-1", 300, 300)

	if err := repo.DeleteOld(1000); err != nil {
		t.Fatalf("DeleteOld: %v", err)
	}

	assertJobGone(t, repo, oldSuccess, "a superseded successful run")
	assertJobKept(t, repo, refusal, "the newest refusal of its type")
	assertJobKept(t, repo, moveFailure, "the newest failure of another type")
}

func TestCleanupOldJobs_LaterSuccessSupersedesAFailure(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "jobs")

	repo := &jobsRepository{Db: db}
	refusal := seedJob(t, repo, model.JobTypeLDAPSync, model.JobStatusRefused, "", 100, 100)
	success := seedJob(t, repo, model.JobTypeLDAPSync, model.JobStatusCompleted, "", 200, 200)

	if err := repo.DeleteOld(1000); err != nil {
		t.Fatalf("DeleteOld: %v", err)
	}

	assertJobGone(t, repo, refusal, "a refusal followed by a success")
	assertJobGone(t, repo, success, "the superseding success, being old itself")
}

func TestCleanupOldJobs_OnlyTheNewestFailureIsKept(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "jobs")

	repo := &jobsRepository{Db: db}
	older := seedJob(t, repo, model.JobTypeLDAPSync, model.JobStatusFailed, "", 100, 100)
	newer := seedJob(t, repo, model.JobTypeLDAPSync, model.JobStatusFailed, "", 200, 200)

	if err := repo.DeleteOld(1000); err != nil {
		t.Fatalf("DeleteOld: %v", err)
	}

	assertJobGone(t, repo, older, "an older failure of the same type")
	assertJobKept(t, repo, newer, "the newest failure of its type")
}

func TestCleanupOldJobs_LeavesUnfinishedAndRecentRows(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "jobs")

	repo := &jobsRepository{Db: db}
	pending := seedJob(t, repo, model.JobTypeLDAPSync, model.JobStatusPending, "", 100, 0)
	running := seedJob(t, repo, model.JobTypeMoveFiles, model.JobStatusRunning, "user-1", 100, 0)
	recent := seedJob(t, repo, model.JobTypeMoveFiles, model.JobStatusCompleted, "user-1", 100, 5000)

	if err := repo.DeleteOld(1000); err != nil {
		t.Fatalf("DeleteOld: %v", err)
	}

	assertJobKept(t, repo, pending, "a pending job")
	assertJobKept(t, repo, running, "a running job")
	assertJobKept(t, repo, recent, "a job that finished after the cutoff")
}

func assertJobKept(t *testing.T, repo *jobsRepository, id, what string) {
	t.Helper()

	job, err := repo.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if job == nil {
		t.Errorf("cleanup removed %s (%s)", what, id)
	}
}

func assertJobGone(t *testing.T, repo *jobsRepository, id, what string) {
	t.Helper()

	job, err := repo.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if job != nil {
		t.Errorf("cleanup kept %s (%s)", what, id)
	}
}
