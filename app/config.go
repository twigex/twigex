// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import "github.com/twigex/twigex/model"

func (a *App) GetClientConfig(authenticated bool) map[string]any {
	config := map[string]any{
		"DefaultLocale":             *a.ConfigStore.Config.ServerSettings.DefaultLocale,
		"AllowUserLanguageOverride": *a.ConfigStore.Config.ServerSettings.AllowUserLanguageOverride,
	}

	if authenticated {
		config["Version"] = model.Version
		config["BuildDate"] = model.BuildDate
		config["BuildHash"] = model.BuildHash
		config["Edition"] = model.Edition
		config["EnterpriseRevision"] = model.EnterpriseRevision
		config["EnterpriseHash"] = model.EnterpriseHash
	}

	return config
}
