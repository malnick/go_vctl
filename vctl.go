package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// A map for Puppet Versions JSON
type PuppetVersions interface{}

// The final map to be passed to template
type Compared map[string]map[string]map[string]string

// Our bare page
type Page struct {
	Title   string
	Body    []byte
	Compare Compared
}

// What do *you* think this does?
func colorize(versions []string) (color string, err error) {
	if len(versions) > 1 {
		for i := 0; i < len(versions); i++ {
			if versions[i] == "Failed" || versions[i+1] == "Failed" {
				return "green", nil
			}
			if versions[i] == versions[i+1] {
				color = "green"
				return color, nil
			} else {
				color = "red"
				return color, nil
			}
		}
	} else {
		return "green", nil
	}
	return "versions not an array?", err
}

func compare(puppet_v map[string]interface{}, qa_v map[string]map[string]string, prod_v map[string]map[string]string) (Compared, error) {
	c := make(map[string]map[string]map[string]string)
	// Setup regex for QA match
	match_qa, err := regexp.Compile(`_qa`)
	if err != nil {
		log.Println("Couldn't compile regex")
		return nil, err
	}
	match_prod, err := regexp.Compile(`_production`)
	if err != nil {
		log.Println("Couldn't compile regex")
		return nil, err
	}
	// Get environments from PuppetVersions, populate top level map
	c["qa"] = make(map[string]map[string]string)
	c["production"] = make(map[string]map[string]string)

	for p_name, pv := range puppet_v {
		// Create regex match for service name
		p_name_arry := strings.Split(p_name, "_")
		match_name := p_name_arry[0]
		match_svc, _ := regexp.Compile(match_name)
		pv_string := pv.(string)
		// Match QA map
		if match_qa.MatchString(p_name) {
			// Add the name and puppet version to QA map
			c["qa"][match_name] = make(map[string]string)
			c["qa"][match_name]["pv"] = pv_string

			// Init new array, add versions for this service
			colorize_arry := []string{}
			colorize_arry = append(colorize_arry, pv_string)
			for svc_name, endpoints := range qa_v {
				if match_svc.MatchString(svc_name) {
					for ep, version := range endpoints {
						c["qa"][match_name][ep] = version
						colorize_arry = append(colorize_arry, version)
						color, _ := colorize(colorize_arry)
						c["qa"][match_name]["color"] = color
					}
				}
			}
		}
		// Do it again for production env
		if match_prod.MatchString(p_name) {
			log.Println("Production MATCH: ", p_name, " ", pv)
			c["production"][match_name] = make(map[string]string)
			c["production"][match_name]["pv"] = pv_string

			// Init new array, add versions for this service
			colorize_arry := []string{}
			colorize_arry = append(colorize_arry, pv_string)
			for svc_name, endpoints := range prod_v {
				if match_svc.MatchString(svc_name) {
					for ep, version := range endpoints {
						c["production"][match_name][ep] = version
						colorize_arry = append(colorize_arry, version)
						color, _ := colorize(colorize_arry)
						c["production"][match_name]["color"] = color
					}
				}
			}
		}
	}

	return c, nil
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
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer resp.Body.Close()
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
	// Query the URI
	timeout := time.Duration(1 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(query)
	if err != nil {
		log.Println("ERROR querying ", query, " ", err)
		return "Failed", err
	}
	defer resp.Body.Close()

	// Get data and unmarshel the JSON to our map
	jsonDataFromHttp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("ERROR unmarsheling data for ", jsonDataFromHttp)
		return "Failed to read JSON", err
	}
	var info_response interface{}
	err = json.Unmarshal(jsonDataFromHttp, &info_response)
	if err != nil {
		return "Failed", err
	}
	// Parse out the version from the response
	info_map := info_response.(map[string]interface{})
	log.Println("Response: ", info_response)
	for _, values := range info_map {
		sub_info_map := values.(map[string]interface{})
		for key, info := range sub_info_map {
			string_info := info.(string)
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
	for _, v := range s {
		switch values := v.(type) {
		case map[string]interface{}:
			for name, endpoints := range values {
				log.Println("Found service: ", name)
				rv[name] = make(map[string]string)
				switch eps := endpoints.(type) {
				case []interface{}:
					for _, ep := range eps {
						switch ep_string := ep.(type) {
						case string:
							query_arry := strings.Fields(ep_string)
							if len(query_arry) == 2 {
								info_ep := query_arry[1]
								version, _ := queryServiceVersion(info_ep)
								rv[name][info_ep] = version
							} else {
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

	// QA
	log.Println("Getting available services...")
	qa_rs, err := getServices("http://is.qa.ec2.srcclr.com:3000/services")
	if err != nil {
		log.Println("Failed getting qa versions")
	}

	log.Println("RUNNING SERVICES QA: ", qa_rs)

	qa_v, err := getVersions(qa_rs)
	if err != nil {
		log.Println("Failed getting versions for ", qa_rs)
	}

	log.Println("Running Versions QA: ", qa_v)

	// PRODUCTION
	log.Println("Getting available services...")
	prod_rs, err := getServices("http://is.ec2.srcclr.com:3000/services")
	if err != nil {
		log.Println("Failed getting production versions")
	}

	log.Println("RUNNING SERVICES QA: ", prod_rs)

	prod_v, err := getVersions(prod_rs)
	if err != nil {
		log.Println("Failed getting versions for ", prod_rs)
	}

	log.Println("Running Versions PRODUCTION: ", prod_v)

	// Build the compared map of maps of strings of other types ... blah blah blah
	pv_map := pv.(map[string]interface{})
	compared, _ := compare(pv_map, qa_v, prod_v)

	for k, v := range compared {
		log.Println(k, " ", v, "\n")
	}

	filename := title + ".html"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return &Page{
			Title:   title,
			Body:    body,
			Compare: compared,
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
	//log.Println("Serving:\n", string(p.Title), string(p.Body))
}

func main() {
	log.Println("Starting vctl...")
	http.HandleFunc("/", viewHandler)
	http.ListenAndServe(":9000", nil)
}
