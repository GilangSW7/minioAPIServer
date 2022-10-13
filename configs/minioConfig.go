package configs

import (
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"os"
)

func Connect() *minio.Client {
	// Initialize minio client object.
	s3Client, err := minio.New(os.Getenv("END_POINT"), &minio.Options{
		Creds: credentials.NewStaticV4(os.Getenv("API_KEY"), os.Getenv("SECRET_KEY"), ""),
		//Secure: true,
	})
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	return s3Client

}
