package textparser

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

//assume for now that new data that has not been built into the inverted index gets stored
//in some JSON file that is available locally

//smallest unit of data that we store in the database
//this will store each "item" in our search engine with all of the necessary information
//for the interverted index
type Recrod struct {
}

//represents raw data that we will parse objects into before they have been transformed into records
//and stored in our database
type Data struct {
	Title   string   `json:"title"`
	Link    string   `json:"link"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

var data []Data

const dbPath = "data/db.json"

//called at the global start
func ensureDataExists() {
	jsonFile, err := os.Open(dbPath)
	if err != nil {
		f, errCreating := os.Create(dbPath)
		if errCreating != nil {
			log.Fatal("Error, could not create database")
			return
		}
		f.Close()
		data = make([]Data, 0)
	} else {
		defer jsonFile.Close()
	}
}

func loadDataFromJSON() {
	ensureDataExists()
	file, err := os.Open(dbPath)
	if err != nil {
		//TODO: log error permanently
		fmt.Println("Error opening the database")
	}
	//parse the raw JSON data into our array of structs
	json.NewDecoder(file).Decode(&data)
}

//this is the method that will intermittently take the raw data from the current JSON file that needs to be converted
//and "flush it" or put it into the inverted index
//this is the "highest level" method which gets called as part of this script
func flushNewDataIntoInvertedIndex() {
	loadDataFromJSON()
	//for now, assume we have the entire content - later build a web crawler that gets the content
	for i := 0; i < len(data); i++ {
		//need to get a unique ID for the data

		//need to tokenize
		//need to filter and remove stop words

		//need to stem

		//count frequency and create `Record`

		//store record in our inverted index

	}
}

func main() {
	ensureDataExists()
	//for some regular time interval once a day?, flush the new data that has been written to our JSON f
}
