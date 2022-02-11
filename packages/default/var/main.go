package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

func Main(args map[string]interface{}) map[string]interface{} {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return wrapErr(errors.New("DATABASE_URL is not set"))
	}

	dbURL, err := dburl.Parse(databaseURL)
	if err != nil {
		return nil, errors.Wrap(err, "parsing DATABASE_URL")
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

	// check db
	if err = db.Ping(); err != nil {
		return wrapErr(err, "connecting to postgres")
	}
	return wrapHTML("success!")
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

func initialMigration(db *gorm.DB) {
	db.AutoMigrate(&model.Note{})
}

// Get gets a Note from the DB
func (p *PG) Get(id string) (*model.Note, error) {
	var note model.Note
	err := p.DB.Where("uuid = ?", id).Take(&note).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "getting note from db")
	}

	return &note, nil
}

// Create creates a note
func (p *PG) Create(note *model.Note) error {
	err := p.DB.Create(note).Error
	if err != nil {
		return errors.Wrap(err, "creating note in db")
	}

	return nil
}
