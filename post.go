package instagram

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
)

type SharedData struct {
	EntryData struct {
		PostPage []struct {
			Graphql struct {
				ShortcodeMedia struct {
					Shortcode        string `json:"shortcode"`
					DisplayResources []struct {
						Src          string `json:"src"`
						ConfigWidth  int    `json:"config_width"`
						ConfigHeight int    `json:"config_height"`
					} `json:"display_resources"`
					VideoUrl string `json:"video_url"`
					IsVideo  bool   `json:"is_video"`
					Owner    struct {
						Username string `json:"username"`
					} `json:"owner"`
					EdgeSidecarToChildren struct {
						Edges []struct {
							Node struct {
								Shortcode        string `json:"shortcode"`
								DisplayResources []struct {
									Src          string `json:"src"`
									ConfigWidth  int    `json:"config_width"`
									ConfigHeight int    `json:"config_height"`
								} `json:"display_resources"`
								VideoUrl string `json:"video_url"`
								IsVideo  bool   `json:"is_video"`
							} `json:"node"`
						} `json:"edges"`
					} `json:"edge_sidecar_to_children"`
				} `json:"shortcode_media"`
			} `json:"graphql"`
		} `json:"PostPage"`
	} `json:"entry_data"`
}

func GetPost(u string) (SharedData, error) {
	res, err := http.Get(u)
	if err != nil {
		return SharedData{}, fmt.Errorf("get: %w", err)
	}
	defer res.Body.Close()
	s := bufio.NewScanner(res.Body)
	for s.Scan() {
		l := s.Text()
		if strings.Contains(l, `<script type="text/javascript">window._sharedData = `) {
			l = strings.Replace(l, `<script type="text/javascript">window._sharedData = `, ``, 1)
			l = strings.Replace(l, `;</script>`, ``, 1)
			var sd SharedData
			if err := json.Unmarshal([]byte(l), &sd); err != nil {
				return SharedData{}, fmt.Errorf("get: %w", err)
			}
			return sd, nil
		}
	}
	return SharedData{}, fmt.Errorf("get: shared data não encontrado")
}

func SavePost(u string) ([]string, error) {
	sd, err := GetPost(u)
	if err != nil {
		return nil, fmt.Errorf("savepost: %w", err)
	}
	if len(sd.EntryData.PostPage) != 1 {
		return nil, fmt.Errorf("savepost: len postpage != 1")
	}
	var arquivos []string
	if len(sd.EntryData.PostPage[0].Graphql.ShortcodeMedia.EdgeSidecarToChildren.Edges) == 0 {
		if len(sd.EntryData.PostPage[0].Graphql.ShortcodeMedia.DisplayResources) == 0 {
			return nil, fmt.Errorf("savepost: sem display resources")
		}
		n := len(sd.EntryData.PostPage[0].Graphql.ShortcodeMedia.DisplayResources)
		uMedia := sd.EntryData.PostPage[0].Graphql.ShortcodeMedia.DisplayResources[n-1].Src
		extensao := "jpg"
		if sd.EntryData.PostPage[0].Graphql.ShortcodeMedia.IsVideo {
			uMedia = sd.EntryData.PostPage[0].Graphql.ShortcodeMedia.VideoUrl
			extensao = "mp4"
		}
		nome := fmt.Sprintf(
			"%v-%v.%v",
			sd.EntryData.PostPage[0].Graphql.ShortcodeMedia.Owner.Username,
			sd.EntryData.PostPage[0].Graphql.ShortcodeMedia.Shortcode,
			extensao,
		)
		b, err := download(uMedia)
		if err != nil {
			return nil, fmt.Errorf("savepost: %w", err)
		}
		if err := ioutil.WriteFile(filepath.Join(dir, nome), b, 0644); err != nil {
			return nil, fmt.Errorf("savepost: %w", err)
		}
		return append(arquivos, nome), nil
	}
	for _, edge := range sd.EntryData.PostPage[0].Graphql.ShortcodeMedia.EdgeSidecarToChildren.Edges {
		if len(edge.Node.DisplayResources) == 0 {
			return nil, fmt.Errorf("savepost: sem display resources")
		}
		n := len(edge.Node.DisplayResources)
		uMedia := edge.Node.DisplayResources[n-1].Src
		extensao := "jpg"
		if edge.Node.IsVideo {
			uMedia = edge.Node.VideoUrl
			extensao = "mp4"
		}
		nome := fmt.Sprintf(
			"%v-%v.%v",
			sd.EntryData.PostPage[0].Graphql.ShortcodeMedia.Owner.Username,
			edge.Node.Shortcode,
			extensao,
		)
		b, err := download(uMedia)
		if err != nil {
			return nil, fmt.Errorf("savepost: %w", err)
		}
		if err := ioutil.WriteFile(filepath.Join(dir, nome), b, 0644); err != nil {
			return nil, fmt.Errorf("savepost: %w", err)
		}
		arquivos = append(arquivos, nome)
	}
	return arquivos, nil
}
