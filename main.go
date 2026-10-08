// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/twigex/twigex/cmd"

	_ "github.com/twigex/twigex/enterprise"
)

func main() {
	cmd.Execute()
}
