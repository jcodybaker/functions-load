package main

import (
	"fmt"
)

func Main(args map[string]interface{}) map[string]interface{} {
	fmt.Print("😀 request received!  Try DigitalOcean App Platform")
	return map[string]interface{}{"body": "logged"}
}
