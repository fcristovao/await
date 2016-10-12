// Copyright (c) 2016 Betalo AB
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
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"context"

	_ "github.com/lib/pq" // Register Postgres driver
)

type postgresqlResource struct {
	url.URL
}

func (r *postgresqlResource) Await(ctx context.Context) error {
	// Keep original resource value unmodified
	dsnURL := r.URL

	// Parse and remove tags from fragment
	tags := parseTags(r.URL.Fragment)
	dsnURL.Fragment = ""

	// Disable TLS/SSL by default
	query, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return err
	}
	if query.Get("sslmode") == "" {
		query.Set("sslmode", "disable")
	}
	dsnURL.RawQuery = query.Encode()

	dsn := dsnURL.String()

	db, err := sql.Open(dsnURL.Scheme, dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return &unavailableError{err}
	}

	if val, ok := tags["tables"]; ok {
		tables := strings.Split(val, ",")
		if err := awaitPostgreSQLTables(db, dsnURL.Path[1:], tables); err != nil {
		}
	}

	return nil
}

func awaitPostgreSQLTables(db *sql.DB, dbName string, tables []string) error {
	if len(tables) == 0 {
		const stmt = `SELECT count(*) FROM information_schema.tables WHERE table_catalog=? AND table_schema='public'`
		var tableCnt int
		if err := db.QueryRow(stmt, dbName).Scan(&tableCnt); err != nil {
			return err
		}

		if tableCnt == 0 {
			return &unavailableError{errors.New("no tables found")}
		}

		return nil
	}

	const stmt = `SELECT table_name FROM information_schema.tables WHERE table_catalog=? AND table_schema='public'`
	rows, err := db.Query(stmt, dbName)
	if err != nil {
		return err
	}
	defer rows.Close()

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
			return &unavailableError{fmt.Errorf("table not found: %s", t)}
		}
	}

	return nil
}
