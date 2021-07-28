package sources

import (
	"github.com/amirgamil/apollo/pkg/apollo/schema"
)

var sources map[string]schema.Record

//TODO: make sourcesMap a global so we don't keep passing large maps in parameters
//TODO: should return map[string]schema.Data so we have control over the IDs
func GetData(sourcesMap map[string]schema.Record) map[string]schema.Data {
	sources = sourcesMap
	//pass in number of sources
	sourcesNewData := make([]map[string]schema.Data, 4)
	data := make(map[string]schema.Data)
	athena := getAthena()
	sourcesNewData[0] = athena
	zeus := getZeus()
	sourcesNewData[1] = zeus
	kindle := getKindle()
	sourcesNewData[2] = kindle
	podcast := getPodcast()
	sourcesNewData[3] = podcast
	//add all data
	for _, sourceData := range sourcesNewData {
		for ID, newData := range sourceData {
			data[ID] = newData
		}
	}
	return data
}
