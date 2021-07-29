package sources

import (
	"io/ioutil"
	"log"
	"os"
)

func ensureFileExists(path string) {
	jsonFile, err := os.Open(path)
	if err != nil {
		file, err := os.Create(path)
		if err != nil {
			log.Println("Error creating the original kindle database: ", err)
			return
		}
		file.Close()
	} else {
		defer jsonFile.Close()
	}
}

func getFilesInFolder(path string, folderName string) []os.FileInfo {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		err := os.Mkdir(folderName, 0755)
		if err != nil {
			log.Println("Error creating the kindle directory!")
		}
	}
	return files
}

func deleteFiles(path string, folderName string) {
	files := getFilesInFolder(path, folderName)
	for _, f := range files {
		err := os.Remove(path + f.Name())
		if err != nil {
			log.Println("Error deleting file: ", f.Name())
		}
	}
}
