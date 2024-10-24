package main

import (
	"log"
	"net/http"

	"github.com/suzuki11109/chathub"
)

func main() {
	cs := chathub.NewServer()
	go cs.Run()

	http.Handle("/ws", cs)
	log.Println("listening at localhost:8081")
	http.ListenAndServe("localhost:8081", nil)
}
