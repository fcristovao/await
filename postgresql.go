// Copyright (C) 2016-2018 Betalo AB
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"

	_ "github.com/lib/pq" // Register Postgres driver
)

type postgresqlResource struct {
	url.URL
}

func (r *postgresqlResource) Await(ctx context.Context) error {
	opts := parseFragment(r.URL.Fragment)

	database := strings.TrimPrefix(r.URL.Path, "/")
	if strings.Contains(database, "/") {
		return fmt.Errorf("invalid database name: %s", database)
	}
	if database == "" {
		if _, ok := opts["tables"]; ok {
			return fmt.Errorf("database name required for awaiting tables")
		}
		// Special database default which usually exists.
		database = "information_schema"
	}

	// Disable TLS/SSL by default
	query, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return err
	}
	if query.Get("sslmode") == "" {
		query.Set("sslmode", "disable")
	}

	dsnURL := r.URL
	dsnURL.Fragment = ""
	dsnURL.Path = database
	dsnURL.RawQuery = query.Encode()
	dsn := dsnURL.String()

	db, err := sql.Open(dsnURL.Scheme, dsn)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	if err := db.Ping(); err != nil {
		return &unavailabilityError{err}
	}

	if val, ok := opts["tables"]; ok {
		var tables []string
		if len(val) > 0 && val[0] != "" {
			tables = strings.Split(val[0], ",")
		}
		if err := awaitPostgreSQLTables(db, database, tables); err != nil {
			return err
		}
	}

	return nil
}

func awaitPostgreSQLTables(db *sql.DB, dbName string, tables []string) error {
	if len(tables) == 0 {
		const stmt = `SELECT count(*) FROM information_schema.tables WHERE table_catalog=$1 AND table_schema='public'`
		var tableCnt int
		if err := db.QueryRow(stmt, dbName).Scan(&tableCnt); err != nil {
			return err
		}

		if tableCnt == 0 {
			return &unavailabilityError{errors.New("no tables found")}
		}

		return nil
	}

	const stmt = `SELECT table_name FROM information_schema.tables WHERE table_catalog=$1 AND table_schema='public'`
	rows, err := db.Query(stmt, dbName)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	var actualTables []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return err
		}
		actualTables = append(actualTables, t)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	contains := func(l []string, s string) bool {
		for _, i := range l {
			if i == s {
				return true
			}
		}
		return false
	}

	for _, t := range tables {
		if !contains(actualTables, t) {
			return &unavailabilityError{fmt.Errorf("table not found: %s", t)}
		}
	}

	return nil
}
