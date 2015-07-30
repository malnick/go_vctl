package main

import (
	"html"
	"log"
	"net/http"
	// My library
	"github.com/malnick/vctl_lib"
)

func config() {

	config := GetConfig()
	log.Printf(config.Pvendpoint)
}

func run() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	log.Fatal(http.ListenAndServe(":8080", nil))

}
