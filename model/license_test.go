// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import (
	"testing"
	"time"
)

func activeLicense() *License {
	return &License{
		StartsAt:  time.Now().Add(-time.Hour).Unix(),
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
		Features:  &Features{Users: NewInt(50)},
	}
}

func lapsedLicense() *License {
	return &License{
		StartsAt:  time.Now().Add(-2 * time.Hour).Unix(),
		ExpiresAt: time.Now().Add(-time.Hour).Unix(),
		Features:  &Features{Users: NewInt(10), Groups: NewBool(true)},
	}
}

func futureLicense() *License {
	return &License{
		StartsAt:  time.Now().Add(time.Hour).Unix(),
		ExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		Features:  &Features{Users: NewInt(99), Groups: NewBool(true)},
	}
}

func TestFeatureReadsFollowChainToTheLiveLicense(t *testing.T) {
	head := lapsedLicense()
	head.Next = activeLicense()
	head.Next.Features.Groups = NewBool(true)

	if !head.HasGroups() {
		t.Error("a lapsed head must hand over to the queued license")
	}
	if got := head.SeatLimit(); got != 50 {
		t.Errorf("seat limit must come from the live license: want 50, got %d", got)
	}
}

func TestLapsedLicenseWithNoRenewalGrantsNothing(t *testing.T) {
	head := lapsedLicense()

	if head.HasGroups() {
		t.Error("a lapsed license with no successor must grant nothing")
	}
	if head.ActiveLicense() != nil {
		t.Error("ActiveLicense must be nil when no license covers now")
	}
}

func TestQueuedLicenseGrantsNothingBeforeItsTerm(t *testing.T) {
	head := activeLicense()
	head.Features.Groups = NewBool(false)
	head.Next = futureLicense()

	if head.HasGroups() {
		t.Error("a queued license must not grant features before its term")
	}
	if got := head.SeatLimit(); got != 50 {
		t.Errorf("want the current license's 50 seats, got %d", got)
	}
}

func TestOrderLicensesSortsByStartDate(t *testing.T) {
	now := time.Now().Unix()

	current := activeLicense()
	later := futureLicense()
	sooner := futureLicense()
	sooner.StartsAt = time.Now().Add(30 * time.Minute).Unix()

	head := OrderLicenses(current, []*License{later, current, sooner}, now)

	if head != current {
		t.Fatal("the license in force must be the head")
	}
	if head.Next != sooner {
		t.Error("the soonest renewal must come first")
	}
	if head.Next.Next != later {
		t.Error("the later renewal must follow")
	}
	if head.Next.Next.Next != nil {
		t.Error("the chain must terminate")
	}
}

func TestOrderLicensesLeadsWithRenewalWhenNothingCurrent(t *testing.T) {
	now := time.Now().Unix()
	queued := futureLicense()

	head := OrderLicenses(nil, []*License{queued}, now)

	if head != queued {
		t.Fatal("the queued renewal must become the head")
	}
	if head.ActiveLicense() != nil {
		t.Error("a queued renewal must not be in force before its term")
	}
	if head.Queued() != queued {
		t.Error("the renewal must be reportable as queued")
	}
}

func TestSelectLicensePrefersMostRecentlyIssued(t *testing.T) {
	now := time.Now().Unix()

	original := activeLicense()
	original.IssuedAt = now - 1000

	upgrade := activeLicense()
	upgrade.IssuedAt = now - 10
	upgrade.Features.Users = NewInt(100)

	if got := SelectLicense([]*License{original, upgrade}, now); got != upgrade {
		t.Error("a mid-term reissue must win over the license it replaces")
	}
}

// The seat guard in AddLicense keys off SeatHardLimit, so it does not run for a
// renewal: an uploaded renewal covering fewer seats than are in use is accepted
// and only bites when its term starts.
func TestRenewalReportsNoSeatLimitUntilItsTermStarts(t *testing.T) {
	renewal := futureLicense()

	if got := renewal.SeatHardLimit(); got != 0 {
		t.Errorf("SeatHardLimit before the term starts: want 0, got %d", got)
	}
	if renewal.BlocksOnSeatLimit() {
		t.Error("a license that has not started must not block")
	}
	if got := *renewal.Features.Users; got != 99 {
		t.Errorf("the raw seat count stays readable: want 99, got %d", got)
	}
}

func TestTooManyUsersJudgesARenewalBeforeItStarts(t *testing.T) {
	renewal := futureLicense()
	renewal.Features.Users = NewInt(50)

	if !renewal.TooManyUsers(60) {
		t.Error("a renewal covering 50 seats must be refused at 60 active users")
	}
	if renewal.TooManyUsers(50) {
		t.Error("exactly the licensed count must be accepted")
	}

	renewal.ExtraUsers = NewInt(5)
	if renewal.TooManyUsers(55) {
		t.Error("the grace band must count toward the upload limit")
	}
	if !renewal.TooManyUsers(56) {
		t.Error("past the grace band must be refused")
	}
}

