package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/xo/dburl"
)

func Main(args map[string]interface{}) map[string]interface{} {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return wrapErr(errors.New("DATABASE_URL is not set"))
	}

	dbURL, err := dburl.Parse(databaseURL)
	if err != nil {
		return wrapErr(err, "parsing DATABASE_URL")
	}

	// Open a DB connection.
	dbPassword, _ := dbURL.User.Password()
	dbName := strings.Trim(dbURL.Path, "/")
	connectionString := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		dbURL.Hostname(),
		dbURL.Port(),
		dbURL.User.Username(),
		dbName, dbPassword,
		dbURL.Query().Get("sslmode"))

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return wrapErr(err, "connecting to postgres")
	}

	defer db.Close()

	testName, _ := args["testname"].(string)
	if testName == "" {
		testName = "default"
	}

	var wait time.Duration
	if waitString, _ := args["wait"].(string); waitString != "" {
		wait, err = time.ParseDuration(waitString)
		if err != nil {
			return wrapErr(err, "parsing duration")
		}
	}

	ctx := context.Background()
	var active, peak, total int
	var pgErr *pq.Error
	if active, peak, total, err = inc(ctx, db, testName); err != nil {
		if errors.As(err, &pgErr) && pgErr.Code == "42P01" {
			err = initDB(ctx, db)
			if err != nil {
				if errors.As(err, &pgErr) {
					wrapErr(err, "initing database: error code"+string(pgErr.Code))
				}
				return wrapErr(err, "initing database")
			}
			active, peak, total, err = inc(ctx, db, testName)
			if err != nil {
				return wrapErr(err, "incrementing after create")
			}
		} else {
			return wrapErr(err, "incrementing")
		}
	}

	if wait != 0 {
		time.Sleep(wait)
	}

	if err = dec(ctx, db, testName); err != nil {
		return wrapErr(err, "decrementing")
	}

	return wrapHTML(fmt.Sprintf("active=%d<br>peak=%d<br>total=%d", active, peak, total))
}

func wrapErr(err error, wrap ...string) map[string]interface{} {
	if len(wrap) == 0 {
		return wrapHTML(`<span style="color: red;">` + err.Error() + "</span>")
	}
	return wrapHTML(`<span style="color: red;">` + wrap[0] + ": " + err.Error() + "\n" + strings.Join(wrap[1:], "\n") + "</span>")
}

func wrapHTML(body string) map[string]interface{} {
	return map[string]interface{}{
		"body": "<html><body><pre>" + string(body) + "</pre></body></html>",
	}
}

func initDB(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
	CREATE TABLE concurrency (
		test_name    varchar(40) NOT NULL,
		con_active   integer NOT NULL,
		con_peak     integer NOT NULL,
		con_total    integer NOT NULL,
		PRIMARY KEY (test_name)
	);
	`)
	return err
}

func inc(ctx context.Context, db *sql.DB, testName string) (active, peak, total int, err error) {
	err = db.QueryRowContext(ctx, `
	INSERT INTO concurrency 
		VALUES ($1, 1, 1, 1)
		ON CONFLICT (test_name)
		DO UPDATE SET 
			con_active = concurrency.con_active + 1,
			con_total = concurrency.con_total + 1,
			con_peak = GREATEST(concurrency.con_peak, concurrency.con_active + 1)
		RETURNING con_active, con_peak, con_total
	`, testName).Scan(&active, &peak, &total)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("inserting: %w", err)
	}
	return
}

func dec(ctx context.Context, db *sql.DB, testName string) error {
	_, err := db.ExecContext(ctx, `
	UPDATE concurrency SET con_active = con_active - 1 WHERE test_name = $1;
	`, testName)
	if err != nil {
		return fmt.Errorf("decrementing: %w", err)
	}
	return nil
}
