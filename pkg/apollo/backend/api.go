package backend

import (
	"fmt"
	"log"
	"os"

	"github.com/amirgamil/apollo/pkg/apollo/schema"
	"github.com/amirgamil/apollo/pkg/apollo/sources"
	jsoniter "github.com/json-iterator/go"
)

//assume for now that new data that has not been built into the inverted index gets stored
//in some JSON file that is available locally

//TODO: fix all error handling

//maps tokens to an array of pointers to records
//maps strings or tokens to array of record ids
var globalInvertedIndex map[string][]string

//global list of all records which are stored locally
//maps strings which are unique ids of each record to the record
var localRecordList map[string]schema.Record

//global list of records pull in from data sources
var sourcesRecordList map[string]schema.Record

//database of inverted index for ALL of the data
//maps strings (i.e tokens) to string ids
const invertedIndexPath = "./data/index.json"

//database of all of records stored locally
//all ids start with lc<an integer>
const localRecordsPath = "./data/local.json"

//database of the records we compute from the sources
//all ids start with sr<potentially more identifiying inforation><an integer>
const sourcesPath = "./data/sources.json"

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
	ensureDataExists(sourcesPath)
	ensureDataExists(invertedIndexPath)
	ensureDataExists(localRecordsPath)
	globalInvertedIndex = make(map[string][]string)
	localRecordList = make(map[string]schema.Record)
	sourcesRecordList = make(map[string]schema.Record)
	loadGlobals()
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

func loadRecordsList(path string, list map[string]schema.Record) {
	jsonFile, err := os.Open(path)
	if err != nil {
		fmt.Println("Error, could not load the inverted index")
		return
	}
	defer jsonFile.Close()
	jsoniter.NewDecoder(jsonFile).Decode(&list)
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
	//flags we pass here are important, need to replace the entire file
	jsonFile, err := os.OpenFile(invertedIndexPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		fmt.Println("Error trying to write the new inverted index to disk")
	}
	defer jsonFile.Close()
	jsoniter.NewEncoder(jsonFile).Encode(globalInvertedIndex)
}

//helper method which writes the current the record list to disk
//parameters determine which record list we write
func writeRecordListToDisk(path string, list map[string]schema.Record) {
	//flags we pass here are important, need to replace the entire file
	jsonFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		fmt.Println("Error trying to write the new inverted index to disk")
	}
	defer jsonFile.Close()
	jsoniter.NewEncoder(jsonFile).Encode(list)
}

//highest-level function that is called at regular intervals to recompute the ENTIRE inverted index to integrate
//new data added via Apollo, resync data from the data sources, and include any saved records to Apollo
func RefreshInvertedIndex() {
	//loads the globals we need including the new data, previous records, and our current inverted index
	loadGlobals()
	//clean inverted index
	globalInvertedIndex = make(map[string][]string)
	//Order is important here
	//Step 1: Write local stored records to the inverted index
	flushSavedRecordsIntoInvertedIndex(localRecordList)
	//Step 2: Write "old" data from data sources to the inverted index
	//"old" = we've previously done work to retrieve and process
	flushSavedRecordsIntoInvertedIndex(sourcesRecordList)
	//Step 3a: resync data from data sources i.e. get all data again
	sourceData := sources.GetData(sourcesRecordList)
	//Step 3b: flush new data from data sources into inverted index, note we DO NOT save these records locally since they are stored
	//in the origin of where we pulled them from. Since we get ALL of the data from our data sources each time this method is called, this
	//prevents creating additional copies in our inverted index
	flushDataSourcesIntoInvertedIndex(sourceData)

	//write data to disk in inverted index and record JSON file
	writeIndexToDisk()
	writeRecordListToDisk(localRecordsPath, localRecordList)
	writeRecordListToDisk(sourcesPath, sourcesRecordList)
}

func loadGlobals() {
	loadInvertedIndex()
	loadRecordsList(localRecordsPath, localRecordList)
	loadRecordsList(sourcesPath, sourcesRecordList)
}

//takes all of the saved records and puts them in our inverted index
func flushSavedRecordsIntoInvertedIndex(recordList map[string]schema.Record) {
	//we already have token frequency data precomputed and saved, so just add it to inverted index directly
	for key, record := range recordList {
		writeTokenFrequenciesToInvertedIndex(record.TokenFrequency, key)
	}
}

func GetRecordFromData(currData schema.Data, uniqueID string) schema.Record {

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
	record := schema.Record{ID: uniqueID, Title: currData.Title, Link: currData.Link, Content: currData.Content, TokenFrequency: frequencyOfTokens}
	return record
}

//method takes data and flushes it into our inverted index
//Note since th
func flushDataSourcesIntoInvertedIndex(data map[string]schema.Data) {
	for uniqueID, currData := range data {
		record := GetRecordFromData(currData, uniqueID)
		sourcesRecordList[uniqueID] = record
		writeTokenFrequenciesToInvertedIndex(record.TokenFrequency, uniqueID)
	}
}

//write a map of tokens to their counts in our inverted index
func writeTokenFrequenciesToInvertedIndex(frequencyOfTokens map[string]int, uniqueID string) {
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

func AddNewEntryToLocalData(data schema.Data) {
	key := fmt.Sprintf("lc%d", len(localRecordList))
	record := GetRecordFromData(data, key)
	localRecordList[key] = record
	writeRecordListToDisk(localRecordsPath, localRecordList)
}
