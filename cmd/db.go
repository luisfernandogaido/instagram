package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"

	"github.com/luisfernandogaido/instagram"
)

var (
	db *sql.DB
)

func init() {
	var (
		err error
		dsn string
	)
	host, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	switch host {
	case "NOTE-GAIDO":
		dsn = "root:Semaver13@/gaidodev"
	case "lemp":
		dsn = "root:1000sonhosreais@/gaidodev"
	default:
		log.Fatal("host desconhecido")
	}
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	if err := db.Ping(); err != nil {
		panic(err)
	}
}

type Post struct {
	Codigo int
	Url    string
}

func postsNaoProcessados() ([]Post, error) {
	query := `
		SELECT codigo, url
		FROM i_post
		WHERE processamento IS NULL
		ORDER BY codigo
`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("postsnaoprocessados: %w", err)
	}
	var (
		post  Post
		posts []Post
	)
	for rows.Next() {
		if err := rows.Scan(&post.Codigo, &post.Url); err != nil {
			return nil, fmt.Errorf("postsnaoprocessados: %w", err)
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func processaPost(p Post, medias []string, errSalvamento error) error {
	var erro sql.NullString
	if errSalvamento != nil {
		erro = sql.NullString{
			String: errSalvamento.Error(),
			Valid:  true,
		}
	}
	query := `
		UPDATE i_post SET
		processamento = NOW(),
		error = ?
		WHERE url = ?
	`
	_, err := db.Exec(query, erro, p.Url)
	if err != nil {
		return fmt.Errorf("processapost: %w", err)
	}
	if errSalvamento == nil {
		query = `
		INSERT INTO i_media
		(cod_post, file)
		VALUES
		(?, ?)`
		for _, m := range medias {
			_, err = db.Exec(query, p.Codigo, m)
			if err != nil {
				return fmt.Errorf("processapost: %w", err)
			}
		}
	}
	return nil
}

func processaPosts() error {
	posts, err := postsNaoProcessados()
	if err != nil {
		return fmt.Errorf("processaposts: %w", err)
	}
	for _, post := range posts {
		medias, err := instagram.SavePost(post.Url)
		err = processaPost(post, medias, err)
		if err != nil {
			return fmt.Errorf("processaposts: %w", err)
		}
	}
	return nil
}
