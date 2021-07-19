package main

import (
	"github.com/amirgamil/apollo/pkg/apollo"
	"github.com/amirgamil/apollo/pkg/apollo/backend"
)

func main() {
	backend.InitializeFilesAndData()
	//server and the pipeline should run on concurrent threads, called regularly, for now manually do it
	backend.FlushNewDataIntoInvertedIndex()
	apollo.Start()
}
