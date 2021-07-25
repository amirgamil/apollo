package sources

import (
	"fmt"
	"io/ioutil"
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

//How I use this?

//Kindle does not directly provide a way to get highlights except via https://read.amazon.com/
//I use a readwise extension to download my highlights into JSON https://readwise.io/bookcision
//I then put this json file in a directory called kindle, when it comes time to sync data, i.e. recompute the invertedIndex
//the getKindle method will check for files in this directory, if any exist, it will take them
//and consolidate them to the kindle.

//note save the kindle db in its original state as opposed to saving it in schema.Data format which would reduce repeated work
//in case I want to use the data as is for the future

func GetKindle() []schema.Data {
	//ensure our kindle database exists
	ensureKindleFileExists()
	//load our kindle file
	kindleGlobal = make(map[string]Book)
	loadKindle()
	//first check for new files in k
	newBooks, err := checkForNewBooks()
	if err != nil {
		return []schema.Data{}
	}
	addNewBooksToDb(newBooks)
	err = writeKindleDbToDisk()
	if err != nil {
		log.Println(err)
	} else {
		//if we successfully write the books to disk, we delete all of the files since
		//we've stored them and no longer need them
		deleteBookFiles()
	}
	bookData := convertBooksToData()
	return bookData
}

func deleteBookFiles() {
	files := getFilesInKindle()
	for _, f := range files {
		err := os.Remove(newBooksPath + f.Name())
		if err != nil {
			log.Println("Error deleting file: ", f.Name())
		}
	}
}

func convertBooksToData() []schema.Data {
	//save each highlight as it's own entry as opposed to each book as it's own entry
	data := make([]schema.Data, 0)
	for _, book := range kindleGlobal {
		//iterate through the higlights
		for _, highlights := range book.Highlights {
			content := fmt.Sprintf("Highlight: \n %s\nNote: %s", highlights.Text, highlights.Note)
			data = append(data, schema.Data{Title: book.Title, Link: highlights.Location.URL, Content: content, Tags: make([]string, 0)})
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
	fmt.Println(kindleGlobal)
	err = jsoniter.NewEncoder(jsonFile).Encode(kindleGlobal)
	return err
}

func addNewBooksToDb(books []Book) {
	for _, book := range books {
		kindleGlobal[book.Title] = book
	}
}

func ensureKindleFileExists() {
	jsonFile, err := os.Open(kindlePath)
	if err != nil {
		file, err := os.Create(kindlePath)
		if err != nil {
			log.Println("Error creating the original kindle database: ", err)
			return
		}
		file.Close()
	} else {
		defer jsonFile.Close()
	}
}

func loadKindle() {
	file, err := os.Open(kindlePath)
	if err != nil {
		log.Println("Error trying to load the kindle database: ", err)
	}
	jsoniter.NewDecoder(file).Decode(&kindleGlobal)
}

func getFilesInKindle() []os.FileInfo {
	files, err := ioutil.ReadDir("./kindle")
	if err != nil {
		err := os.Mkdir("kindle", 0755)
		if err != nil {
			log.Println("Error creating the kindle directory!")
		}
	}
	return files
}

func checkForNewBooks() ([]Book, error) {
	files := getFilesInKindle()
	books := make([]Book, 0)
	for _, f := range files {
		if f.Name() == ".DS_Store" {
			continue
		}
		//open the file
		file, err := os.Open(newBooksPath + f.Name())
		if err != nil {
			log.Println("Error trying to open kindle file: ", f.Name, " with err: ", err)
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
