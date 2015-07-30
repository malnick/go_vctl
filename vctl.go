package main

import (
	"io/ioutil"
	"log"
)

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func main() {
	p1 := &Page{Title: "VersionCtl", Body: []byte("Eventually a go template...")}
	p1.save()

	p2, err := loadPage("VersionCtl")
	if err != nil {
		log.Println(err)
	}
	log.Println(string(p2.Body))
}
