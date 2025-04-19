package main

import (
	"database/sql"
	"fmt"

	_ "github.com/marcboeker/go-duckdb"
)

func parseRows(rows *sql.Rows) ([]map[string]any, error) {
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]any

	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))

		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		rowMap := make(map[string]any)
		for i, col := range columns {
			var v any
			val := values[i]

			bytes, ok := val.([]byte)
			if ok {
				v = string(bytes)
			} else {
				v = val
			}
			rowMap[col] = v
		}

		results = append(results, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func runQuery(query string) ([]map[string]any, error) {
	db, err := sql.Open("duckdb", "")
	if err != nil {
		return nil, fmt.Errorf("couldn't open DuckDB connection: %w", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE OR REPLACE SECRET secret (TYPE s3, PROVIDER credential_chain);`)
	if err != nil {
		return nil, fmt.Errorf("couldn't create DuckDB secret: %w", err)
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("couldn't execute DuckDB query: %w", err)
	}

	results, err := parseRows(rows)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse query results: %w", err)
	}

	return results, nil
}
