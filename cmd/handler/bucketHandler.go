package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
	"minioAPI/cmd/model"
	"minioAPI/configs"
	"net/http"
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
