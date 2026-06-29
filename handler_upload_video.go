package main

import (
	"fmt"
	"strings"
	"os"
	"io"
	"context"
	"net/http"
	"mime"
	"crypto/rand"
	"encoding/base64"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	const maxBytes = 1 << 30
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading video to S3", videoID, "by user", userID)

	meta, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to load video metadata", err)
		return
	}
	if meta.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Wrong user ID for given video", err)
		return
	}

	file, header, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	mediatype, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error when parsing header", err)
		return
	}
	if mediatype != "video/mp4" {
		respondWithError(w, http.StatusBadRequest, "Wrong media type", err)
		return
	}
	defer file.Close()

	fileExt := strings.Split(mediatype, "/")[1]
	vidPath := "tubely-upload" + "." + fileExt
	vidFile, err := os.CreateTemp("", vidPath)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error when creating file", err)
		return
	}
	defer os.Remove(vidFile.Name())
	defer vidFile.Close()

	_, err = io.Copy(vidFile, file)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error when copying file to disk", err)
		return
	}


	vidFile.Seek(0, io.SeekStart)

	procVidPath, err := processVideoForFastStart(vidFile.Name())
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error when preprocessing video", err)
		return
	}
	procVidFile, err := os.Open(procVidPath)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error when opening video file", err)
		return
	}
	defer os.Remove(procVidFile.Name())
	defer procVidFile.Close()

	aspect, err := getVideoAspectRatio(vidFile.Name())
	keyPrefix := "other"
	switch aspect {
	case "16:9":
		keyPrefix = "landscape"
	case "9:16":
		keyPrefix = "portrait"
	default:
	}

	keyMain := make([]byte, 32)
	rand.Read(keyMain)
	fString := base64.RawURLEncoding.EncodeToString(keyMain)
	key := fmt.Sprintf("%s/%s.%s", keyPrefix, fString, fileExt)

	paramsPut := s3.PutObjectInput{
		Bucket:			&cfg.s3Bucket,
		Key:			&key,
		Body:			procVidFile,
		ContentType:	&mediatype,
	}
	cfg.s3Client.PutObject(context.Background(), &paramsPut)

	newURL := fmt.Sprintf("%s,%s", cfg.s3Bucket, key)
	meta.VideoURL = &newURL

	err = cfg.db.UpdateVideo(meta)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to update video URL", err)
		return
	}

	signMeta, err := cfg.dbVideoToSignedVideo(meta)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error when signing URL", err)
	}

	respondWithJSON(w, http.StatusOK, signMeta)
}
