package main

import (
	"log"
	"net/http"

	"github.com/nytlabs/st-core/server"
)

func main() {

	log.Println("serving on 7071")

	err := http.ListenAndServe(":7071", server.NewServer().NewRouter())
	if err != nil {
		log.Panicf(err.Error())
	}

}
