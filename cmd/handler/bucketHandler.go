package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
	"log"
	"minioAPI/cmd/model"
	"minioAPI/configs"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)

// connect to minio
var s3Client = configs.Connect()

func RequestPostHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Receive request for make bucket list")

	var bucketReq model.Bucket
	jsonDecoder := json.NewDecoder(r.Body)
	err := jsonDecoder.Decode(&bucketReq)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error Decode json Body"))
		return
	}

	err = s3Client.MakeBucket(context.Background(), bucketReq.Name, minio.MakeBucketOptions{Region: bucketReq.Region, ObjectLocking: true})
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error When Create Bucket"))
		return
	}
	jsonResponseByte, _ := json.Marshal(bucketReq)
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponseByte)
}

func RequestGetHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Receive request for get bucket list")

	buckets, err := s3Client.ListBuckets(context.Background())
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server Error"))
		return
	} else if len(buckets) < 1 {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("no Bucket Created yet"))

	} else {
		jsonResponseByte, _ := json.Marshal(buckets)
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponseByte)
	}
}

func RequestGetObjectBucketListHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Receive request for get bucket")

	vars := mux.Vars(r)
	name := vars["name"]

	found, err := s3Client.BucketExists(context.Background(), name)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server Error"))
		return
	}
	if !found {
		fmt.Println("Bucket not found")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Bucket not Found"))
	}
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	objectCh := s3Client.ListObjects(ctx, name, minio.ListObjectsOptions{
		Recursive: true,
	})

	var objetList = make([]minio.ObjectInfo, 0)

	for object := range objectCh {
		if object.Err != nil {
			fmt.Println(object.Err)
			return
		}
		objetList = append(objetList, object)
	}

	if len(objetList) > 0 {
		jsonResponseByte, _ := json.Marshal(objetList)
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponseByte)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("no Object Uploaded yet"))
	}
}

func RequestRemoveBucketHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Receive request for delete bucket")

	vars := mux.Vars(r)
	name := vars["name"]

	err := s3Client.RemoveBucket(context.Background(), name)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Bucket Delete is Success"))
}

func RequestUploadObjectToBuckettHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Receive request for get bucket")

	chBucket := make(chan string)
	vars := mux.Vars(r)
	name := vars["name"]

	go cekBucketExist(name, chBucket)

	r.ParseMultipartForm(10 << 20)
	file, handler, err := r.FormFile("myFile")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error Retrieving the File"))
		fmt.Println(err)
		return
	}
	defer file.Close()
	ext1 := filepath.Ext(handler.Filename)
	if ext1 != ".jpg" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Extension not supported"))
		return
	}

	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	uploadInfo, err := s3Client.PutObject(context.Background(), name, handler.Filename, file, handler.Size, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		fmt.Println(err)
		return
	}
	jsonResponseByte, _ := json.Marshal(uploadInfo)
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponseByte)
}

func cekBucketExist(name string, chBucket chan<- string) {
	found, err := s3Client.BucketExists(context.Background(), name)

	if err != nil {
		fmt.Println(err)
		chBucket <- "internal server Error"
		return
	}
	if !found {
		fmt.Println("Bucket not found")
		chBucket <- "Bucket not found"
		return
	}
	chBucket <- "found"
}

func RequestSchedulerDeleteObjectBucket(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Receive request for Scheduler Delete Object")

	objectsCh := make(chan minio.ObjectInfo)
	vars := mux.Vars(r)
	name := vars["name"]

	count := 0

	go func() {
		defer close(objectsCh)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		for object := range s3Client.ListObjects(ctx, name, minio.ListObjectsOptions{
			Recursive: true,
		}) {
			if object.Err != nil {
				log.Fatalln(object.Err)
			}
			if object.LastModified.Before(time.Now().AddDate(0, 0, -1)) {
				objectsCh <- object
				count++
			}

		}
	}()

	opts := minio.RemoveObjectsOptions{
		GovernanceBypass: true,
	}

	for rErr := range s3Client.RemoveObjects(context.Background(), name, objectsCh, opts) {
		fmt.Println("Error detected during deletion: ", rErr)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Scheduler is Running and Succes delete " + strconv.Itoa(count)))
}
