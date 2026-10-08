// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package catalog

import (
	"context"
	"fmt"
)

type Catalog interface {
	Tables(ctx context.Context) ([]TableName, error)
	Close() error
}

func Open(ctx context.Context, c Connection) (Catalog, error) {
	switch c.Type {
	case Mysql:
		return openSQL(ctx, c, mysqlDialect)
	case Postgre:
		return openSQL(ctx, c, psqlDialect)
	default:
		return nil, fmt.Errorf("catalog: unsupported store %s", c.Type)
	}
}
