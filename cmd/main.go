package main

import (
	"log"
	"time"
)

func main() {
	for {
		err := processaPosts()
		if err != nil {
			log.Println(err)
		}
		time.Sleep(time.Second * 5)
	}
}
