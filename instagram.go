package instagram

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var (
	dir string
)

func init() {
	host, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	switch host {
	case "NOTE-GAIDO":
		dir = "C:\\GoPrograms\\i"
	case "lemp":
		dir = "/var/www/html"
	default:
		log.Fatal("host desconhecido")
	}
}

func download(u string) ([]byte, error) {
	res, err := http.Get(u)
	if err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}
	return b, nil
}
