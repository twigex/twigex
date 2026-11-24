// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

const (
	ErrorNoFields             ErrorCode = "no_fields"
	ErrorMissingRowLimit      ErrorCode = "missing_row_limit"
	ErrorUnknownField         ErrorCode = "unknown_field"
	ErrorWrongFieldKind       ErrorCode = "wrong_field_kind"
	ErrorDatasetHasErrors     ErrorCode = "dataset_has_errors"
	ErrorInvalidTableName     ErrorCode = "invalid_table_name"
	ErrorFieldHasNoSQL        ErrorCode = "field_has_no_sql"
	ErrorMeasuresInRowsOnly   ErrorCode = "measures_in_rows_only"
	ErrorUnknownFilterType    ErrorCode = "unknown_filter_type"
	ErrorMissingFilterValue   ErrorCode = "missing_filter_value"
	ErrorBadFilterGroup       ErrorCode = "bad_filter_group"
	ErrorMixedFilterGroup     ErrorCode = "mixed_filter_group"
	ErrorNotADateField        ErrorCode = "not_a_date_field"
	ErrorUnknownDateStep      ErrorCode = "unknown_date_step"
	ErrorUnknownDateRange     ErrorCode = "unknown_date_range"
	ErrorBadDate              ErrorCode = "bad_date"
	ErrorFieldShownTwice      ErrorCode = "field_shown_twice"
	ErrorNoJoinPath           ErrorCode = "no_join_path"
	ErrorBadJoin              ErrorCode = "bad_join"
	ErrorDifferentConnections ErrorCode = "different_connections"
	ErrorViewMixedWithOthers  ErrorCode = "view_mixed_with_others"
	ErrorNeedsPrimaryKey      ErrorCode = "needs_primary_key"
	ErrorFormulaLoop          ErrorCode = "formula_loop"
	ErrorFormulaMixesDatasets ErrorCode = "formula_mixes_datasets"
	ErrorSortFieldNotShown    ErrorCode = "sort_field_not_shown"
	ErrorUnknownDatabaseType  ErrorCode = "unknown_database_type"
	ErrorUnknownMeasureType   ErrorCode = "unknown_measure_type"
	ErrorFieldReferenceInSQL  ErrorCode = "field_reference_in_sql"
	ErrorViewPathsStartApart  ErrorCode = "view_paths_start_apart"

	// ErrorColumnNamesClash: two fields would get the same column name in
	// the SQL, such as a.b__c and a__b.c, or two long names alike in their
	// first 63 bytes, where Postgres cuts names off.
	ErrorColumnNamesClash ErrorCode = "column_names_clash"
)

// RequestError is why a request cannot be turned into a query. Unlike
// ModelError, which describes a model, it is a failure, so it is a Go error.
type RequestError struct {
	Code ErrorCode

	// Field is the field, dataset or table the error is about, if any.
	Field string
}

func (requestError *RequestError) Error() string {
	if requestError.Field == "" {
		return string(requestError.Code)
	}

	return string(requestError.Code) + ": " + requestError.Field
}

func newRequestError(code ErrorCode, field string) *RequestError {
	return &RequestError{Code: code, Field: field}
}
