// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

// The preference name for a notification is hand-written in two maps whose
// values differ only by an app/email prefix and an irregular suffix, so a row
// can point at the wrong toggle and still look right. These checks pin the
// invariants that mistake breaks.

import (
	"strings"
	"testing"

	"github.com/twigex/twigex/model"
)

func TestAppNotificationsUseAppPreferences(t *testing.T) {
	for notificationType, preference := range appNotifications {
		if !strings.HasPrefix(preference, "notify_app_") {
			t.Errorf("%s is gated on %q, which is not an in-app preference", notificationType, preference)
		}
	}
}

func TestEmailNotificationsUseEmailPreferences(t *testing.T) {
	for notificationType, preference := range emailNotifications {
		if !strings.HasPrefix(preference, "notify_email_") {
			t.Errorf("%s is gated on %q, which is not an email preference", notificationType, preference)
		}
	}
}

func TestBothMapsCoverTheSameNotificationTypes(t *testing.T) {
	for notificationType := range appNotifications {
		if _, ok := emailNotifications[notificationType]; !ok {
			t.Errorf("%s has an in-app preference but no email one", notificationType)
		}
	}
	for notificationType := range emailNotifications {
		if _, ok := appNotifications[notificationType]; !ok {
			t.Errorf("%s has an email preference but no in-app one", notificationType)
		}
	}
}

func TestEveryPreferenceHasADefault(t *testing.T) {
	for _, preferences := range []map[string]string{appNotifications, emailNotifications} {
		for notificationType, preference := range preferences {
			if _, ok := model.DefaultNotifications[preference]; !ok {
				t.Errorf("%s reads %q, which DefaultNotifications does not define, so the toggle can never be found", notificationType, preference)
			}
		}
	}
}
