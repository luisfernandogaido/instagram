package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

type SharedData struct {
	EntryData struct {
		PostPage []struct {
			GraphSQL struct {
				ShortcodeMedia struct {
					Id                    string `json:"id"`
					DisplayURL            string `json:"display_url"`
					VideoURL              string `json:"video_url"`
					EdgeSidecarToChildren struct {
						Edges []struct {
							Node struct {
								DisplayURL string `json:"display_url"`
								VideoURL   string `json:"video_url"`
							} `json:"node"`
						} `json:"edges"`
					} `json:"edge_sidecar_to_children"`
				} `json:"shortcode_media"`
			} `json:"graphql"`
		} `json:"PostPage"`
	} `json:"entry_data"`
}

type TrelloExport struct {
	Actions []struct {
		Data struct {
			Card struct {
				Name string `json:"name"`
			} `json:"card"`
		} `json:"data"`
	} `json:"actions"`
}

func main() {
	if _, err := os.Stat("./json.json"); !os.IsNotExist(err) {
		bytes, err := ioutil.ReadFile("./json.json")
		if err != nil {
			log.Fatal(err)
		}
		te := TrelloExport{}
		if err := json.Unmarshal(bytes, &te); err != nil {
			log.Fatal(err)
		}
		out := ""
		for _, c := range te.Actions {
			if strings.HasPrefix(c.Data.Card.Name, "https://instagram.com/") {
				out += c.Data.Card.Name + "\r\n"
			}
		}
		if err := ioutil.WriteFile("./links.txt", []byte(out), 0664); err != nil {
			log.Fatal(err)
		}
		return
	}
	bytes, err := ioutil.ReadFile("./i.txt")
	if err != nil {
		log.Fatal(err)
	}
	urls := strings.Split(string(bytes), "\r\n")
	for k, url := range urls {
		if err := pegaMidia(url); err != nil {
			log.Println(err)
		}
		fmt.Println(float64(k+1) / float64(len(urls)))
	}
}

func pegaMidia(endereco string) error {
	res, err := http.Get(endereco)
	if err != nil {
		return err
	}
	bytes, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	html := string(bytes)
	iniSharedData := `<script type="text/javascript">window._sharedData = `
	fimSharedData := `;</script>`
	p := strings.Index(html, iniSharedData)
	q := strings.Index(html[p:], fimSharedData)
	sharedData := strings.Replace(html[p:p+q], iniSharedData, "", 1)
	sd := SharedData{}
	if err := json.Unmarshal([]byte(sharedData), &sd); err != nil {
		return err
	}
	sm := sd.EntryData.PostPage[0].GraphSQL.ShortcodeMedia
	enderecosDownload := make([]string, 0)
	if len(sm.EdgeSidecarToChildren.Edges) == 0 {
		if sm.VideoURL != "" {
			enderecosDownload = append(enderecosDownload, sm.VideoURL)
		} else {
			enderecosDownload = append(enderecosDownload, sm.DisplayURL)
		}
	} else {
		for _, edge := range sm.EdgeSidecarToChildren.Edges {
			if edge.Node.VideoURL != "" {
				enderecosDownload = append(enderecosDownload, edge.Node.VideoURL)
			} else {
				enderecosDownload = append(enderecosDownload, edge.Node.DisplayURL)
			}
		}
	}
	for _, enderecoDownload := range enderecosDownload {
		nome := path.Base(enderecoDownload)
		res2, err := http.Get(enderecoDownload)
		if err != nil {
			return err
		}
		bytes, err = ioutil.ReadAll(res2.Body)
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(nome, bytes, 0664); err != nil {
			return err
		}
		res2.Body.Close()
	}
	return nil
}
