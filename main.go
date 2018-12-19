package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const (
	dirname = "./data"
)

func main() {
	usernames := make(map[string]bool)
	b, err := ioutil.ReadFile("./in.txt")
	if err != nil {
		log.Fatal(err)
	}
	links := strings.Split(string(b), "\r\n")
	out, err := os.Create("./out.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	os.Mkdir(dirname, 0755)
	for i, link := range links {
		index := strings.Index(link, "?")
		if index == -1 {
			index = len(link)
		}
		link = strings.Replace(link[0:index], "//instagram", "//www.instagram", 1)
		if usernames[link] {
			continue
		}
		usernames[link] = true
		fmt.Printf("Perfil: %v. %v de %v.\n", link, i+1, len(links))
		out.WriteString(link + "\r\n")
		perfil, err := PerfilCarrega(link)
		if _, ok := usernames[perfil.UserName]; ok {
			continue
		}
		usernames[perfil.UserName] = true
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Username: %v, nodes: %v\n", perfil.UserName, len(perfil.Nodes))
		for i, node := range perfil.Nodes {
			fmt.Println(i, node.Shortcode)
			post, err := PostCarrega(node.Shortcode)
			if err != nil {
				log.Fatal(err)
			}
			post.Download(dirname)
		}
	}
}
