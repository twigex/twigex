// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package upgrade

import (
	"errors"
	"fmt"
	"strings"

	"github.com/twigex/twigex/model"
)

const (
	VersionKey       = "app.version"
	MinUpgradableKey = "app.min_upgradable_version"
	UpgradedAtKey    = "app.upgraded_at"

	SkipEnv = "TWIGEX_SKIP_VERSION_CHECK"
)

type Schema struct {
	Installed bool
	Version   int64
	Dirty     bool
}

// Check reports whether this binary must not run against the database as it
// stands. An empty recorded version stands in as model.LastUngatedVersion.
func Check(recorded string, schema Schema) error {
	if !IsReal(model.Version) || !schema.Installed {
		return nil
	}

	if schema.Dirty {
		return errors.New(dirtyMessage(schema.Version))
	}

	cur, assumed := recorded, false
	if cur == "" {
		cur, assumed = model.LastUngatedVersion, true
	}

	return decide(cur, model.Version, model.MinUpgradableVersion, assumed)
}

// decide refuses a downgrade only against a recorded version. An assumed one is
// a guess, and a guess one release out (0.14.0 against a 0.14.0-rc4 binary)
// must not read as a rollback.
func decide(cur, bin, floor string, assumed bool) error {
	switch order := Compare(cur, bin); {
	case order == 0:
		return nil
	case order > 0 && !assumed:
		return errors.New(downgradeMessage(cur, bin))
	case IsReal(floor) && Compare(cur, floor) < 0:
		return errors.New(blockedMessage(cur, bin, floor))
	default:
		return nil
	}
}

func dirtyMessage(version int64) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Migration %d did not finish, and the database is marked dirty.\n", version)
	b.WriteString("Its schema is half applied, so neither the recorded version nor the\n")
	b.WriteString("migration state can be trusted.\n\n")
	b.WriteString("Restore the backup taken before that upgrade.")

	return b.String()
}

func downgradeMessage(cur, bin string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "This database was last used by %s, which is newer than this %s binary.\n", cur, bin)
	b.WriteString("Starting an older release against a newer schema can corrupt data.\n\n")
	fmt.Fprintf(&b, "Install %s or later, or restore the backup taken before that upgrade.\n\n", cur)
	fmt.Fprintf(&b, "Set %s=1 to override. Data loss is likely.", SkipEnv)

	return b.String()
}

func blockedMessage(cur, bin, floor string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Cannot upgrade directly from %s to %s.\n", cur, bin)
	fmt.Fprintf(&b, "This release requires %s or later to have been run already.\n", floor)

	steps := Plan(cur, bin)
	if len(steps) == 0 {
		b.WriteString("\nNo upgrade path is available for a version this old. Restore the backup\n")
		b.WriteString("taken before the upgrade, or contact support.\n")
	} else {
		width := len(bin)
		for _, s := range steps {
			if len(s.Version) > width {
				width = len(s.Version)
			}
		}

		b.WriteString("\nUpgrade one release at a time:\n\n")
		for i, s := range steps {
			fmt.Fprintf(&b, "  %d. %-*s  %s\n", i+1, width, s.Version, s.Reason)
		}

		last := stopReason(bin)
		if last == "" {
			last = "this release"
		}

		fmt.Fprintf(&b, "  %d. %-*s  %s\n", len(steps)+1, width, bin, last)
		fmt.Fprintf(&b, "\nInstall %s, start it once and let it finish migrating, then continue.\n", steps[0].Version)
	}

	fmt.Fprintf(&b, "\nSet %s=1 to override. Data loss is likely.", SkipEnv)

	return b.String()
}
