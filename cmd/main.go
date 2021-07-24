package main

import (
	"github.com/amirgamil/apollo/pkg/apollo"
	"github.com/amirgamil/apollo/pkg/apollo/backend"
)

func main() {
	backend.InitializeFilesAndData()
	//server and the pipeline should run on concurrent threads, called regularly, for now manually do it
	//start the server on a concurrent thread so that when we need to refresh the inverted index, this happens on
	//different threads
	// backend.RefreshInvertedIndex()
	apollo.Start()
	//two days in miliseconds
	// ticker := time.NewTicker(2 * 24 * 60 * 60 * time.Millisecond)
	// done := make(chan bool)
	// go func() {
	// 	for {
	// 		select {
	// 		case <-done:
	// 			return
	// 		default:
	// 			backend.RefreshInvertedIndex()
	// 		}
	// 	}
	// }()
	// ticker.Stop()
	//once a day, takes all the records, pulls from the data sources,
}
