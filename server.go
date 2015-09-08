package main

import (
	"log"
	"net/http"

	"dkb/service"
)

func main() {
	service.Register()
	log.Fatal(http.ListenAndServe(":8901", nil))
}
