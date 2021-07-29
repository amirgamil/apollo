package apollo

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
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
	result, err := schema.Scrape(linkToScraoe)
	if err != nil {
		log.Fatal("Error trying to scrape a digital artifact!")
		w.WriteHeader(http.StatusExpectationFailed)
	} else {
		json.NewEncoder(w).Encode(result)
	}
}

func addData(w http.ResponseWriter, r *http.Request) {
	var newData schema.Data
	err := jsoniter.NewDecoder(r.Body).Decode(&newData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		backend.AddNewEntryToLocalData(newData)
		w.WriteHeader(http.StatusAccepted)
	}
}

func search(w http.ResponseWriter, r *http.Request) {
	searchQuery := r.FormValue("q")
	//add support here for older browsers?
	if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip, deflate") {
		return
	}
	w.Header().Set("Content-Encoding", "gzip")
	// w.Header().Set("Content-Type", "application/json")
	fmt.Println(searchQuery)
	//TODO: add logic for OR
	results, err := backend.Search(searchQuery, "AND")

	if err != nil {
		w.WriteHeader(http.StatusNoContent)
	} else {
		// fmt.Println("results : ", results)
		gz := gzip.NewWriter(w)
		defer gz.Close()
		jsoniter.NewEncoder(gz).Encode(results)
	}
}

func authenticatePassword(w http.ResponseWriter, r *http.Request) {
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

func Start() {
	r := mux.NewRouter()
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
