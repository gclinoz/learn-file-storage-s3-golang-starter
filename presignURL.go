package main

import (
	"time"
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
	psClient := s3.NewPresignClient(s3Client)

	presignParam := s3.GetObjectInput{
		Bucket:	&bucket,
		Key:	&key,
	}

	psResq, err := psClient.PresignGetObject(context.Background(), &presignParam, s3.WithPresignExpires(expireTime))
	if err != nil {
		return "", err
	}

	return psResq.URL, nil
}
