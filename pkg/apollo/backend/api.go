package backend

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	jsoniter "github.com/json-iterator/go"
)

//assume for now that new data that has not been built into the inverted index gets stored
//in some JSON file that is available locally

//smallest unit of data that we store in the database
//this will store each "item" in our search engine with all of the necessary information
//for the interverted index
type Record struct {
	//unique identifier
	ID string `json:"id"`
	//title
	Title string `json:"title"`
	//potential link to the source if applicable
	Link string `json:"link"`
	//map of tokens to their frequency
	tokenFrequency map[string]int //TODO - is this exported?
}

//represents raw data that we will parse objects into before they have been transformed into records
//and stored in our database
type Data struct {
	Title   string   `json:"title"`
	Link    string   `json:"link"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
	//TODO: add metadata, should be able to search based on type of record, document, podcast, personal etc.
}

//TODO: fix all error handling

//list of newly added data that has yet to be added to our inverted index - this is flushed periodically
var data []Data

//maps tokens to an array of pointers to records
//maps strings or tokens to array of record ids
var globalInvertedIndex map[string][]string

//global list of every single record
var globalRecordList map[string]Record

//database of all NEW data that has been accumalated that needs to be flushed into the inverted index
const dbPath = "./data/db.json"

//database of inverted index for ALL of the data
//maps strings (i.e tokens) to string ids
const invertedIndexPath = "./data/index.json"

//database of all of records
const recordsPath = "./data/records.json"

func createFile(path string) {
	f, errCreating := os.Create(path)
	if errCreating != nil {
		log.Fatal("Error, could not create database for path: ", path, " with: ", errCreating)
		return
	}
	f.Close()
}

//called at the global start
func ensureDataExists(path string) {
	jsonFile, err := os.Open(path)
	if err != nil {
		createFile(path)
	} else {
		defer jsonFile.Close()
	}
}

//helper function which should be called when the program is initialized so that the necessary files and paths
//exist in our database
func InitializeFilesAndData() {
	ensureDataExists(dbPath)
	ensureDataExists(invertedIndexPath)
	ensureDataExists(recordsPath)

}

//loads the inverted path from disk to memory
//opt to use a more optimized JSON decoding and encoding library than Go's native one as our inverted index JSON files grow in size and cloud money ain't free
func loadInvertedIndex() {
	jsonFile, err := os.Open(invertedIndexPath)
	if err != nil {
		fmt.Println("Error, could not load the inverted index")
		return
	}
	defer jsonFile.Close()
	//TODO: not sure if we can decode into pointers?
	jsoniter.NewDecoder(jsonFile).Decode(&globalInvertedIndex)
}

func loadRecordsList() {
	jsonFile, err := os.Open(recordsPath)
	if err != nil {
		fmt.Println("Error, could not load the inverted index")
		return
	}
	defer jsonFile.Close()
	jsoniter.NewDecoder(jsonFile).Decode(&globalRecordList)
}

//loads new data that needs to be flushed into our records and inverted index
func loadNewData() {
	data = make([]Data, 0)
	jsonFile, err := os.Open(dbPath)
	defer jsonFile.Close()
	if err != nil {
		//TODO: log error permanently
		fmt.Println("Error opening the database")
	}
	//parse the raw JSON data into our array of structs
	json.NewDecoder(jsonFile).Decode(&data)
}

//takes a string of tokens and returns a map of each token to its frequency
func countFrequencyTokens(tokens []string) map[string]int {
	frequencyWords := make(map[string]int)
	for _, token := range tokens {
		_, isInMap := frequencyWords[token]
		if isInMap {
			frequencyWords[token] += 1
		} else {
			frequencyWords[token] = 1
		}
	}
	return frequencyWords
}

//helper method which writes the current the inverted index to disk
func writeIndexToDisk() {
	jsonFile, err := os.OpenFile(invertedIndexPath, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		fmt.Println("Error trying to write the new inverted index to disk")
	}
	defer jsonFile.Close()
	jsoniter.NewEncoder(jsonFile).Encode(globalInvertedIndex)
}

//helper method which empties the JSON file containing the previously new data to add our search engine's data
func emptyTempDatabase() {
	//call os.Create which will create a new empty JSON file, don't want to keep storing the old data since we don't need it
	createFile(dbPath)
}

//this is the method that will intermittently take the raw data from the current JSON file that needs to be converted
//and "flush it" or put it into the inverted index
//this is the "highest level" method which gets called as part of this script
func FlushNewDataIntoInvertedIndex() {
	loadNewData()
	loadInvertedIndex()
	//for now, assume we have the entire content - later build a web crawler that gets the content
	for i := 0; i < len(data); i++ {
		currData := data[i]
		//need to get a unique ID for the data - use the number of records we have so far (i.e. length of the record list)
		uniqueID := fmt.Sprint(len(globalRecordList))
		//tokenize, stem, and filter
		tokens := Analyze(currData.Content)

		//count frequency and create `Record`
		frequencyOfTokens := countFrequencyTokens(tokens)

		//adds meta level tags defined into the data - how do we set the frequency? Since these are global tags
		//we push some more probability on them since the user said these were important to index by
		//use a simple heuristic of pushing ~20% of "counts" on them
		//TODO: is there a more intellignet heuristic we can use here
		frequencyToAdd := len(tokens) / 5
		for _, metaTag := range currData.Tags {
			_, metaTagInMap := frequencyOfTokens[metaTag]
			if metaTagInMap {
				frequencyOfTokens[metaTag] += frequencyToAdd
			} else {
				frequencyOfTokens[metaTag] = frequencyToAdd
			}
		}

		//store record in our tokens list
		record := Record{ID: uniqueID, Title: currData.Title, Link: currData.Link, tokenFrequency: frequencyOfTokens}
		globalRecordList[uniqueID] = record

		//loop through final frequencyOfTokens and add it to our inverted index database
		for key, _ := range frequencyOfTokens {
			_, keyInInvertedIndex := globalInvertedIndex[key]
			if keyInInvertedIndex {
				globalInvertedIndex[key] = append(globalInvertedIndex[key], uniqueID)
			} else {
				globalInvertedIndex[key] = []string{uniqueID}
			}
		}

	}
	//write data to disk in inverted index JSON file
	writeIndexToDisk()

	//empty the database file since we have flushed into the index and persisted to disk
	emptyTempDatabase()
}
