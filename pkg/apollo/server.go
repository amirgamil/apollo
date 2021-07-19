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
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	indexFile, err := os.Open("./static/index.html")
	if err != nil {
		io.WriteString(w, "error reading index")
		return
	}
	defer indexFile.Close()

	io.Copy(w, indexFile)
}

func search(w http.ResponseWriter, r *http.Request) {
	searchQuery := r.FormValue("q")
	w.Header().Set("Content-Type", "application/json")
	fmt.Println(searchQuery)
	results := backend.Search(searchQuery, "AND")
	if results != nil {
		fmt.Println("results : ", results)
		json.NewEncoder(w).Encode(results)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func Start() {
	r := mux.NewRouter()

	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:8993",
		WriteTimeout: 60 * time.Second,
		ReadTimeout:  60 * time.Second,
	}

	//will need to some kind of API call to ingest data
	r.Methods("POST").Path("/search").HandlerFunc(search)
	r.HandleFunc("/", index)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	log.Printf("Server listening on %s\n", srv.Addr)
	log.Fatal(srv.ListenAndServe())

}
