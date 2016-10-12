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

	_ "github.com/go-sql-driver/mysql" // Register MySQL driver
)

type mysqlResource struct {
	url.URL
}

func (r *mysqlResource) Await(ctx context.Context) error {
	// Keep original resource value unmodified
	dsnURL := r.URL

	// Parse and remove tags from fragment
	tags := parseTags(r.URL.Fragment)
	dsnURL.Fragment = ""

	// Comply to Go's MySQL driver DSN convention
	dsnURL.Host = "tcp(" + dsnURL.Host + ")"

	dsn := dsnURL.String()

	// Comply to Go's MySQL driver DSN convention
	dsn = strings.TrimPrefix(dsn, "mysql://")

	db, err := sql.Open(dsnURL.Scheme, dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return &unavailableError{err}
	}

	if val, ok := tags["tables"]; ok {
		var tables []string
		if val != "" {
			tables = strings.Split(val, ",")
		}
		if err := awaitMySQLTables(db, dsnURL.Path[1:], tables); err != nil {
			return err
		}
	}

	return nil
}

func awaitMySQLTables(db *sql.DB, dbName string, tables []string) error {
	if len(tables) == 0 {
		const stmt = `SELECT count(*) FROM information_schema.tables WHERE table_schema=?`
		var tableCnt int
		if err := db.QueryRow(stmt, dbName).Scan(&tableCnt); err != nil {
			return err
		}

		if tableCnt == 0 {
			return &unavailableError{errors.New("no tables found")}
		}

		return nil
	}

	const stmt = `SELECT table_name FROM information_schema.tables WHERE table_schema=?`
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
