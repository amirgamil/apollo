package sources

import (
	"fmt"
	"log"
	"os"

	"github.com/amirgamil/apollo/pkg/apollo/schema"
	jsoniter "github.com/json-iterator/go"
)

//define the schemas here because they're only applicable to the kindle file
type Book struct {
	ASIN       string      `json:"asin"`
	Authors    string      `json:"authors"`
	Highlights []Highlight `json:"highlights"`
	Title      string      `json:"title"`
}

type Highlight struct {
	Text       string      `json:"text"`
	IsNoteOnly bool        `json:"isNoteOnly"`
	Location   Location    `json:"location"`
	Note       interface{} `json:"note"`
}

type Location struct {
	URL   string `json:"url"`
	Value int    `json:"value"`
}

var kindleGlobal map[string]Book

const kindlePath = "./data/kindle.json"

//where new books will be places
const newBooksPath = "./kindle/"

//Kindle does not directly provide a way to get highlights except via https://read.amazon.com/
//I use a readwise extension to download my highlights into JSON https://readwise.io/bookcision
//I then put this json file in a directory called kindle, when it comes time to sync data, i.e. recompute the invertedIndex
//the getKindle method will check for files in this directory, if any exist, it will take them
//and consolidate them to the kindle.

//note save the kindle db in its original state as opposed to saving it in schema.Data format which would reduce repeated work
//in case I want to use the data as is for the future

func getKindle() map[string]schema.Data {
	//ensure our kindle database exists
	ensureFileExists(kindlePath)
	//load our kindle file
	kindleGlobal = make(map[string]Book)
	loadKindle()
	//first check for new files in kindle folder
	newBooks, err := checkForNewBooks()
	if err != nil {
		return make(map[string]schema.Data)
	}
	addNewBooksToDb(newBooks)
	err = writeKindleDbToDisk()
	if err != nil {
		log.Println(err)
	} else {
		//if we successfully write the books to disk, we delete all of the files since
		//we've stored them and no longer need them
		deleteFiles(newBooksPath, newBooksPath)
	}
	bookData := convertBooksToData()
	return bookData
}

func convertBooksToData() map[string]schema.Data {
	//save each highlight as it's own entry as opposed to each book as it's own entry
	data := make(map[string]schema.Data)
	for _, book := range kindleGlobal {
		//iterate through the higlights
		for index, highlight := range book.Highlights {
			//check if this is highlight is already saved
			keyInMap := fmt.Sprintf("srkd%s%d", book.ASIN, index)
			if _, isInMap := sources[keyInMap]; !isInMap {
				note := ""
				if highlight.Note != nil {
					highlightString, isString := highlight.Note.(string)
					if isString {
						note = highlightString
					}
				}
				content := fmt.Sprintf("Highlight: \n\n %s\n\nNote: %s", highlight.Text, note)
				data[keyInMap] = schema.Data{Title: book.Title, Link: highlight.Location.URL, Content: content, Tags: make([]string, 0)}
			}
		}
	}
	return data
}

func writeKindleDbToDisk() error {
	jsonFile, err := os.OpenFile(kindlePath, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	err = jsoniter.NewEncoder(jsonFile).Encode(kindleGlobal)
	return err
}

func addNewBooksToDb(books []Book) {
	for _, book := range books {
		kindleGlobal[book.Title] = book
	}
}

func loadKindle() {
	file, err := os.Open(kindlePath)
	if err != nil {
		log.Println("Error trying to load the kindle database: ", err)
	}
	jsoniter.NewDecoder(file).Decode(&kindleGlobal)
}

func checkForNewBooks() ([]Book, error) {
	files := getFilesInFolder(newBooksPath, "kindle")
	books := make([]Book, 0)
	for _, f := range files {
		if f.Name() == ".DS_Store" {
			continue
		}
		//open the file
		file, err := os.Open(newBooksPath + f.Name())
		if err != nil {
			log.Println("Error trying to open kindle file: ", f.Name(), " with err: ", err)
			return []Book{}, err
		}
		var newBook Book
		err = jsoniter.NewDecoder(file).Decode(&newBook)
		if err != nil {
			log.Println("Uh oh, error decoding book at: ", f.Name(), " with: err: ", err)
			return []Book{}, err
		} else {
			books = append(books, newBook)
		}
	}
	return books, nil
}
