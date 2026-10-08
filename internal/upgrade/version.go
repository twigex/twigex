// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package upgrade

import (
	"strings"

	"github.com/twigex/twigex/internal/parse"
)

// IsReal reports whether v is a release version, so that a local build never
// trips the gate.
func IsReal(v string) bool {
	return v != "" && v[0] >= '0' && v[0] <= '9'
}

// Compare returns -1, 0 or 1. It must stay in agreement with compareVersions in
// frontend/src/store/version.js.
func Compare(a, b string) int {
	numsA, preA := split(a)
	numsB, preB := split(b)

	length := len(numsA)
	if len(numsB) > length {
		length = len(numsB)
	}

	for i := 0; i < length; i++ {
		var x, y int
		if i < len(numsA) {
			x = numsA[i]
		}

		if i < len(numsB) {
			y = numsB[i]
		}

		if x != y {
			if x < y {
				return -1
			}

			return 1
		}
	}

	switch {
	case preA == preB:
		return 0
	case preA == "":
		return 1
	case preB == "":
		return -1
	case preA < preB:
		return -1
	default:
		return 1
	}
}

func split(v string) ([]int, string) {
	core, pre, _ := strings.Cut(v, "-")

	parts := strings.Split(core, ".")
	nums := make([]int, len(parts))
	for i, p := range parts {
		nums[i] = parse.Int(p, 0)
	}

	return nums, pre
}
