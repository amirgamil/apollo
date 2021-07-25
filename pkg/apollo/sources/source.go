package sources

import (
	"github.com/amirgamil/apollo/pkg/apollo/schema"
)

func GetData() []schema.Data {
	data := make([]schema.Data, 0)
	athena := getAthena()
	data = append(data, athena...)
	zeus := getZeus()
	data = append(data, zeus...)
	return data
}
