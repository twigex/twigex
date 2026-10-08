// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package cmd

import (
	"github.com/twigex/twigex/app"
	"github.com/twigex/twigex/config"
	"github.com/twigex/twigex/store"
)

func openApp() (*app.App, error) {
	c, err := config.NewConfig()
	if err != nil {
		return nil, err
	}

	st, err := store.NewNewStore(c.Config)
	if err != nil {
		return nil, err
	}

	a := &app.App{
		ConfigStore: *c,
		Store:       *st,
	}

	if err := a.LoadSystemSettings(); err != nil {
		return nil, err
	}

	return a, nil
}
