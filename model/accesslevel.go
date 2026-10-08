// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import (
	"encoding/json"
	"fmt"
)

// AccessLevel is a per-share grant level, ordered by privilege so inherited
// grants combine with MAX (most-permissive). It is not a system role: it stacks
// under the FilePermissions gate as a resource-scoped ACL. The integer is
// private; the API speaks the names from String/ParseAccessLevel, so the storage
// can later become a capability bitmask without changing the wire contract.
type AccessLevel int

const (
	AccessViewer  AccessLevel = 10
	AccessEditor  AccessLevel = 20
	AccessManager AccessLevel = 30
)

func (a AccessLevel) CanEdit() bool { return a >= AccessEditor }

// CanShare gates expanding the audience or changing access: inviting users and
// groups, creating public links, and removing or editing others' access. This
// is Manager and above, so editing content does not imply sharing.
func (a AccessLevel) CanShare() bool { return a >= AccessManager }

func (a AccessLevel) Valid() bool {
	return a == AccessViewer || a == AccessEditor || a == AccessManager
}

func (a AccessLevel) String() string {
	switch a {
	case AccessManager:
		return "manager"
	case AccessEditor:
		return "editor"
	default:
		return "viewer"
	}
}

func ParseAccessLevel(s string) (AccessLevel, error) {
	switch s {
	case "viewer":
		return AccessViewer, nil
	case "editor":
		return AccessEditor, nil
	case "manager":
		return AccessManager, nil
	default:
		return 0, fmt.Errorf("unknown access level %q", s)
	}
}

func (a AccessLevel) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a *AccessLevel) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	lvl, err := ParseAccessLevel(s)
	if err != nil {
		return err
	}

	*a = lvl
	return nil
}
