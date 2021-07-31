package schema

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	readability "github.com/go-shiori/go-readability"
)

func Scrape(link string) (Data, error) {
	log.Println(strings.Contains(link, "www.youtube.com"))
	//handle YouTube videos
	if strings.Contains(link, "www.youtube.com") {
		return HandleYouTubeVideo(link)
	}
	article, err := readability.FromURL(link, 30*time.Second)
	//add goquery and if it fails, return Text()?
	if err != nil {
		return Data{}, err
	}
	regex, _ := regexp.Compile("(<[^>]+>)")
	cleanContent := regex.ReplaceAllString(article.TextContent, "")
	return Data{Title: article.Title, Link: link, Content: cleanContent, Tags: make([]string, 0)}, nil
}

func HandleYouTubeVideo(link string) (Data, error) {
	command := "youtube-dl -o '%(title)s' --write-srt --sub-lang en --skip-download " + link
	cmd := exec.Command("bash", "-c", command)
	err := cmd.Run()
	var out bytes.Buffer
	cmd.Stdout = &out
	log.Println(out.String())
	if err != nil {
		log.Println("Error running bash script: ", err)
		return Data{}, errors.New("Error downloading the youtube video!")
	}
	content, title, err := readFromSubtitlesFile()
	if err != nil {
		return Data{}, errors.New("Error loading the subtitles of the video!")
	}
	return Data{Title: title, Link: link, Content: content, Tags: make([]string, 0)}, nil
}

func readFromSubtitlesFile() (string, string, error) {
	files, err := ioutil.ReadDir("./")
	if err != nil {
		log.Println("Error reading the files of the YouTube video: ", err)
		return "", "", nil
	}
	for _, file := range files {
		//find the vtt file which is the format of the downloaded subtitles
		r, err := regexp.MatchString(".vtt", file.Name())
		if err == nil && r {
			//found file
			output, err := readVTTFile(file.Name())
			removeVTTFille(file.Name())
			if err != nil {
				return "", "", err
			}
			regexTitle, _ := regexp.Compile(`(\.en)?\.vtt`)
			title := regexTitle.ReplaceAllString(file.Name(), "")
			return output, title, nil
		}
	}
	return "", "", errors.New("Error reading subtitles!")
}

func removeVTTFille(name string) error {
	err := os.Remove(name)
	if err != nil {
		log.Println("Error removing file: ", err)
		return err
	}
	return nil
}

func readVTTFile(fileName string) (string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Println("Error opening the VTT file: ", err)
		return "", err
	}
	output, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println("Error reading the VTT file: ", err)
		return "", err
	}
	//rule to remove everything but the text of the vtt file
	regexRule := "\n?([0-9]+):([0-9]+):([0-9]+).([0-9]+) --> ([0-9]+):([0-9]+):([0-9]+).([0-9]+)"
	r, _ := regexp.Compile(regexRule)
	if r.Match(output) {
		textVideo := r.ReplaceAllString(string(output), "")
		textVideo = strings.ReplaceAll(textVideo, "WEBVTT\nKind: captions\nLanguage: en", "")
		return textVideo, nil
	} else {
		log.Println("Error trying to match regex: ")
		return "", err
	}
}
