package services

import (
	"context"
	"fmt"
	"log"
	"sync"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"sbminimalistics.com/cctv-visionfusion/downloader/internal/config"
)

var (
	serviceInstance *minio.Client
	createOnce      sync.Once
)

type DownloadLinkService struct {
	ctx context.Context
	mu  sync.Mutex
}

func GetS3Service() *minio.Client {
	if serviceInstance == nil {
		panic("S3Service requested before initialization")
	}
	return serviceInstance
}

func InitS3Service(ctx context.Context) *minio.Client {
	endpoint := fmt.Sprintf("%s:%d", config.Config.Seaweed.Host, config.Config.Seaweed.Port)
	useSSL := false

	fmt.Println("InitService s3:", config.Config.Seaweed.User, config.Config.Seaweed.Password)

	client, err := minio.New(endpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(config.Config.Seaweed.User, config.Config.Seaweed.Password, ""),
		Secure:       useSSL,
		BucketLookup: minio.BucketLookupPath,
		Region:       "us-east-1",
	})
	// client.TraceOn(os.Stdout)
	if err != nil {
		log.Fatalf("Failed to initialize storage client: %v", err)
	}
	serviceInstance = client

	bucketName := config.Config.Seaweed.Bucket

	// create bucker
	// exists, err := s3Client.BucketExists(ctx, bucketName)
	// if err != nil {
	// 	log.Fatalf("Error checking bucket: %v", err)
	// }

	// if !exists {
	// 	fmt.Printf("Bucket %s does not exist, creating it now...\n", bucketName)
	// 	err = s3Client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	// 	if err != nil {
	// 		log.Fatalf("Failed to create bucket: %v", err)
	// 	}
	// }

	fmt.Printf("successfully connected to SeaweedFS, bucket '%s' is ready.\n", bucketName)

	return serviceInstance
}
