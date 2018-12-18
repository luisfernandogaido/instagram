package main

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

const (
	tentativas = 30
)

var (
	reSharedData    = regexp.MustCompile(`<script type="text/javascript">window._sharedData = (.*);</script>`)
	rePageContainer = regexp.MustCompile(`/static/bundles/metro/ProfilePageContainer\.js/([^.]+)\.js`)
	reQueryId       = regexp.MustCompile(`s.pagination},queryId:"([^"]+)"`)
)

//todo tornar types intermediários privados e ver se o JSON quebra. deixar púclico somente o filão.

type Node struct {
	TypeName           string `json:"__typename"`
	Id                 string `json:"id"`
	Shortcode          string `json:"shortcode"`
	EdgeMediaToComment struct {
		Count int `json:"count"`
	} `json:"edge_media_to_comment"`
	DisplayUrl  string `json:"display_url"`
	EdgeLikedBy struct {
		Count int `json:"count"`
	} `json:"edge_liked_by"`
	ThumbnailSrc     string `json:"thumbnail_src"`
	IsVideo          bool   `json:"is_video"`
	VideoViewCount   int    `json:"video_view_count"`
	TakenAtTimestamp int    `json:"taken_at_timestamp"`
}

type EdgeOwnerToTimelineMedia struct {
	Count    int `json:"count"`
	PageInfo struct {
		HasNextPage bool   `json:"has_next_page"`
		EndCursor   string `json:"end_cursor"`
	} `json:"page_info"`
	Edges []struct {
		Node `json:"node"`
	}
}

type SharedData struct {
	EntryData struct {
		ProfilePage []struct {
			Graphql struct {
				User struct {
					Biography      string `json:"biography"`
					EdgeFollowedBy struct {
						Count int `json:"count"`
					} `json:"edge_followed_by"`
					EdgeFollow struct {
						Count int `json:"count"`
					} `json:"edge_follow"`
					FullName                 string `json:"full_name"`
					Id                       string `json:"id"`
					ProfilePicUrlHd          string `json:"profile_pic_url_hd"`
					Username                 string `json:"username"`
					EdgeOwnerToTimelineMedia `json:"edge_owner_to_timeline_media"`
				} `json:"user"`
			} `json:"graphql"`
		} `json:"ProfilePage"`
	} `json:"entry_data"`
	RhxGis string `json:"rhx_gis"`
}

type QueryHashResponse struct {
	Data struct {
		User struct {
			EdgeOwnerToTimelineMedia `json:"edge_owner_to_timeline_media"`
		} `json:"user"`
	} `json:"data"`
	Status string `json:"status"`
}

type Perfil struct {
	Biografia     string
	Seguidores    int
	Seguindo      int
	Nome          string
	Id            string
	UserName      string
	Foto          string
	Publicacoes   int
	TemMaisPagina bool
	FinalCursor   string
	QueryHash     string
	Nodes         []Node
}

type Variables struct {
	Id    string `json:"id"`
	First int    `json:"first"`
	After string `json:"after"`
}

func PerfilCarrega(u string) (Perfil, error) {
	res, err := http.Get(u)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return Perfil{}, err
	}
	conteudo := string(b)
	matches := reSharedData.FindStringSubmatch(conteudo)
	if len(matches) != 2 {
		return Perfil{}, errors.New("sharedData não encontrado no perfil")
	}
	var sd SharedData
	if err := json.Unmarshal([]byte(matches[1]), &sd); err != nil {
		return Perfil{}, err
	}
	if len(sd.EntryData.ProfilePage) != 1 {
		return Perfil{}, errors.New("sharedData não tem exatamente um ProfilePage")
	}
	perfil := Perfil{
		Biografia:     sd.EntryData.ProfilePage[0].Graphql.User.Biography,
		Seguidores:    sd.EntryData.ProfilePage[0].Graphql.User.EdgeFollowedBy.Count,
		Seguindo:      sd.EntryData.ProfilePage[0].Graphql.User.EdgeFollow.Count,
		Nome:          sd.EntryData.ProfilePage[0].Graphql.User.FullName,
		Id:            sd.EntryData.ProfilePage[0].Graphql.User.Id,
		UserName:      sd.EntryData.ProfilePage[0].Graphql.User.Username,
		Foto:          sd.EntryData.ProfilePage[0].Graphql.User.ProfilePicUrlHd,
		Publicacoes:   sd.EntryData.ProfilePage[0].Graphql.User.EdgeOwnerToTimelineMedia.Count,
		TemMaisPagina: sd.EntryData.ProfilePage[0].Graphql.User.EdgeOwnerToTimelineMedia.PageInfo.HasNextPage,
		FinalCursor:   sd.EntryData.ProfilePage[0].Graphql.User.EdgeOwnerToTimelineMedia.PageInfo.EndCursor,
		Nodes:         make([]Node, 0),
	}
	matches = rePageContainer.FindStringSubmatch(conteudo)
	if len(matches) != 2 {
		return Perfil{}, errors.New("arquivo ProfilePageContainer não encontrado")
	}
	res2, err := http.Get("https://www.instagram.com" + matches[0])
	if err != nil {
		return Perfil{}, err
	}
	defer res2.Body.Close()
	b2, err := ioutil.ReadAll(res2.Body)
	if err != nil {
		return Perfil{}, err
	}
	matches = reQueryId.FindStringSubmatch(string(b2))
	if len(matches) != 2 {
		return Perfil{}, errors.New("query_hash não encontrado em ProfilePageContainer")
	}
	perfil.QueryHash = matches[1]
	edgeOwnerToTimelineMedia := sd.EntryData.ProfilePage[0].Graphql.User.EdgeOwnerToTimelineMedia
	for {
		for _, node := range edgeOwnerToTimelineMedia.Edges {
			perfil.Nodes = append(perfil.Nodes, node.Node)
		}
		if !edgeOwnerToTimelineMedia.PageInfo.HasNextPage {
			break
		}
		variables := Variables{
			Id:    perfil.Id,
			First: 12,
			After: edgeOwnerToTimelineMedia.PageInfo.EndCursor,
		}
		jVariables, err := json.Marshal(variables)
		if err != nil {
			return Perfil{}, err
		}
		values := url.Values{
			"query_hash": []string{perfil.QueryHash},
			"variables":  []string{string(jVariables)},
		}
		req, err := http.NewRequest("GET", "https://www.instagram.com/graphql/query/?"+values.Encode(), nil)
		if err != nil {
			return Perfil{}, err
		}
		signature := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%v:%v", sd.RhxGis, string(jVariables)))))
		req.Header.Set("x-instagram-gis", signature)
		b = b[:0]
		tries := 0
		for {
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				return Perfil{}, err
			}
			if res.StatusCode != 200 {
				tries++
				if tries == tentativas {
					return Perfil{}, errors.New(
						fmt.Sprintf(
							"%v tentativas consecutivas de carregamento de mídias perfil sem sucesso",
							tentativas,
						),
					)
				}
				time.Sleep(time.Second * 4)
				continue
			}
			b, err = ioutil.ReadAll(res.Body)
			if err != nil {
				return Perfil{}, err
			}
			break
		}
		var qhr QueryHashResponse
		if err := json.Unmarshal(b, &qhr); err != nil {
			return Perfil{}, err
		}
		edgeOwnerToTimelineMedia = qhr.Data.User.EdgeOwnerToTimelineMedia
		//todo remover carregamento de apenas duas páginas de mídias de perfil
		//edgeOwnerToTimelineMedia.PageInfo.HasNextPage = false
		res.Body.Close()
		time.Sleep(time.Second * 4)
	}
	return perfil, nil
}
