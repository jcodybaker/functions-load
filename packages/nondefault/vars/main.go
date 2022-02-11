package main

import (
	"fmt"
	"os"
)

func Main(args map[string]interface{}) map[string]interface{} {
	return wrapHTML(fmt.Sprintf(
		"PROJECT_LEVEL: %q\nPACKAGE_LEVEL: %q\nACTION_LEVEL: %q\n",
		os.Getenv("PROJECT_LEVEL"),
		os.Getenv("PACKAGE_LEVEL"),
		os.Getenv("ACTION_LEVEL"),
	))
}

func wrapHTML(body string) map[string]interface{} {
	return map[string]interface{}{
		"body": "<html><body><pre>" + string(body) + "</pre></body></html>",
	}
}
