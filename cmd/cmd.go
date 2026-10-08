// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package cmd

import (
	"github.com/twigex/twigex/config"
)

func openConfig() (*config.ConfigStore, error) {
	c, err := config.NewConfig()

	if err != nil {
		return nil, err
	}

	return c, nil
}
