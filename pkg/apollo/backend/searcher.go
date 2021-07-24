package backend

import (
	"fmt"
	"math"
	"sort"

	"github.com/amirgamil/apollo/pkg/apollo/schema"
)

//TODO: should search titles too (and put high probability mass on those tokens)

//given a query string a search type (AND / OR ) returns a list of matches ordered by relevance
func Search(query string, searchType string) []schema.Record {
	//1. Gets results of a query
	//keep it in a Go map that acts as a set
	results := make(map[string]bool)
	//2. Apply same analysis as when ingesting data i.e. tokenizing and stemming
	queries := Analyze(query)
	if len(queries) == 0 {
		return make([]schema.Record, 0)
	}
	//Support for AND / OR (TODO: eventually add NOT)
	if searchType == "AND" {
		//3. Get list of relevant records from the invertedIndex
		//temp set holding records we've matched so far for convenience
		//avoid quadratic complexity by sequentially removing records which don't accumulate matches as we move
		//through the queries
		tempRecords := make(map[string]bool)
		//get records for first query
		recordsFirstQueryMatch := globalInvertedIndex[queries[0]]
		for _, recordID := range recordsFirstQueryMatch {
			tempRecords[recordID] = true
		}
		for recordID, _ := range tempRecords {
			record := getRecordFromID(recordID)
			for i := 1; i < len(queries); i++ {
				_, tokenInRecord := record.TokenFrequency[queries[i]]
				if !tokenInRecord {
					//token from our intersection does not exist in this record, so remove it, don't need to keep checking
					delete(tempRecords, recordID)
					break
				}
			}
		}
		//now have all of the records which match all of the queries
		for recordID, _ := range tempRecords {
			results[recordID] = true
		}
	} else if searchType == "OR" {
		//3. Get list of relevant records from the invertedIndex
		for _, query := range queries {
			recordsWithQuery := globalInvertedIndex[query]
			for _, recordID := range recordsWithQuery {
				_, inMap := results[recordID]
				if !inMap {
					results[recordID] = true
				}
			}
		}
	}

	//4. Sory by relevance - assign a score to each record that matches how relevant it is
	//Use the inverse document frequency
	return rank(results, queries)

}

//helper method which return a record from the associated id
func getRecordFromID(id string) schema.Record {
	if id[:2] == "lc" {
		return localRecordList[id]
	} else {
		return sourcesRecordList[id]
	}
}

//idf = log(total number of documents / number of documents that contain term) - ensures tokens which are rarer get a higher score
func idf(token string) float64 {
	return math.Log10(float64(len(localRecordList)+len(sourcesRecordList)) / float64(len(globalInvertedIndex[token])))
}

//ranks an unordered list of records based on relevance, uses the inverse document frequency which is a
//document-level statistic that scores how relevant a document (record in our case) matches our query
//then multiplty by the number of times the token gets mentioned in the token
//returns an ordered list of records from most to least relevant
func rank(results map[string]bool, queries []string) []schema.Record {
	type recordRank struct {
		record schema.Record
		score  float64
	}
	//defining a fixed-size array is faster and more memory efficieny
	rankedResults := make([]schema.Record, len(results))
	unsortedResults := make([]recordRank, len(results))
	i := 0
	for recordID, _ := range results {
		record := getRecordFromID(recordID)
		score := float64(0)
		for _, token := range queries {
			idfVal := idf(token)
			score += idfVal * float64(record.TokenFrequency[token])
		}
		unsortedResults[i] = recordRank{record: record, score: score}
		i += 1
	}
	fmt.Println(len(results), unsortedResults)
	//sort by highest order score to lowest
	sort.Slice(unsortedResults, func(i, j int) bool {
		return unsortedResults[i].score > unsortedResults[j].score
	})

	i = 0
	//put sorted records into needed format and return
	for _, val := range unsortedResults {
		rankedResults[i] = val.record
		i += 1
	}
	return rankedResults
}
