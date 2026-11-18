// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import "github.com/twigex/twigex/internal/i18n"

type AppError struct {
	ID      string
	Message string
	Status  int
	Raw     bool

	// Values fill the translation's {{.Name}} places, when it has any.
	Values map[string]any
}

func NewAppError(id string, status int) *AppError {
	return &AppError{
		ID:      id,
		Message: id,
		Status:  status,
		Raw:     false,
	}
}

func NewAppErrorRaw(id string, status int) *AppError {
	return &AppError{
		ID:      id,
		Message: id,
		Status:  status,
		Raw:     true,
	}
}

func (a *AppError) Translate(locale string) string {
	if a.Raw {
		return a.Message
	}

	if a.Values != nil {
		return i18n.Tf(locale, a.ID, a.Values)
	}

	msg := i18n.T(locale, a.ID)

	return msg
}
