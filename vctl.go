package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

// Request for Puppet Versions: "http://puppet.ec2.srcclr.com:1015/versions"
//type PuppetVersions struct {
//	ServiceName string
//	AnalyticsVersionProduction       string `json:"analytics_version_production"`
//	AnalyticsVersionQa               string `json:"analytics_version_qa"`
//	BlogVersionProduction            string `json:"blog_version_production"`
//	FrontendStaticVersionDev         string `json:"frontend_static_version_dev"`
//	FrontendStaticVersionProduction  string `json:"frontend_static_version_production"`
//	FrontendStaticVersionQa          string `json:"frontend_static_version_qa"`
//	LibrarianVersionProduction       string `json:"librarian_version_production"`
//	LibrarianVersionQa               string `json:"librarian_version_qa"`
//	NotificationsVersionProduction   string `json:"notifications_version_production"`
//	NotificationsVersionQa           string `json:"notifications_version_qa"`
//	PlatformVersionProduction        string `json:"platform_version_production"`
//	PlatformVersionQa                string `json:"platform_version_qa"`
//	ScmAgentVersionProduction        string `json:"scm_agent_version_production"`
//	ScmAgentVersionQa                string `json:"scm_agent_version_qa"`
//	SearchVersionProduction          string `json:"search_version_production"`
//	SearchVersionQa                  string `json:"search_version_qa"`
//	VulnerabilitiesVersionProduction string `json:"vulnerabilities_version_production"`
//	VulnerabilitiesVersionQa         string `json:"vulnerabilities_version_qa"`
//	WebhooksVersionProduction        string `json:"webhooks_version_production"`
//	WebhooksVersionQa                string `json:"webhooks_version_qa"`
//}

type PuppetVersions map[string]interface{}

// Our bare page
type Page struct {
	Title string
	Body  []byte
	Pv    PuppetVersions
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

func loadPage(title string) (*Page, error) {
	pv, err := puppetversions("http://puppet.ec2.srcclr.com:1015/versions")
	if err != nil {
		log.Println("Failed to get Puppet Versions from http://puppet.ec2.srcclr.com:1015/versions\n")
		log.Println(err)
	}

	filename := title + ".html"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	log.Println("VERSIONS\n")
	log.Println("%v", pv["scm_agent_version_production"])

	return &Page{
			Title: title,
			Body:  body,
			Pv:    pv,
		},
		nil
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title := "versionctl"
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}

	// Parse the template, execute and write it to stdout for good measure
	t, _ := template.ParseFiles("versionctl.html")
	t.Execute(w, p)
	log.Println("Serving:\n", string(p.Title), string(p.Body))
}

func main() {
	http.HandleFunc("/", viewHandler)
	http.ListenAndServe(":9000", nil)
}
