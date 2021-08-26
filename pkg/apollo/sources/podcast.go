package sources

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/amirgamil/apollo/pkg/apollo/schema"
)

type RSS struct {
	Title       string       `xml:"channel>title"`
	Description string       `xml:"channel>description"`
	Episodes    []EpisodeXML `xml:"channel>item"`
}

type EpisodeXML struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
}

//This source pulls data from a personal podcast I host with my friend
//Check it out: https://tinyurl.com/theconversationlab

const newEpisodesPath = "./podcast/"

var podcastsGlobal map[string]schema.Data

//helper method to manage which episodes we've deleted once we've confirmed they've been saved
var episodesToDelete []string

//follows similar approach to Kindle, podcast folder - put new txt files there
func getPodcast() map[string]schema.Data {
	episodesToDelete = make([]string, 0)
	podcastsGlobal = make(map[string]schema.Data)
	rss, err := readXMLFile()
	if err != nil {
		return make(map[string]schema.Data)
	}
	newEpisodes, err := checkForNewEpisodes(rss)
	if err != nil {
		log.Println(err)
		return make(map[string]schema.Data)
	}
	addNewEpisodesToDb(newEpisodes)
	if err != nil {
		log.Println(err)
	} else {
		//use a special delete files in the podcast and not the one in the util
		//since the special case where the text of a podcast content exists in our folders but
		//the RSS feed has not updated and so we don't have access to the podcast metadata. In this case, we cannot delete
		//file because we haven't saved it yet
		for _, episode := range episodesToDelete {
			err := os.Remove(episode)
			if err != nil {
				log.Println("Error deleting podcast transcript file: ", episode, " err")
			}
		}
	}
	return podcastsGlobal
}

func readXMLFile() (RSS, error) {
	resp, err := http.Get("https://media.rss.com/theconversationlab/feed.xml")
	if err != nil {
		log.Println("Error getting the XML file: ", err)
		return RSS{}, err
	}
	defer resp.Body.Close()
	var podcastXML RSS
	err = xml.NewDecoder(resp.Body).Decode(&podcastXML)
	if err != nil {
		log.Println("Error parsing the the XML file: ", err)
		return RSS{}, err
	}
	return podcastXML, nil
}

var episodeNotFound = errors.New("Epsiode not found in the RSS feed!")

func findEpisodeInRSSWithName(name string, rssFeed RSS) (EpisodeXML, error) {
	for _, episode := range rssFeed.Episodes {
		if strings.HasPrefix(episode.Title, name) {
			return episode, nil
		}
	}
	return EpisodeXML{}, episodeNotFound
}

func addNewEpisodesToDb(episodes []schema.Data) {
	regex, _ := regexp.Compile("[0-9]+")
	for _, episode := range episodes {
		//trim any leading or trailing spaces
		episode.Title = strings.Trim(episode.Title, " ")
		episodeNumber := regex.FindString(episode.Title)
		keyInMap := fmt.Sprintf("srpd%s", episodeNumber)
		podcastsGlobal[keyInMap] = episode
	}
}

func checkForNewEpisodes(rssFeed RSS) ([]schema.Data, error) {
	files := getFilesInFolder(newEpisodesPath, "podcast")
	newEpisodes := make([]schema.Data, 0)
	for _, f := range files {
		if f.Name() == ".DS_Store" {
			continue
		}
		regex, _ := regexp.Compile("Episode [0-9]+")
		if regex.MatchString(f.Name()) {
			//grab the episode name e.g. "Episode 1"
			episodeTitle := regex.FindString(f.Name())
			//check for corresponding episode in the RSS feed
			episode, err := findEpisodeInRSSWithName(episodeTitle, rssFeed)
			if err == episodeNotFound {
				//skip and leave file as is, will refresh once the RSS feed shows it
				continue
			} else if err != nil {
				return []schema.Data{}, err
			}
			//confirmed we have an episode and it's in the RSS feed so grab the transcript from the file
			//open the file
			path := newEpisodesPath + f.Name()
			file, err := os.Open(path)
			if err != nil {
				return []schema.Data{}, err
			}
			fileBody, err := ioutil.ReadAll(file)
			if err != nil {
				return []schema.Data{}, err
			}
			transcript := string(fileBody)
			title := fmt.Sprintf("The Conversation Lab - %s", episode.Title)
			newEpisodes = append(newEpisodes, schema.Data{Title: title, Link: episode.Link, Content: transcript, Tags: make([]string, 0)})
			episodesToDelete = append(episodesToDelete, path)

		}
	}
	return newEpisodes, nil
}
