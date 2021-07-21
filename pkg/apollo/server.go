package apollo

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/amirgamil/apollo/pkg/apollo/backend"
	"github.com/gorilla/mux"
	jsoniter "github.com/json-iterator/go"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

const dbPath = "./data/db.json"

//hold rows of data that have yet to be flushed in the inverted index

//TODO: will need to add some intelligent, heuristic based methods when syncing with other modules to check if it's a link that it gets scraped

var data []backend.Data

func index(w http.ResponseWriter, r *http.Request) {
	indexFile, err := os.Open("./static/index.html")
	if err != nil {
		io.WriteString(w, "error reading index")
		return
	}
	defer indexFile.Close()

	io.Copy(w, indexFile)
}

func scrape(w http.ResponseWriter, r *http.Request) {
	linkToScraoe := r.FormValue("q")
	w.Header().Set("Content-Type", "application/json")
	result, err := backend.Scrape(linkToScraoe)
	if err != nil {
		log.Fatal("Error trying to parse an article!")
		w.WriteHeader(http.StatusExpectationFailed)
	} else {
		json.NewEncoder(w).Encode(result)
	}
}

func addData(w http.ResponseWriter, r *http.Request) {
	var newData backend.Data
	err := jsoniter.NewDecoder(r.Body).Decode(&newData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		data = append(data, newData)
		err = writeDataToDisk()
		if err != nil {
			w.WriteHeader(http.StatusExpectationFailed)
		} else {
			w.WriteHeader(http.StatusAccepted)
		}
	}
}

func search(w http.ResponseWriter, r *http.Request) {
	searchQuery := r.FormValue("q")
	w.Header().Set("Content-Type", "application/json")
	fmt.Println(searchQuery)
	results := backend.Search(searchQuery, "AND")
	if results != nil {
		fmt.Println("results : ", results)
		jsoniter.NewEncoder(w).Encode(results)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

//writes the current cache in memory to disk i.e. saves the database for persistent storage
func writeDataToDisk() error {
	jsonFile, err := os.OpenFile(dbPath, os.O_WRONLY|os.O_CREATE, 0755)
	defer jsonFile.Close()
	//error may occur when reading from an empty file for the first time
	if err != nil {
		return err
	}
	err = jsoniter.NewEncoder(jsonFile).Encode(data)
	if err != nil {
		return err
	}
	return nil
}

func loadData() {
	file, err := os.Open(dbPath)
	if err != nil {
		fmt.Println("Error loading the database with new data!")
	}
	err = jsoniter.NewDecoder(file).Decode(&data)
	if err != nil {
		fmt.Println("Error parsing data into JSON")
	}
}

func Start() {
	r := mux.NewRouter()
	loadData()
	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:8993",
		WriteTimeout: 60 * time.Second,
		ReadTimeout:  60 * time.Second,
	}

	//will need to some kind of API call to ingest data
	r.Methods("POST").Path("/search").HandlerFunc(search)
	r.Methods("POST").Path("/scrape").HandlerFunc(scrape)
	r.Methods("POST").Path("/addData").HandlerFunc(addData)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	r.PathPrefix("/").HandlerFunc(index)
	log.Printf("Server listening on %s\n", srv.Addr)
	log.Fatal(srv.ListenAndServe())

}
