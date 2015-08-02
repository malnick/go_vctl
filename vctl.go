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
type RunningServices map[string]interface{}

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

func parseInfo(name string, info map[string]string) (version string, err error) {
	for k, v := range info {
		for key, value := range info {
			version := info["version"]
			return version, nil
		}
	}
	return "Didn't parse the info map", err
}

func getVersions(services map[string]interface{}) (running_versions map[string]interface{}, err error) {
	for key, service_name := range services {
		for service, ip_arry := range service_name {
			for _, ip_address := range ip_arry {
				query_arry := strings.Fields(ip_address)

				// If our string has one component it's a service address
				if len(query_arry) == 1 {
					mgmt_ip := query_arry[0]

					// Add info endpoint
					info_uri_slice := []string{mgmt_ip, "/info"}
					info_uri := strings.Join(info_uri_slice, "")
					log.Println("Querying SERVICE address for ", service_name, ": ", mgmt_ip)
					// Query the URI
					resp, err := http.Get(info_uri)
					defer resp.Body.Close()
					if err != nil {
						log.Println("ERROR querying ", service_name, " ", err)
						return nil, err
					}
					// Get data and unmarshel the JSON to our map
					jsonDataFromHttp, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						log.Println("ERROR unmarsheling data for ", service_name, " from ", jsonDataFromHttp)
						return nil, err
					}
					var info_response map[string]interface{}
					err = json.Unmarshal(jsonDataFromHttp, &info_response)
					if err != nil {
						return nil, err
					}
					// Parse out the version from the response
					log.Println("INFO for ", service_name, ":\n", info_response)
					versions := parseInfo(service_name, info_response)

					return versions, nil

					// If our string has two components then use the mgmt address
				} else if len(query_arry) == 2 {
					mgmt_ip := query_arry[1]
					log.Println("Querying MGMT address for ", service_name, ": ", mgmt_ip)
					resp, err := http.Get(mgmt_ip)
					defer resp.Body.Close()
					if err != nil {
						log.Println("ERROR querying ", service_name, " ", err)
						return nil, err
					}

					jsonDataFromHttp, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						log.Println("ERROR unmarsheling data for ", service_name, " from ", jsonDataFromHttp)
						return nil, err
					}
					// If our string as more than two components, error
				} else if len(query_arry) > 2 {
					return nil, err
				}
			}
		}
	}

}

func loadPage(title string) (*Page, error) {
	// Get the versions from the puppet master
	log.Println("Getting Puppet Versions - Make sure VPN is on!")
	pv, err := puppetversions("http://puppet.ec2.srcclr.com:1015/versions")
	if err != nil {
		log.Println("Failed to get Puppet Versions from http://puppet.ec2.srcclr.com:1015/versions\n")
		log.Println(err)
	}

	// Get production running services, prs
	log.Println("Getting available services...")
	prs, err := getServices("http://localhost:3000/services")
	if err != nil {
		log.Println("Failed getting production versions")
	}

	// Get running versions
	log.Println("Getting running versions for ", prs)
	rv, err := getVersions(prs)
	if err != nil {
		log.Println("Failed getting versions for ", prs)
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
