package main

import (
	"copilots-world-map/countrydb"
	"copilots-world-map/server"
	"flag"
	"log"
	"net/http"
	"time"
)

func main() {
	addr := flag.String("addr", ":8080", "port to listen on")
	flag.Parse()
	log.Printf("port %s\n", *addr)

	handler, err := server.New(countrydb.NewDynamoDB())
	if err != nil {
		log.Fatalf("new server: %v\n", err)
	}

	s := &http.Server{
		Addr:              *addr,
		Handler:           handler,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
	}
	log.Fatal(s.ListenAndServe())
}