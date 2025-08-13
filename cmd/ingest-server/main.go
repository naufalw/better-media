package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type UploadRequest struct {
    FileName string `json:"file_name" binding:"required"`
}

func main(){
	godotenv.Load()

	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {

		result := generatePresigned(c)

	    c.JSON(http.StatusUnauthorized, gin.H{
	      "message": result,
	    })
	 })
  	router.Run()
}

func generatePresigned(c *gin.Context) string {
	accessKeyId, accessKeySecret := os.Getenv("S3_ACCESS_KEY_ID"), os.Getenv("S3_ACCESS_KEY_SECRET")
	accountId := os.Getenv("S3_ACCOUNT_ID")
	bucketName := os.Getenv("S3_BUCKET_NAME")
	cfg, err := config.LoadDefaultConfig(c, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret, "")), config.WithRegion("auto"))

	if err != nil {
		log.Fatal(err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId))
	})

	presignClient := s3.NewPresignClient(client)
	presignResult, err := presignClient.PresignPostObject(c, &s3.PutObjectInput{
		Bucket: &bucketName,
		Key: aws.String("makan.mp4"),
	})

	if err != nil {
		log.Fatal(err , "Cant get url")
	}

	return presignResult.URL

}
