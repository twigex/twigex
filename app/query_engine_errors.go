// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"errors"
	"net/http"
	"path"
	"strings"

	"github.com/twigex/twigex/collimato"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/queryengine"
	"github.com/twigex/twigex/tlog"
)

// Why the engine could not answer a chart, as translation IDs. Each message
// may name what it is about with {{.Detail}}.
const (
	queryErrorUnknownField         = "collimato.query.unknown_field"
	queryErrorUnsupportedModel     = "collimato.query.unsupported_model"
	queryErrorModelError           = "collimato.query.model_error"
	queryErrorNeedsPrimaryKey      = "collimato.query.needs_primary_key"
	queryErrorNoJoinPath           = "collimato.query.no_join_path"
	queryErrorDifferentConnections = "collimato.query.different_connections"
	queryErrorViewMixed            = "collimato.query.view_mixed"
	queryErrorConnectionMissing    = "collimato.query.connection_missing"
	queryErrorTimeLimit            = "collimato.query.time_limit"
	queryErrorInvalidRequest       = "collimato.query.invalid_request"
)

var queryErrorsByCode = map[queryengine.ErrorCode]string{
	queryengine.ErrorDatasetHasErrors:     queryErrorModelError,
	queryengine.ErrorBadJoin:              queryErrorModelError,
	queryengine.ErrorInvalidTableName:     queryErrorModelError,
	queryengine.ErrorFieldHasNoSQL:        queryErrorModelError,
	queryengine.ErrorFormulaLoop:          queryErrorModelError,
	queryengine.ErrorFormulaMixesDatasets: queryErrorModelError,
	queryengine.ErrorUnknownMeasureType:   queryErrorModelError,
	queryengine.ErrorFieldReferenceInSQL:  queryErrorModelError,
	queryengine.ErrorViewPathsStartApart:  queryErrorModelError,
	queryengine.ErrorNeedsPrimaryKey:      queryErrorNeedsPrimaryKey,
	queryengine.ErrorNoJoinPath:           queryErrorNoJoinPath,
	queryengine.ErrorDifferentConnections: queryErrorDifferentConnections,
	queryengine.ErrorViewMixedWithOthers:  queryErrorViewMixed,
	queryengine.ErrorUnknownDatabaseType:  queryErrorConnectionMissing,
}

// QueryError is why the engine could not answer a chart.
type QueryError struct {
	// ID is the translation of the reason. Empty when the database's own
	// message is given instead.
	ID     string
	Detail string

	DatabaseMessage string
}

func (queryError *QueryError) Error() string {
	if queryError.ID == "" {
		return queryError.DatabaseMessage
	}

	return queryError.ID + ": " + queryError.Detail
}

func newQueryError(id string, detail string) *QueryError {
	return &QueryError{ID: id, Detail: detail}
}

// appError is the error the reader gets. A request the engine cannot answer
// is the request's fault (400); a database that does not answer is not
// Twigex's (502), and a query stopped at the time limit took too long (504).
func (queryError *QueryError) appError() *model.AppError {
	if queryError.ID == "" {
		return model.NewAppErrorRaw(queryError.DatabaseMessage, http.StatusBadGateway)
	}

	status := http.StatusBadRequest

	isConnectionFailure := strings.HasPrefix(queryError.ID, "connection.verification.")
	if isConnectionFailure {
		status = http.StatusBadGateway
	}

	if queryError.ID == queryErrorTimeLimit {
		status = http.StatusGatewayTimeout
	}

	appError := model.NewAppError(queryError.ID, status)
	appError.Values = map[string]any{"Detail": queryError.Detail}

	return appError
}

func (workspace engineWorkspace) requestError(err error) *QueryError {
	var requestError *queryengine.RequestError
	if !errors.As(err, &requestError) {
		return newQueryError(queryErrorInvalidRequest, err.Error())
	}

	if requestError.Code == queryengine.ErrorUnknownField {
		if workspace.comesFromUnsupportedFile(requestError.Field) {
			return newQueryError(queryErrorUnsupportedModel, requestError.Field)
		}

		return newQueryError(queryErrorUnknownField, requestError.Field)
	}

	id, known := queryErrorsByCode[requestError.Code]
	if !known {
		return newQueryError(queryErrorInvalidRequest, requestError.Error())
	}

	return newQueryError(id, requestError.Field)
}

// comesFromUnsupportedFile says whether a field's dataset is named like a
// saved file the engine cannot use. The builder names a dataset's file after
// it, so orders.count comes from orders.yml. A file written by hand may hold
// datasets of other names, which then count as unknown fields.
func (workspace engineWorkspace) comesFromUnsupportedFile(fieldKey string) bool {
	datasetName, _, _ := strings.Cut(fieldKey, ".")

	for _, file := range workspace.unsupported {
		fileDataset := strings.TrimSuffix(file.Name, path.Ext(file.Name))
		if fileDataset == datasetName {
			return true
		}
	}

	return false
}

// databaseError says why the database did not answer. A connection failure
// is named as saving the connection would name it; any other error is the
// database's own message.
func databaseError(workspaceID string, err error) *QueryError {
	if queryengine.QueryReachedTimeLimit(err) {
		return newQueryError(queryErrorTimeLimit, "")
	}

	tlog.Warnw("The database did not answer a chart's query",
		"workspace_id", workspaceID,
		"error", err,
	)

	connectionFailureID, isConnectionFailure := connectionFailures[queryengine.ConnectionFailureOf(err)]
	if isConnectionFailure {
		return newQueryError(connectionFailureID, "")
	}

	return &QueryError{DatabaseMessage: err.Error()}
}

var modelErrorMessages = map[collimato.UnsupportedReason]string{
	collimato.UnsupportedWrittenByHand: "collimato.model.written_by_hand",
	collimato.UnsupportedCannotRead:    "collimato.model.cannot_read",
	collimato.UnsupportedHasErrors:     "collimato.model.has_errors",
}

// TranslateModelErrors writes why each model cannot be used in the reader's
// language.
func TranslateModelErrors(fieldList *collimato.FieldList, locale string) {
	for position, modelError := range fieldList.ModelErrors {
		values := map[string]any{"Name": modelError.Name, "Detail": modelError.Detail}

		appError := model.NewAppError(modelErrorMessages[modelError.Reason], http.StatusOK)
		appError.Values = values

		fieldList.ModelErrors[position].Message = appError.Translate(locale)
	}
}
