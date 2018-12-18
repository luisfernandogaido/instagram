package main

import (
	"log"
)

func main() {
	perfil, err := PerfilCarrega("https://www.instagram.com/mongodb")
	if err != nil {
		log.Fatal(err)
	}
	for _, n := range perfil.Nodes {
		post, err := PostCarrega(n.Shortcode)
		if err != nil {
			log.Fatal(err)
		}
		if err := post.Download("D:\\Users\\81092610\\Desktop\\posts"); err != nil {
			log.Fatal(err)
		}
	}
}
