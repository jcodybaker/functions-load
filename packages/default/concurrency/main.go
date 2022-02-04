package main

import (
	"errors"
	"os"
	"strings"
)

func Main(args map[string]interface{}) map[string]interface{} {

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return wrapErr(errors.New("DATABASE_URL is not set"))
	}

	return wrapHTML("body string")
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