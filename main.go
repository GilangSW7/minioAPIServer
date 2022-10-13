package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"minioAPI/cmd/handler"
	"net/http"
	"time"
)

func main() {
	router := mux.NewRouter()

	srvOptions := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%s", "8080"),
		WriteTimeout: time.Second * 5,
		ReadTimeout:  time.Second * 5,
		IdleTimeout:  time.Second * 5,
		Handler:      router,
	}

	//handler
	router.HandleFunc("/bucket", handler.RequestPostHandler).Methods(http.MethodPost)
	router.HandleFunc("/bucket", handler.RequestGetHandler).Methods(http.MethodGet)
	router.HandleFunc("/bucket/{name}", handler.RequestGetObjectBucketListHandler).Methods(http.MethodGet)
	router.HandleFunc("/bucket/{name}", handler.RequestRemoveBucketHandler).Methods(http.MethodDelete)

	//start Server
	log.Fatalln(srvOptions.ListenAndServe())

}