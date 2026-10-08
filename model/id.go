// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import (
	"strings"

	"github.com/google/uuid"
)

func NewID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}