func TestTooManyUsersIgnoresUnlimitedAndTrueUp(t *testing.T) {
	unlimited := futureLicense()
	unlimited.Features.Users = NewInt(0)

	if unlimited.TooManyUsers(10000) {
		t.Error("unlimited seats must never be too many")
	}

	trueUp := futureLicense()
	trueUp.Features.Users = NewInt(5)
	trueUp.EnforcementMode = EnforcementTrueUp

	if trueUp.TooManyUsers(500) {
		t.Error("true_up settles the overage at renewal and must not refuse")
	}

	var absent *License
	if absent.TooManyUsers(100) {
		t.Error("an absent license must not refuse")
	}
}

func TestSeatHardLimitAddsGrace(t *testing.T) {
	l := activeLicense()
	l.ExtraUsers = NewInt(5)

	if got := l.SeatLimit(); got != 50 {
		t.Errorf("soft limit: want 50, got %d", got)
	}
	if got := l.SeatHardLimit(); got != 55 {
		t.Errorf("hard limit: want 55, got %d", got)
	}
}

func TestSeatHardLimitWithoutGrace(t *testing.T) {
	if got := activeLicense().SeatHardLimit(); got != 50 {
		t.Errorf("want 50, got %d", got)
	}
}

// Unlimited must stay unlimited: a grace band on top of "no limit" would
// otherwise turn 0 into a real ceiling.
func TestSeatHardLimitUnlimitedIgnoresGrace(t *testing.T) {
	l := activeLicense()
	l.Features.Users = NewInt(0)
	l.ExtraUsers = NewInt(5)

	if got := l.SeatHardLimit(); got != 0 {
		t.Errorf("want 0 (unlimited), got %d", got)
	}
}

func TestSeatLimitsAreZeroWithoutLicense(t *testing.T) {
	var l *License

	if got := l.SeatLimit(); got != 0 {
		t.Errorf("SeatLimit: want 0, got %d", got)
	}
	if got := l.SeatHardLimit(); got != 0 {
		t.Errorf("SeatHardLimit: want 0, got %d", got)
	}
	if l.BlocksOnSeatLimit() {
		t.Error("an absent licence must not block")
	}
}

func TestExpiredLicenseStopsBlocking(t *testing.T) {
	l := activeLicense()
	l.ExpiresAt = time.Now().Add(-time.Minute).Unix()

	if l.BlocksOnSeatLimit() {
		t.Error("an expired licence must not block")
	}
	if got := l.SeatLimit(); got != 0 {
		t.Errorf("expired SeatLimit: want 0, got %d", got)
	}
}

// Licences issued before the mode existed carry no value, and must keep the
// original blocking behaviour rather than silently loosening.
func TestBlocksOnSeatLimitDefaultsToBlocking(t *testing.T) {
	if !activeLicense().BlocksOnSeatLimit() {
		t.Error("a licence with no enforcement mode must block")
	}
}

func TestTrueUpNeverBlocks(t *testing.T) {
	l := activeLicense()
	l.EnforcementMode = EnforcementTrueUp

	if l.BlocksOnSeatLimit() {
		t.Error("true_up must not block")
	}
}

func TestStorageLimitBytes(t *testing.T) {
	l := activeLicense()
	l.Deployment = DeploymentSaaS

	if got := l.StorageLimitBytes(); got != 0 {
		t.Errorf("no limits block: want 0, got %d", got)
	}

	l.Limits = &Limits{StorageGB: NewInt(2)}
	if got, want := l.StorageLimitBytes(), int64(2*1024*1024*1024); got != want {
		t.Errorf("want %d, got %d", want, got)
	}

	l.Limits = &Limits{StorageGB: NewInt(0)}
	if got := l.StorageLimitBytes(); got != 0 {
		t.Errorf("zero GB means unlimited: want 0, got %d", got)
	}
}

func TestSetDefaultsFillsDeploymentAndEnforcement(t *testing.T) {
	l := activeLicense()
	l.SetDefaults()

	if l.Deployment != DeploymentSelfHosted {
		t.Errorf("deployment: want %q, got %q", DeploymentSelfHosted, l.Deployment)
	}
	if l.EnforcementMode != EnforcementBlock {
		t.Errorf("enforcement: want %q, got %q", EnforcementBlock, l.EnforcementMode)
	}
	if l.IsSaaS() {
		t.Error("a self-hosted licence must not report as SaaS")
	}
}

// A ceiling on a self-hosted licence is malformed, and must be ignored rather
// than throttling a customer who supplies their own storage.
func TestStorageLimitIgnoredWhenSelfHosted(t *testing.T) {
	l := activeLicense()
	l.Deployment = DeploymentSelfHosted
	l.Limits = &Limits{StorageGB: NewInt(2)}

	if got := l.StorageLimitBytes(); got != 0 {
		t.Errorf("self-hosted must be unlimited: want 0, got %d", got)
	}
}
