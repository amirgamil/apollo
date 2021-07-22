package sources

import (
	"encoding/gob"
	"log"
	"os"

	"github.com/amirgamil/apollo/pkg/apollo/backend"
)

const zeusPath = "./zeus/db.gob"

type List struct {
	Key  string   `json:"key"`
	Data []string `json:"data"`
	//rule represents markdown of what a unit of the list looks like
	Rule string `json:"rule"`
}

func getZeus() {

}

func loadZeusData() {
	//set of paths to ignore
	ignore := map[string]bool{"podcasts": true}
	cache := make(map[string]*List)
	dataToIndex := make([]backend.Data, 0)
	file, err := os.Open(zeusPath)
	if err != nil {
		log.Fatal("Error loading data from zeus")
	}
	gob.NewDecoder(file).Decode(&cache)
	for key, val := range cache {
		_, toIgnore := ignore[key]
		if !toIgnore {
			getDataFromList(val, dataToIndex)
		}
	}
}

func getDataFromList(list *List, data []backend.Data) {
	for _, listData := range list.Data {
		//try to scrape link if applicable

		//if it fails, send back the link, using tag words from the link
	}
}

//takes a rule and returns whether we should *try to* scrape a list item in Zeus
func shouldScrapeLink(rule string) {

}
