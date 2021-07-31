package schema

//smallest unit of data that we store in the database
//this will store each "item" in our search engine with all of the necessary information
//for the interverted index
type Record struct {
	//unique identifier
	ID string `json:"id"`
	//title
	Title string `json:"title"`
	//potential link to the source if applicable
	Link string `json:"link"`
	//text content to display on results page
	Content string `json:"content"`
	//map of tokens to their frequency
	TokenFrequency map[string]int `json:"tokenFrequency"`
}

type SearchResult struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Content string `json:"content"`
}

//represents raw data that we will parse objects into before they have been transformed into records
//and stored in our database
type Data struct {
	Title   string   `json:"title"`
	Link    string   `json:"link"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
	//TODO: add metadata, should be able to search based on type of record, document, podcast, personal etc.
}

//how we send back the result of a search query to the client
type Payload struct {
	Time   int64          `json:"time"`
	Length int            `json:"length"`
	Query  []string       `json:"query"`
	Data   []SearchResult `json:"data"`
}
