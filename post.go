package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
)

type Post struct {
	Tipo      string
	Id        string
	ShortCode string
	Dimensoes struct {
		Altura  int
		Largura int
	}
	VideoUrl              string
	Video                 bool
	Url                   string
	Legenda               string
	LegendaAcessibilidade string
	Timestamp             int
	Perfil
	Posts []Post
}

func download(p Post, dirname string) error {
	var (
		u   string
		ext string
	)
	if p.Video {
		u = p.VideoUrl
		ext = ".mp4"
	} else {
		u = p.Url
		ext = ".jpg"
	}
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	filename := filepath.Join(dirname, p.Perfil.UserName+" - "+p.ShortCode+ext)
	return ioutil.WriteFile(filename, b, 0644)
}

func (p Post) Download(dirname string) error {
	if len(p.Posts) == 0 {
		return download(p, dirname)
	}
	for _, post := range p.Posts {
		if err := download(post, dirname); err != nil {
			return err
		}
	}
	return nil
}

type shortcodeMedia struct {
	TypeName   string `json:"__typename"`
	Id         string `json:"id"`
	Shortcode  string `json:"shortcode"`
	Dimensions struct {
		Height int `json:"height"`
		Width  int `json:"width"`
	} `json:"dimensions"`
	DisplayUrl           string `json:"display_url"`
	AccessibilityCaption string `json:"accessibility_caption"`
	VideoUrl             string `json:"video_url"`
	IsVideo              bool   `json:"is_video"`
	EdgeMediaToCaption   struct {
		Edges []struct {
			Node struct {
				Text string `json:"text"`
			} `json:"node"`
		} `json:"edges"`
	} `json:"edge_media_to_caption"`
	TakenAtTimestamp int `json:"taken_at_timestamp"`
	Owner            struct {
		Id            string `json:"id"`
		ProfilePicUrl string `json:"profile_pic_url"`
		Username      string `json:"username"`
		FullName      string `json:"full_name"`
	} `json:"owner"`
	EdgeSidecarToChildren struct {
		Edges []struct {
			Node shortcodeMedia `json:"node"`
		} `json:"edges"`
	} `json:"edge_sidecar_to_children"`
}

type sharedDataPost struct {
	EntryData struct {
		PostPage []struct {
			Graphql struct {
				ShortcodeMedia shortcodeMedia `json:"shortcode_media"`
			} `json:"graphql"`
		} `json:"PostPage"`
	} `json:"entry_data"`
	RhxGis string `json:"rhx_gis"`
}

func PostCarrega(shortcode string) (Post, error) {
	res, err := http.Get("https://www.instagram.com/p/" + shortcode + "/")
	if err != nil {
		return Post{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return Post{}, errors.New(fmt.Sprintf("carrega post: %v", res.Status))
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return Post{}, err
	}
	matches := reSharedData.FindStringSubmatch(string(b))
	if len(matches) != 2 {
		return Post{}, errors.New("sharedData de post não encontrado")
	}
	var sdp sharedDataPost
	if err := json.Unmarshal([]byte(matches[1]), &sdp); err != nil {
		return Post{}, err
	}
	if len(sdp.EntryData.PostPage) != 1 {
		return Post{}, errors.New("sharedData não tem exatamente um PostPage")
	}
	media := sdp.EntryData.PostPage[0].Graphql.ShortcodeMedia
	post := Post{
		Tipo:      media.TypeName,
		Id:        media.Id,
		ShortCode: media.Shortcode,
		Dimensoes: struct {
			Altura  int
			Largura int
		}{Altura: media.Dimensions.Height, Largura: media.Dimensions.Height},
		VideoUrl:              media.VideoUrl,
		Video:                 media.IsVideo,
		Url:                   media.DisplayUrl,
		LegendaAcessibilidade: media.AccessibilityCaption,
		Timestamp:             media.TakenAtTimestamp,
		Perfil: Perfil{
			Id:       media.Owner.Id,
			Foto:     media.Owner.ProfilePicUrl,
			UserName: media.Owner.Username,
			Nome:     media.Owner.FullName,
		},
	}
	if len(media.EdgeMediaToCaption.Edges) > 0 {
		post.Legenda = media.EdgeMediaToCaption.Edges[0].Node.Text
	}
	for _, edge := range media.EdgeSidecarToChildren.Edges {
		node := edge.Node
		p := Post{
			Tipo:      node.TypeName,
			Id:        node.Id,
			ShortCode: node.Shortcode,
			Dimensoes: struct {
				Altura  int
				Largura int
			}{Altura: node.Dimensions.Height, Largura: node.Dimensions.Width},
			VideoUrl:              node.VideoUrl,
			Video:                 node.IsVideo,
			Url:                   node.DisplayUrl,
			LegendaAcessibilidade: node.AccessibilityCaption,
			Timestamp:             node.TakenAtTimestamp,
			Perfil: Perfil{
				Id:       media.Owner.Id,
				Foto:     media.Owner.ProfilePicUrl,
				UserName: media.Owner.Username,
				Nome:     media.Owner.FullName,
			},
		}
		if len(node.EdgeMediaToCaption.Edges) > 0 {
			p.Legenda = node.EdgeMediaToCaption.Edges[0].Node.Text
		}
		post.Posts = append(post.Posts, p)
	}
	return post, nil
}
