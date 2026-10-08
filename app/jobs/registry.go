// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package jobs

import (
	"context"
	"encoding/json"
	"time"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
	"github.com/twigex/twigex/tlog"
)

const scheduleTick = time.Minute

type Job interface {
	Type() string
	Run(ctx context.Context, job *model.Job) error
}

// Scheduled is implemented by a job that should also be enqueued on an
// interval. It is asked on every tick rather than at registration, so a change
// to configuration or to the licence applies without a restart.
type Scheduled interface {
	Schedule() (time.Duration, bool)
}

type Registry struct {
	store store.JobsStore
	jobs  map[string]Job
}

func NewRegistry(store store.JobsStore, jobs ...Job) *Registry {
	r := &Registry{store: store, jobs: make(map[string]Job, len(jobs))}

	for _, j := range jobs {
		r.jobs[j.Type()] = j
	}

	return r
}

func (r *Registry) Start(ctx context.Context) {
	types := make([]string, 0, len(r.jobs))
	for t := range r.jobs {
		types = append(types, t)
	}

	tlog.Infow("jobs starting", "job_types", types)

	go r.claim(ctx)
	go r.schedule(ctx)
}

func (r *Registry) claim(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			tlog.Infow("job claiming stopping")
			return
		}

		job, err := r.store.ClaimNextPending()
		if err != nil {
			tlog.Errorw("job fetch failed", "error", err.Error())
			time.Sleep(time.Second)
			continue
		}

		if job == nil {
			time.Sleep(time.Second)
			continue
		}

		impl, ok := r.jobs[job.Type]
		if !ok {
			tlog.Errorw("no job registered for type", "job_id", job.ID, "job_type", job.Type)
			_ = r.store.UpdateStatus(job.ID, model.JobStatusFailed, "no job registered for type "+job.Type)
			continue
		}

		tlog.Infow("job picked", "job_id", job.ID, "job_type", job.Type)

		go r.run(ctx, job, impl)

		time.Sleep(100 * time.Millisecond)
	}
}

func (r *Registry) schedule(ctx context.Context) {
	ticker := time.NewTicker(scheduleTick)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			tlog.Infow("job scheduling stopping")
			return
		case now := <-ticker.C:
			for _, j := range r.jobs {
				scheduled, ok := j.(Scheduled)
				if !ok {
					continue
				}

				interval, enabled := scheduled.Schedule()
				if !enabled {
					continue
				}

				r.enqueueIfDue(j.Type(), now, interval)
			}
		}
	}
}

// Due-ness comes from the last row rather than an in-memory timestamp, so a
// restart does not trigger a run and two instances agree on when one is owed.
func (r *Registry) enqueueIfDue(jobType string, now time.Time, interval time.Duration) {
	latest, err := r.store.GetLatestByType(jobType)
	if err != nil {
		tlog.Errorw("could not read the last run", "job_type", jobType, "error", err)
		return
	}

	if !dueForRun(latest, now, interval) {
		return
	}

	if _, err := Enqueue(r.store, jobType); err != nil {
		tlog.Errorw("could not enqueue a scheduled run", "job_type", jobType, "error", err)
		return
	}

	tlog.Infow("enqueued a scheduled run", "job_type", jobType)
}

func dueForRun(latest *model.Job, now time.Time, interval time.Duration) bool {
	if latest == nil {
		return true
	}

	// One already queued or in flight, so a second would duplicate the work.
	if !model.IsJobFinished(latest.Status) {
		return false
	}

	return now.Sub(time.Unix(latest.CreatedAt, 0)) >= interval
}

// Enqueue adds work for the registry to claim. A system job has no user, so the
// column holds NULL and the row never appears in anybody's job list.
func Enqueue(jobStore store.JobsStore, jobType string) (*model.Job, error) {
	now := time.Now().Unix()

	return jobStore.Create(model.Job{
		ID:        model.NewID(),
		Type:      jobType,
		Status:    model.JobStatusPending,
		Payload:   json.RawMessage("{}"),
		CreatedAt: now,
		UpdatedAt: now,
	})
}
