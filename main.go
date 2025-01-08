package main

import (
	"fmt"
	"net/http"
	"stocks-api/router"
)

func main() {
	r := router.Router()
	fmt.Println("Starting server on the port 8080")
	err := http.ListenAndServe(":8080", r)
	fmt.Println(err)
	// log.Fatal(http.ListenAndServe(":8080", r))

}
