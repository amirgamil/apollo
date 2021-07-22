package sources

import (
	"errors"
	"json"
	"log"
	"os"

	"github.com/amirgamil/apollo/pkg/apollo/backend"
)

const athenaPath = "./athena/data.json"

type thought struct {
	H string   `json:"h"`
	B string   `json:"b"`
	T []string `json: "t"`
}

func getAthena() []backend.Data {
	data, err := loadData()
	if err != nil {
		log.Fatal(err)
		return []backend.Data{}
	}
	dataToIndex := convertToReqFormat(data)
	return dataToIndex
}

func loadAthenaData() ([]thought, error) {
	var data []thought
	file, err := os.Open(athenaPath)
	if err != nil {
		return []thought{}, errors.New("Error loading data from Athena!")
	}
	json.NewDecoder(file).Decode(&data)
	return data, nil
}

//takes a lists of thoughts and converts it into the require data struct we need for the api
func convertToReqFormat(data []thought) []backend.Data {
	dataToIndex := make([]backend.Data, len(data))
	for i, thought := range data {
		dataToIndex[i] = backend.Data{Title: thought.H, Content: thought.B, Link: "https://athena.amirbolous.com", Tags: thought.T}
	}
	return dataToIndex
}
