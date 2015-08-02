package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

// A map for Puppet Versions JSON
type PuppetVersions map[string]interface{}

// A map for Production Versions JSON
type ProductionVersions map[string]interface{}

// Our bare page
type Page struct {
	Title string
	Body  []byte
	Pv    PuppetVersions
	Prodv ProductionVersions
}

func puppetversions(url string) (PuppetVersions, error) {
	// Get response and handle any errors on return
	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Read JSON from request
	jsonDataFromHttp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshel the JSON to our struct
	var v PuppetVersions
	err = json.Unmarshal(jsonDataFromHttp, &v)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func getServices(url string) (map[string]interface{}, error) {
	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	jsonDataFromHttp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	available_services := make(map[string]interface{})

	err = json.Unmarshal(jsonDataFromHttp, &available_services)
	if err != nil {
		return nil, err
	}

	log.Println("AVAILABLE SERVICES for ", url, ":\n", available_services)

	return available_services, nil
}

func loadPage(title string) (*Page, error) {
	// Get the versions from the puppet master
	log.Println("Getting Puppet Versions - Make sure VPN is on!")
	pv, err := puppetversions("http://puppet.ec2.srcclr.com:1015/versions")
	if err != nil {
		log.Println("Failed to get Puppet Versions from http://puppet.ec2.srcclr.com:1015/versions\n")
		log.Println(err)
	}

	// Get running versions
	log.Println("Getting available services...")
	prv, err := getServices("http://localhost:3000/services")
	if err != nil {
		log.Println("Failed getting production versions")
	}

	filename := title + ".html"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	log.Println("Puppet Versions: ", pv)

	return &Page{
			Title: title,
			Body:  body,
			Pv:    pv,
			Prodv: prv,
		},
		nil
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Loading page view...")
	title := "versionctl"
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}

	// Parse the template, execute and write it to stdout for good measure
	log.Println("Parsing go template...")
	t, _ := template.ParseFiles("versionctl.html")
	t.Execute(w, p)
	log.Println("Serving:\n", string(p.Title), string(p.Body))
}

func main() {
	log.Println("Starting vctl...")
	http.HandleFunc("/", viewHandler)
	http.ListenAndServe(":9000", nil)
}
