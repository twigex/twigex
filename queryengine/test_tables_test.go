// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import "strings"

// testTablesSQL makes the tables the query tests read, with the same rows in
// every database. BOOLEAN, TRUE, FALSE and DOUBLE PRECISION mean the same in
// MySQL and Postgres. Only a stored time without a time zone has a different
// name: DATETIME in MySQL (MySQL's TIMESTAMP converts times itself), TIMESTAMP
// in Postgres.
func testTablesSQL(databaseType DatabaseType) string {
	timeType := "DATETIME"
	if databaseType == DatabasePostgres {
		timeType = "TIMESTAMP"
	}

	return strings.ReplaceAll(testTablesTemplate, "{time type}", timeType)
}

const archiveSchemaSQL = `
CREATE SCHEMA archive;

CREATE TABLE archive.regions (
	code       VARCHAR(10) NOT NULL,
	closed_on  DATE NULL
);
`

const testTablesTemplate = `
CREATE TABLE sales (
	id        BIGINT PRIMARY KEY,
	country   VARCHAR(50) NULL,
	channel   VARCHAR(50) NOT NULL,
	amount    DECIMAL(12, 2) NOT NULL,
	cost      DECIMAL(12, 2) NOT NULL,
	quantity  INT NOT NULL,
	is_paid   BOOLEAN NULL,
	big_code  BIGINT NOT NULL,
	note      VARCHAR(50) NULL,
	rating    DOUBLE PRECISION NULL
);

INSERT INTO sales VALUES
	(1, 'Latvia',  'Online', 100.50,  60.00, 2, TRUE,  9007199254740993, '50% off',    4.5),
	(2, 'Latvia',  'Shop',   200.00, 150.00, 1, FALSE, 2,                'half_price', 3.0),
	(3, 'Poland',  'Shop',    50.25,  20.00, 3, TRUE,  3,                '5000 units', NULL),
	(4, NULL,      'Online',  10.00,   5.00, 1, NULL,  4,                NULL,         2.0),
	(5, 'Estonia', 'Online', 300.00, 100.00, 5, TRUE,  5,                'free',       5.0);

CREATE TABLE events (
	id           INT PRIMARY KEY,
	happened_at  {time type} NULL
);

INSERT INTO events VALUES
	(1, '2025-03-02 23:59:59'),
	(2, '2025-03-03 00:00:00'),
	(3, '2025-03-03 01:30:00'),
	(4, '2025-03-31 12:00:00'),
	(5, '2025-04-01 00:00:00'),
	(6, NULL);

CREATE TABLE regions (
	id    INT PRIMARY KEY,
	name  VARCHAR(50) NOT NULL
);

INSERT INTO regions VALUES
	(1, 'North'),
	(2, 'South'),
	(3, 'East');

CREATE TABLE stores (
	id          INT PRIMARY KEY,
	name        VARCHAR(50) NOT NULL,
	format      VARCHAR(50) NOT NULL,
	region_id   INT NULL,
	floor_area  INT NOT NULL
);

INSERT INTO stores VALUES
	(1, 'Riga Centre',  'Large', 1,    500),
	(2, 'Riga Airport', 'Small', 1,    120),
	(3, 'Daugavpils',   'Large', 2,    400),
	(4, 'Liepaja',      'Small', NULL, 150),
	(5, 'Ventspils',    'Small', 9,    100);

CREATE TABLE managers (
	id        INT PRIMARY KEY,
	store_id  INT NOT NULL,
	name      VARCHAR(50) NOT NULL
);

INSERT INTO managers VALUES
	(1, 1, 'Anna'),
	(2, 2, 'Janis'),
	(3, 3, 'Liga');

CREATE TABLE channels (
	code   VARCHAR(10) PRIMARY KEY,
	label  VARCHAR(50) NOT NULL
);

INSERT INTO channels VALUES
	('web',  'Website'),
	('shop', 'In store'),
	('app',  'Mobile app');

CREATE TABLE store_sales (
	id        INT PRIMARY KEY,
	store_id  INT NULL,
	channel   VARCHAR(10) NOT NULL,
	amount    DECIMAL(10, 2) NOT NULL,
	quantity  INT NOT NULL
);

INSERT INTO store_sales VALUES
	(1,  1,    'shop', 100.00, 2),
	(2,  1,    'web',   50.00, 1),
	(3,  1,    'shop',  25.50, 1),
	(4,  2,    'shop',  10.00, 1),
	(5,  2,    'web',   40.00, 4),
	(6,  3,    'shop', 300.00, 3),
	(7,  3,    'post',  60.00, 2),
	(8,  4,    'web',   70.00, 1),
	(9,  NULL, 'web',   15.00, 1),
	(10, 8,    'shop',  20.00, 2);
`
