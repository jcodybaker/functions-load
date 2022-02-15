package main

import (
	"fmt"
	"time"
)

func Main(args map[string]interface{}) map[string]interface{} {
	var wait time.Duration
	if waitString, _ := args["wait"].(string); waitString != "" {
		var err error
		wait, err = time.ParseDuration(waitString)
		if err != nil {
			return map[string]interface{}{
				"body": fmt.Sprintf("🤮 failed to parse wait parameter: %v\n", err),
			}
		}
	}
	body := "😵‍💫 no sleep\n"
	if wait != 0 {
		time.Sleep(wait)
		body = fmt.Sprintf("🤩 slept %s\n", wait.String())
	}
	return map[string]interface{}{
		"body": body,
	}
}
