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
	"github.com/amirgamil/apollo/pkg/apollo/schema"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	jsoniter "github.com/json-iterator/go"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

//records which are stored locally, which have been added via Apollo directly
const localRecordsPath = "./data/local.json"

//hold rows of data that have yet to be flushed in the inverted index

//TODO: will need to add some intelligent, heuristic based methods when syncing with other modules to check if it's a link that it gets scraped

var data []schema.Record

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
    (*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
    (*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func index(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}
	indexFile, err := os.Open("./static/index.html")
	if err != nil {
		io.WriteString(w, "error reading index")
		return
	}
	defer indexFile.Close()
	io.Copy(w, indexFile)
}

func scrape(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}
	linkToScraoe := r.FormValue("q")
	w.Header().Set("Content-Type", "application/json")
	result, err := schema.Scrape(linkToScraoe)
	if err != nil {
		log.Fatal("Error trying to scrape a digital artifact!")
		w.WriteHeader(http.StatusExpectationFailed)
	} else {
		json.NewEncoder(w).Encode(result)
	}
}

func addData(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}
	var newData schema.Data
	err := jsoniter.NewDecoder(r.Body).Decode(&newData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		//get the record from the data
		//note all records stored locally on apollo start with prefix lc
		record := backend.GetRecordFromData(newData, fmt.Sprintf("lc%d", len(data)))
		data = append(data, record)
		err = writeRecordToDisk()
		if err != nil {
			w.WriteHeader(http.StatusExpectationFailed)
		} else {
			w.WriteHeader(http.StatusAccepted)
		}
	}
}

func search(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}
	searchQuery := r.FormValue("q")
	w.Header().Set("Content-Type", "application/json")
	fmt.Println(searchQuery)
	results, err := backend.Search(searchQuery, "AND")
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
	} else {
		// fmt.Println("results : ", results)
		jsoniter.NewEncoder(w).Encode(results)
	}
}

func authenticatePassword(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}
	type Request struct {
		Password string `json:"password"`
	}
	var request Request
	json.NewDecoder(r.Body).Decode(&request)
	if isValidPassword(request.Password) {
		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func isValidPassword(password string) bool {
	err := godotenv.Load()
	check(err)
	truePass := os.Getenv("PASSWORD")
	return truePass == password
}

//writes the current cache in memory to disk i.e. saves the database for persistent storage
func writeRecordToDisk() error {
	jsonFile, err := os.OpenFile(localRecordsPath, os.O_WRONLY|os.O_CREATE, 0755)
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
	file, err := os.Open(localRecordsPath)
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
		Addr:         "0.0.0.0:8993",
		WriteTimeout: 60 * time.Second,
		ReadTimeout:  60 * time.Second,
	}

	//will need to some kind of API call to ingest data
	r.Methods("POST").Path("/search").HandlerFunc(search)
	r.Methods("POST").Path("/scrape").HandlerFunc(scrape)
	r.Methods("POST").Path("/addData").HandlerFunc(addData)
	r.Methods("POST").Path("/authenticate").HandlerFunc(authenticatePassword)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	r.PathPrefix("/").HandlerFunc(index)
	log.Printf("Server listening on %s\n", srv.Addr)
	log.Fatal(srv.ListenAndServe())

}
