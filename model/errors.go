// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import "github.com/twigex/twigex/internal/i18n"

type AppError struct {
	ID      string
	Message string
	Status  int
	Raw     bool
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

	msg := i18n.T(locale, a.ID)

	return msg
}
