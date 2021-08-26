package backend

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/amirgamil/apollo/pkg/apollo/schema"
)

//TODO: should search titles too (and put high probability mass on those tokens)

//given a query string a search type (AND / OR ) returns a list of matches ordered by relevance
func Search(query string, searchType string, currentSearchResults map[string]string) (schema.Payload, error) {
	//1. Gets results of a query
	//keep it in a Go map that acts as a set
	startTime := time.Now()
	results := make(map[string]bool)
	//2. Apply same analysis as when ingesting data i.e. tokenizing and stemming
	queries := Analyze(query)
	if len(queries) == 0 {
		return schema.Payload{}, errors.New("No valid queries!")
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
	records := rank(results, queries, currentSearchResults)
	//convert searched time to miliseconds
	time := int64(time.Now().Sub(startTime))
	return schema.Payload{Time: time, Data: records, Query: queries, Length: len(records)}, nil

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
func rank(results map[string]bool, queries []string, currentSearchResults map[string]string) []schema.SearchResult {
	type recordRank struct {
		result schema.SearchResult
		score  float64
	}
	//defining a fixed-size array is faster and more memory efficieny
	rankedResults := make([]schema.SearchResult, len(results))
	unsortedResults := make([]recordRank, len(results))
	i := 0
	queriesChained := strings.Join(queries, " ")
	fmt.Println(queriesChained)
	regex, _ := regexp.Compile(queriesChained)
	for recordID, _ := range results {
		record := getRecordFromID(recordID)
		score := float64(0)
		for _, token := range queries {
			idfVal := idf(token)
			score += idfVal * float64(record.TokenFrequency[token])
		}
		content := getSurroundingText(regex, record.Content)
		fmt.Println(strings.ReplaceAll(record.Title, " ", "!"))
		//add regex highlighted of the full content which is readily available when a user clicks on an item to view details
		//this way, we don't need to every single record's contents and can speed up searches
		currentSearchResults[record.Title] = regex.ReplaceAllString(record.Content, fmt.Sprintf(`<span class="highlighted">%s</span>`, queriesChained))
		unsortedResults[i] = recordRank{result: schema.SearchResult{Title: record.Title, Link: record.Link, Content: content}, score: score}
		i += 1
	}
	//sort by highest order score to lowest
	sort.Slice(unsortedResults, func(i, j int) bool {
		return unsortedResults[i].score > unsortedResults[j].score
	})

	i = 0
	//put sorted records into needed format and return
	for _, val := range unsortedResults {
		rankedResults[i] = val.result
		i += 1
	}
	return rankedResults
}

//helper method to get small window of matching result
//don't send the full text back to the client cause this is too slow
func getSurroundingText(regexp *regexp.Regexp, content string) string {
	indices := regexp.FindStringIndex(strings.ToLower(content))
	//TODO? make greedy? match different variations
	//if we find no match, then we've matched a token that's stem is not included
	//in the actual text, so just return the first section
	if indices == nil {
		if len(content) > 150 {
			return content[:150]
		}
		return content
	}
	//want to get a small window with the highlighted content 100 characters on each side
	start := indices[0] - 15
	end := indices[1] + 100
	if start < 0 && end >= len(content) {
		//if the entire content is smaller than the window, then just display all of the content
		start = 0
		end = len(content)
	} else if start < 0 {
		//if the match is nearer to the front, shift the window "to the right" and display more on tailend
		start = 0
	} else if end >= len(content) {
		//if the match is nearer to the end, shift the window "to the left" and display more on the front
		end = len(content)
	}
	return content[start:end]
}
