// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import (
	"encoding/hex"
	"unicode/utf8"
)

// EncodeLDAPID renders an immutable directory identifier as a stable string.
// Active Directory returns objectGUID as 16 raw bytes while OpenLDAP returns
// entryUUID as text, so binary values are hex encoded and printable ones are kept
// readable. The choice is made from the bytes themselves and is therefore the same
// on every login for a given directory.
func EncodeLDAPID(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}

	if isPrintableUTF8(raw) {
		return string(raw)
	}

	return hex.EncodeToString(raw)
}

func isPrintableUTF8(b []byte) bool {
	if !utf8.Valid(b) {
		return false
	}

	for _, r := range string(b) {
		if r < 0x20 || r == 0x7f {
			return false
		}
	}

	return true
}
