package main

import (
	"log"

	"gateway-platform-plugin/server/helpers"
	"gateway-platform-plugin/server/router"
)

func main() {
	app := helpers.MustBootstrap()
	if err := router.RunHTTP(app); err != nil {
		log.Fatal(err)
	}
}
