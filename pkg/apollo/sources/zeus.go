package sources

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/amirgamil/apollo/pkg/apollo/schema"
)

const zeusPath = "../zeus/db.gob"

type List struct {
	Key  string   `json:"key"`
	Data []string `json:"data"`
	//rule represents markdown of what a unit of the list looks like
	Rule string `json:"rule"`
}

//TODO: add some intelligent code to not scrape if seen before, otherwise server costs to the moon lol
func getZeus() []schema.Data {
	//set of paths to ignore
	ignore := map[string]bool{"podcasts": true, "startups": true}
	cache := make(map[string]*List)
	dataToIndex := make([]schema.Data, 0)
	file, err := os.Open(zeusPath)
	if err != nil {
		log.Fatal("Error loading data from zeus")
	}
	gob.NewDecoder(file).Decode(&cache)
	for key, val := range cache {
		_, toIgnore := ignore[key]
		if !toIgnore {
			getDataFromList(val, &dataToIndex)
		}
	}
	return dataToIndex
}

func getDataFromList(list *List, data *[]schema.Data) {
	for index, listData := range list.Data {
		//create model of the document first - recall items in Zeus are stored as rendered markdown which means HTML
		listDoc, err := goquery.NewDocumentFromReader(strings.NewReader(listData))
		if err != nil {
			log.Fatal("Error parsing item in list component!")
		}
		var newItem schema.Data
		//use some heuristics to decide whether we should `scrape` a link or
		//just put it raw in our database
		//need to navigate to the `body` of the pased HTML since goquery automatically populates html, head, and body
		body := listDoc.Find("body")
		firstChild := body.Children().Nodes[0]
		secondChild := firstChild.FirstChild
		//If we only have an a tag or one inside another tag, this is probably an item we want to scrape (e.g. /articles)
		if firstChild.Data == "a" || secondChild.Data == "a" {
			newItem, err = scrapeLink(listDoc)
			if err != nil {
				fmt.Println("Error parsing link in list: ", listData, " defaulting to use link")
			}
		} else {
			fmt.Println(body.Text())
			//otherwise, there's other content which we assume will (hopefully be indexable), may be adapted to be more intelligent
			newItem = schema.Data{Title: fmt.Sprintf("%s %d", list.Key, index), Link: "zeus.amirbolous.com/" + list.Key, Content: body.Text(), Tags: make([]string, 0)}
		}

		//if it fails, send back the link, using tag words from the link
		*data = append(*data, newItem)
	}
}

//takes a document which is suspected to be an article or something that's scrapable and attempts to scrape it
func scrapeLink(listDoc *goquery.Document) (schema.Data, error) {
	var data schema.Data
	var err error
	listDoc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, hasLink := s.Attr("href")
		if hasLink {
			data, err = schema.Scrape(link) //TODO: check regex, scrape w. Text()?
			if err != nil {
				//add URL directly as data, to have our tokenizer extract something meaningful, we try to replace
				//as many symbol we might find in URLs with spaces so the tokenizer can extract a couple of meaningful words
				//from the title-
				cleanedUpData := strings.ReplaceAll(link, "/", " ")
				cleanedUpData = strings.ReplaceAll(link, "-", " ")
				//Throw in the parent's title as well which might be useful, since most links are of the form <p><a></a></p>
				cleanedUpData += s.Parent().Text()
				data = schema.Data{Title: s.Parent().Text(), Content: cleanedUpData, Link: link, Tags: make([]string, 0)}
			}
		} else {
			data = schema.Data{}
		}
	})
	return data, err
}
