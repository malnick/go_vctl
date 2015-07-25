package main

import (
	"log"
	"menteslibres.net/gosexy/rest"
	"net/url"
)

func main() {
	// Enable debug mode
	// rest.Debug = true

	var err error

	// Get the Puppet Master versions
	log.Printf("vctl is starting up...\n")

	// Request vars
	versions := rest.Response{}
	requestURL := "http://puppet.ec2.srcclr.com:1015/versions"
	requestVariables := url.Values{}

	err = rest.Get(&versions, requestURL, requestVariables)

	if err == nil {

		// Printing response dump.
		log.Printf("Got response!")
		log.Printf("Response code: %d", versions.StatusCode)
		log.Printf("Response protocol version: %s", versions.Proto)
		log.Printf("Response length: %d", versions.ContentLength)
		log.Printf("Response header: %v", versions.Header)
		log.Printf("Response body: %s", string(versions.Body))
	} else {
		// Yes, we had an error.
		log.Printf("Request failed: %s", err.Error())
	}
}
