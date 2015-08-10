package main

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

var qa_urls = []string{
	"http://is.qa.ec2.srcclr.com:3000/services",
	"http://10.0.3.103:3000/services",
	"http://10.0.3.126:3000/services",
}

var prod_urls = []string{
	"http://is.ec2.srcclr.com:3000/services",
}

// A map for Puppet Versions JSON
type PuppetVersions interface{}

// The final map to be passed to template
type Compared map[string]map[string]map[string]string

// Our bare page
type Page struct {
	Title   string
	Body    []byte
	Compare Compared
	Time    string
}

// What do *you* think this does?
func colorize(versions []string) (color string, err error) {
	match_fail, _ := regexp.Compile(`Failed`)
	if len(versions) > 1 {
		for i := 0; i < len(versions); i++ {
			if match_fail.MatchString(versions[i]) {
				return "orange", nil
			}
		}
		for i := 0; i < len(versions); i++ {
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

func getServices(urls []string) (map[string][]string, error) {
	available_services := make(map[string][]string)

	for _, url := range urls {
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

		var services map[string]map[string][]string //interface{}

		err = json.Unmarshal(jsonDataFromHttp, &services)
		if err != nil {
			return nil, err
		}

		for _, v := range services {
			log.Println("Values: ", v)
			for name, endpoints := range v {
				log.Println("Name: ", name, " ", endpoints)
				available_services[name] = []string{}
				for _, ep := range endpoints {
					available_services[name] = append(available_services[name], ep)
				}
			}
		}
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
		return fmt.Sprintf("Failed: %s", err), err
	}
	defer resp.Body.Close()

	// Get data and unmarshel the JSON to our map
	jsonDataFromHttp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("ERROR unmarsheling data for ", jsonDataFromHttp)
		return fmt.Sprintf("Failed: %s", err), err
	}
	var info_response interface{}
	err = json.Unmarshal(jsonDataFromHttp, &info_response)
	if err != nil {
		return fmt.Sprintf("Failed: %s", err), err
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

func getVersions(services map[string][]string) (runningversions map[string]map[string]string, err error) {

	rv := make(map[string]map[string]string)

	for name, endpoints := range services {
		rv[name] = make(map[string]string)
		for _, endpoint := range endpoints {
			query_arry := strings.Fields(endpoint)
			if len(query_arry) == 2 {
				info_ep := query_arry[1]
				version, _ := queryServiceVersion(info_ep)
				rv[name][endpoint] = version
			} else {
				info_ep := query_arry[0]
				version, _ := queryServiceVersion(info_ep)
				rv[name][endpoint] = version
			}
		}
	}
	runningversions = rv
	return runningversions, nil
}

func refreshState() {
	for {
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
		qa_rs, err := getServices(qa_urls)
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
		prod_rs, err := getServices(prod_urls)
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
		datafile, err := os.Create("compared.gob")
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		dataEncoder := gob.NewEncoder(datafile)
		dataEncoder.Encode(compared)
		datafile.Close()

		log.Println("Sleeping...")
		time.Sleep(time.Second * 10)
	}
}

func loadPage(title string) (*Page, error) {
	var t = time.Now().String()
	log.Println("TIME", t)

	// Hit and quit it
	go refreshState()

	//read state file to compared
	var compared Compared
	datafile, err := os.Open("compared.gob")
	if err != nil {
		log.Println(err)
	}

	dataDecoder := gob.NewDecoder(datafile)
	err = dataDecoder.Decode(&compared)
	if err != nil {
		log.Println(err)
	}

	datafile.Close()

	filename := title + ".html"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &Page{
			Title:   title,
			Body:    body,
			Compare: compared,
			Time:    t,
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
	fmt.Println(`____   ____                  .__              ____________________.____      `)
	fmt.Println(`\   \ /   /___________  _____|__| ____   ____ \_   ___ \__    ___/|    |     `)
	fmt.Println(` \   Y   // __ \_  __ \/  ___/  |/  _ \ /    \/    \  \/ |    |   |    |     `)
	fmt.Println(`  \     /\  ___/|  | \/\___ \|  (  <_> )   |  \     \____|    |   |    |___  `)
	fmt.Println(`   \___/  \___  >__|  /____  >__|\____/|___|  /\______  /|____|   |_______ \ `)
	fmt.Println(`              \/           \/               \/        \/                  \/ `)

	log.Println("Starting vctl...")
	http.HandleFunc("/", viewHandler)
	http.ListenAndServe(":9000", nil)
}
