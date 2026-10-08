// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/twigex/twigex/model"
)

func TestDueForRun(t *testing.T) {
	now := time.Unix(10000, 0)
	interval := time.Hour

	cases := []struct {
		name   string
		latest *model.Job
		want   bool
	}{
		{
			name:   "never run",
			latest: nil,
			want:   true,
		},
		{
			name:   "one already pending",
			latest: &model.Job{Status: model.JobStatusPending, CreatedAt: 0},
			want:   false,
		},
		{
			name:   "one already running",
			latest: &model.Job{Status: model.JobStatusRunning, CreatedAt: 0},
			want:   false,
		},
		{
			name:   "last run inside the interval",
			latest: &model.Job{Status: model.JobStatusCompleted, CreatedAt: now.Unix() - 60},
			want:   false,
		},
		{
			name:   "last run older than the interval",
			latest: &model.Job{Status: model.JobStatusCompleted, CreatedAt: now.Unix() - 3601},
			want:   true,
		},
		{
			name:   "exactly at the interval",
			latest: &model.Job{Status: model.JobStatusCompleted, CreatedAt: now.Unix() - 3600},
			want:   true,
		},
		// A refusal is finished, so the schedule carries on rather than stalling
		// on a directory that was briefly unreadable.
		{
			name:   "last run refused and old enough",
			latest: &model.Job{Status: model.JobStatusRefused, CreatedAt: now.Unix() - 3601},
			want:   true,
		},
		{
			name:   "last run failed and old enough",
			latest: &model.Job{Status: model.JobStatusFailed, CreatedAt: now.Unix() - 3601},
			want:   true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := dueForRun(c.latest, now, interval); got != c.want {
				t.Errorf("dueForRun = %v, want %v", got, c.want)
			}
		})
	}
}

type fakeJob struct {
	jobType string
}

func (f fakeJob) Type() string                          { return f.jobType }
func (f fakeJob) Run(context.Context, *model.Job) error { return nil }

type fakeScheduledJob struct {
	fakeJob
	interval time.Duration
	enabled  bool
}

func (f fakeScheduledJob) Schedule() (time.Duration, bool) {
	return f.interval, f.enabled
}

// Only a job that opts in through Scheduled is ever enqueued on a timer, and
// the registry has to keep dispatching the ones that do not.
func TestRegistryDistinguishesScheduledJobs(t *testing.T) {
	onDemand := fakeJob{jobType: "on_demand"}
	scheduled := fakeScheduledJob{fakeJob: fakeJob{jobType: "scheduled"}, interval: time.Hour, enabled: true}

	r := NewRegistry(nil, onDemand, scheduled)

	if len(r.jobs) != 2 {
		t.Fatalf("registered %d jobs, want 2", len(r.jobs))
	}

	if _, ok := r.jobs["on_demand"].(Scheduled); ok {
		t.Error("an on demand job must not satisfy Scheduled")
	}

	got, ok := r.jobs["scheduled"].(Scheduled)
	if !ok {
		t.Fatal("a scheduled job must satisfy Scheduled")
	}

	interval, enabled := got.Schedule()
	if interval != time.Hour || !enabled {
		t.Errorf("Schedule = %v %v, want 1h true", interval, enabled)
	}
}
