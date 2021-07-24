package schema

import (
	"errors"
	"os/exec"
	"time"

	readability "github.com/go-shiori/go-readability"
)

func Scrape(link string) (Data, error) {
	article, err := readability.FromURL(link, 30*time.Second)
	if err != nil {
		return Data{}, err
	}
	return Data{Title: article.Title, Link: link, Content: article.TextContent, Tags: make([]string, 0)}, nil
}

func HandleYouTubeVideo(link string) error {
	command := "youtube-dl --skip-download --write-auto-sub " + link
	cmd := exec.Command("bash", "-c", command)
	err := cmd.Run()
	if err != nil {
		return errors.New("Error downloading the youtube video!")
	}
	return nil
}

func downloadYTSubtitles() {

}

func readFromSubtitlesFile() {

}
