// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

const FallbackLocale = "en"

var (
	bundle     *i18n.Bundle
	localizers = make(map[string]*i18n.Localizer)
	mu         sync.RWMutex
)

func Init(localesDir string) error {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	files, err := os.ReadDir(localesDir)
	if err != nil {
		return err
	}

	for _, f := range files {
		if filepath.Ext(f.Name()) == ".json" {
			if _, err := bundle.LoadMessageFile(filepath.Join(localesDir, f.Name())); err != nil {
				return err
			}
		}
	}

	return nil
}

func getLocalizer(locale string) *i18n.Localizer {
	mu.RLock()
	l, ok := localizers[locale]
	mu.RUnlock()

	if ok {
		return l
	}

	mu.Lock()
	defer mu.Unlock()

	if l, ok := localizers[locale]; ok {
		return l
	}

	l = i18n.NewLocalizer(bundle, locale, FallbackLocale)
	localizers[locale] = l

	return l
}

func T(locale, messageID string) string {
	cfg := &i18n.LocalizeConfig{MessageID: messageID}

	msg, _ := getLocalizer(locale).Localize(cfg)

	return msg
}

func Tf(locale, messageID string, data map[string]any) string {
	cfg := &i18n.LocalizeConfig{MessageID: messageID, TemplateData: data}

	msg, _ := getLocalizer(locale).Localize(cfg)

	return msg
}
