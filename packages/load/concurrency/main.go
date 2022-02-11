package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

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
	connectionString := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s", dbURL.Hostname(), dbURL.Port(), dbURL.User.Username(), dbName, dbPassword, dbURL.Query().Get("sslmode"))

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return wrapErr(err, "connecting to postgres")
	}

	defer db.Close()

	testName, _ := args["testname"].(string)
	if testName == "" {
		testName = "default"
	}

	ctx := context.Background()
	var active, peak int
	var pgErr *pq.Error
	if active, peak, err = inc(ctx, db, testName); err != nil {
		if errors.As(err, &pgErr) && pgErr.Code != "42702" { // TODO - add not found code
			err = initDB(ctx, db)
			if err != nil {
				return wrapErr(err, "initing database")
			}
			active, peak, err = inc(ctx, db, testName)
			if err != nil {
				return wrapErr(err, "incrementing after create")
			}
		} else {
			return wrapErr(err, "incrementing")
		}
	}
	if err = dec(ctx, db, testName); err != nil {
		return wrapErr(err, "decrementing")
	}

	var additional string
	if pgErr != nil {
		additional = fmt.Sprintf("%v", pgErr.Code)
	}

	return wrapHTML(fmt.Sprintf("active=%d<br>peak=%d<br>%s", active, peak, additional))
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
		PRIMARY KEY (test_name)
	);
	`)
	return err
}

func inc(ctx context.Context, db *sql.DB, testName string) (current, peak int, err error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("beginning tx: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
		err = tx.Commit()
	}()
	_, err = tx.ExecContext(ctx, `
	INSERT INTO concurrency 
		VALUES ($1, 1, 1)
		ON CONFLICT (test_name)
		DO UPDATE SET con_active = con_active + 1, con_peak = MAX(con_peak, con_active);
	`, testName)
	if err != nil {
		return 0, 0, fmt.Errorf("inserting: %w", err)
	}
	err = tx.QueryRowContext(
		ctx,
		`SELECT con_active, con_peak FROM concurrency WHERE test_name = $1`,
		testName,
	).Scan(&current, &peak)
	if err != nil {
		return 0, 0, fmt.Errorf("querying: %w", err)
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
