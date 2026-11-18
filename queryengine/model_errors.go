// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

type ErrorCode string

const (
	ErrorDuplicateName        ErrorCode = "duplicate_name"
	ErrorUnknownJoinTarget    ErrorCode = "unknown_join_target"
	ErrorBrokenViewPath       ErrorCode = "broken_view_path"
	ErrorUnknownChosenField   ErrorCode = "unknown_chosen_field"
	ErrorUsesBrokenDataset    ErrorCode = "uses_broken_dataset"
	ErrorDuplicateFieldInView ErrorCode = "duplicate_field_in_view"

	// ErrorViewReachesDatasetTwice: two of a view's paths reach the same
	// dataset through different datasets. A query joins each dataset once,
	// so one of them would silently use the other's join.
	ErrorViewReachesDatasetTwice ErrorCode = "view_reaches_dataset_twice"

	// ErrorReservedName: a dataset is named like a table the engine makes
	// inside its own queries, such as all_groups or part_1.
	ErrorReservedName ErrorCode = "reserved_name"

	// ErrorNameTooLong: a dataset's name is longer than 63 bytes, which
	// Postgres cuts names to.
	ErrorNameTooLong ErrorCode = "name_too_long"
)

type ModelError struct {
	Code ErrorCode

	// Source is the dataset or view the error is in.
	Source string

	// Field is the field, join target or path step the error is about,
	// or empty when the error is about the whole dataset.
	Field string
}
