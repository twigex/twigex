// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package upgrade

// Stop is a release that raised the floor, and the version it required had
// already been run.
type Stop struct {
	Version string
	MinFrom string
	Reason  string
}

// Stops records every release that raised model.MinUpgradableVersion, oldest
// first. Append one when skipping a release would lose or corrupt data.
var Stops = []Stop{
	{
		Version: "0.15.0",
		MinFrom: "0.14.0",
		Reason:  "first public release",
	},
}

type Step struct {
	Version string
	Reason  string
}

// Plan returns the releases that must be run between cur and bin, in order.
// Each step is handed to an operator as an instruction, so a floor must never
// name a version that was not released.
func Plan(cur, bin string) []Step {
	var steps []Step

	target := bin
	for {
		floor := floorFor(target)
		if floor == "" || Compare(cur, floor) >= 0 {
			break
		}

		steps = append([]Step{{Version: floor, Reason: reasonFor(floor, target)}}, steps...)
		target = floor
	}

	return steps
}

// floorFor takes the newest matching entry, so Stops has to stay sorted.
func floorFor(v string) string {
	floor := ""
	for _, s := range Stops {
		if Compare(s.Version, v) <= 0 {
			floor = s.MinFrom
		}
	}

	return floor
}

func reasonFor(version, requiredBy string) string {
	if reason := stopReason(version); reason != "" {
		return reason
	}

	return "required by " + requiredBy
}

func stopReason(version string) string {
	for _, s := range Stops {
		if s.Version == version {
			return s.Reason
		}
	}

	return ""
}
