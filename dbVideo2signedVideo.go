package main

import (
	"time"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	if video.VideoURL == nil {
		return video, nil
	}

	bucket := strings.Split(*video.VideoURL, ",")[0]
	key := strings.Split(*video.VideoURL, ",")[1]

	presignURL, err := generatePresignedURL(cfg.s3Client, bucket, key, time.Minute * 30)
	if err != nil {
		return database.Video{}, err
	}
	video.VideoURL = &presignURL

	return video, nil
}
