// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"encoding/json"
	"os"
	"testing"
)

func englishMessageIDs(t *testing.T) map[string]bool {
	t.Helper()

	content, err := os.ReadFile("../i18n/en.json")
	if err != nil {
		t.Fatalf("read en.json: %v", err)
	}

	entries := []struct {
		ID string `json:"id"`
	}{}

	err = json.Unmarshal(content, &entries)
	if err != nil {
		t.Fatalf("decode en.json: %v", err)
	}

	ids := map[string]bool{}

	for _, entry := range entries {
		ids[entry.ID] = true
	}

	return ids
}

func TestEveryEmailNotificationHasASubject(t *testing.T) {
	ids := englishMessageIDs(t)

	for notificationType := range emailNotifications {
		key, known := notificationEmailSubjects[notificationType]
		if !known {
			t.Errorf("%s sends email but has no subject", notificationType)
			continue
		}

		if !ids[key] {
			t.Errorf("%s uses subject %s, which en.json does not have", notificationType, key)
		}
	}

	if !ids["email.notification.subject.default"] {
		t.Error("the fallback subject is missing from en.json")
	}
}
