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

//func parseInfo(info map[string]string) (version string, err error) {
//	for k, v := range info {
//		for key, value := range info {
//			version := info["version"]
//			return version, nil
//		}
//	}
//	return "Didn't parse the info map", err
//}

func queryServiceVersion(endpoint string) (version string, err error) {
	log.Println("Querying SERVICE address: ", endpoint)
	// Query the URI
	resp, err := http.Get(endpoint)
	defer resp.Body.Close()
	if err != nil {
		log.Println("ERROR querying ", endpoint, " ", err)
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

// If our string has one component it's a service address
//				if len(query_arry) == 1 {
//					mgmt_ip := query_arry[0]
//
//					// Add info endpoint
//					info_uri_slice := []string{mgmt_ip, "/info"}
//					info_uri := strings.Join(info_uri_slice, "")
//
//					version, _ := parseInfo(info_response)
//
//					runningversions.Service[service_name][info_uri] = append(runningversions.Service[service_name][info_uri], version)

// If our string has two components then use the mgmt address
//				} else if len(query_arry) == 2 {
//					mgmt_ip := query_arry[1]
//
//					// Add info endpoint
//					info_uri_slice := []string{mgmt_ip, "/info"}
//					info_uri := strings.Join(info_uri_slice, "")
//					log.Println("Querying SERVICE address for ", service_name, ": ", mgmt_ip)
//					// Query the URI
//					resp, err := http.Get(info_uri)
//					defer resp.Body.Close()
//					if err != nil {
//						log.Println("ERROR querying ", service_name, " ", err)
//						return nil, err
//					}
//					// Get data and unmarshel the JSON to our map
//					jsonDataFromHttp, err := ioutil.ReadAll(resp.Body)
//					if err != nil {
//						log.Println("ERROR unmarsheling data for ", service_name, " from ", jsonDataFromHttp)
//						return nil, err
//					}
//					var info_response map[string]interface{}
//					err = json.Unmarshal(jsonDataFromHttp, &info_response)
//					if err != nil {
//						return nil, err
//					}
//					// Parse out the version from the response
//					log.Println("INFO for ", service_name, ":\n", info_response)
//					versions := parseInfo(service_name, info_response)
//
//					return versions, nil
//
//					// If our string as more than two components, error
//				} else if len(query_arry) > 2 {
//					return nil, err
//				}
//	return runningversions, err
//}

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
