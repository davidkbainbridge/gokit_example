package main

import (
	"log"
	"net/http"

	"github.com/davidkbainbridge/gokit_example/service"
)

func main() {
	service.Register()
	log.Fatal(http.ListenAndServe(":8901", nil))
}
