package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// A map for Puppet Versions JSON
type PuppetVersions map[string]interface{}

// A map for Production Versions JSON
type RunningServices interface{}

// Our bare page
type Page struct {
	Title string
	Body  []byte
	Pv    PuppetVersions
	Rv    RunningServices
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

func getServices(url string) (interface{}, error) {
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

	var available_services interface{}

	err = json.Unmarshal(jsonDataFromHttp, &available_services)
	if err != nil {
		return nil, err
	}

	return available_services, nil
}

func queryServiceVersion(endpoint string) (version string, err error) {
	log.Println("Querying SERVICE address: ", endpoint)
	query_arry := []string{"http://", endpoint, "/info"}
	query := strings.Join(query_arry, "")
	log.Println("Query string: ", query)
	// Query the URI
	resp, err := http.Get(query)
	defer resp.Body.Close()
	if err != nil {
		log.Println("ERROR querying ", query, " ", err)
		return "Failed to get server response for endpoint", err
	}
	// Get data and unmarshel the JSON to our map
	jsonDataFromHttp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("ERROR unmarsheling data for ", jsonDataFromHttp)
		return "Failed to read JSON", err
	}
	var info_response interface{}
	err = json.Unmarshal(jsonDataFromHttp, &info_response)
	if err != nil {
		return "Failed to get info response ", err
	}
	// Parse out the version from the response
	log.Println("INFO for ", endpoint, ":\n", info_response)

	info_map := info_response.(map[string]interface{})
	for _, values := range info_map {
		log.Println("String: ", values)
		sub_info_map := values.(map[string]interface{})
		for key, info := range sub_info_map {
			string_info := info.(string)
			log.Println("Sub info: ", string_info)
			if key == "version" {
				log.Println("Version: ", string_info)
				return string_info, nil
			}
		}
	}
	return version, nil
}

func getVersions(services interface{}) (runningversions map[string]map[string]string, err error) {

	rv := make(map[string]map[string]string)

	s := services.(map[string]interface{})
	for k, v := range s {
		log.Println("Ranging over ", k)
		switch values := v.(type) {
		case map[string]interface{}:
			for name, endpoints := range values {
				log.Println("Found service: ", name)
				rv[name] = make(map[string]string)
				switch eps := endpoints.(type) {
				case []interface{}:
					for _, ep := range eps {
						log.Println("Endpoint: ", ep)
						switch ep_string := ep.(type) {
						case string:
							query_arry := strings.Fields(ep_string)
							if len(query_arry) == 2 {
								log.Println("IP 1: ", query_arry[0])
								log.Println("IP 2: ", query_arry[1])
								info_ep := query_arry[1]
								version, _ := queryServiceVersion(info_ep)
								rv[name][info_ep] = version
							} else {
								log.Println("IP 1: ", query_arry[0])
								info_ep := query_arry[0]
								rv[name][info_ep] = "blah"
							}
						}
					}
				}
			}
		}
	}
	runningversions = rv
	return runningversions, nil
}

func loadPage(title string) (*Page, error) {
	// Get the versions from the puppet master
	log.Println("Getting Puppet Versions - Make sure VPN is on!")
	pv, err := puppetversions("http://puppet.ec2.srcclr.com:1015/versions")
	if err != nil {
		log.Println("Failed to get Puppet Versions from http://puppet.ec2.srcclr.com:1015/versions\n")
		log.Println(err)
	}
	log.Println("Puppet Versions: ", pv)

	// Get running services, prs
	log.Println("Getting available services...")
	prs, err := getServices("http://is.qa.ec2.srcclr.com:3000/services")
	if err != nil {
		log.Println("Failed getting production versions")
	}

	log.Println("RUNNING SERVICES @localhost:3000/services: ", prs)

	// Get running versions
	log.Println("Getting running versions for ", prs)
	rv, err := getVersions(prs)
	if err != nil {
		log.Println("Failed getting versions for ", prs)
	}

	log.Println("Running Versions: ", rv)

	filename := title + ".html"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return &Page{
			Title: title,
			Body:  body,
			Pv:    pv,
			Rv:    rv,
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
