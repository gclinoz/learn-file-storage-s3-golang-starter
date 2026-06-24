package main

import (
	"fmt"
	"net/http"
	"io"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
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


	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	ct := header.Header.Get("Content-Type")
	defer file.Close()

	img, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to read image data", err)
		return
	}

	meta, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to load image metadata", err)
		return
	}
	if meta.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Wrong user ID for given video", err)
		return
	}

	tmb := thumbnail{
		data:		img,
		mediaType:	ct,
	}
	videoThumbnails[videoID] = tmb

	newURL := fmt.Sprintf("http://localhost:%s/api/thumbnails/%s", cfg.port, videoIDString)
	meta.ThumbnailURL = &newURL
	err = cfg.db.UpdateVideo(meta)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to update thumbnail URL", err)
		return
	}

	respondWithJSON(w, http.StatusOK, meta)
}
