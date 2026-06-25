package main

import (
	"fmt"
	"net/http"
	"io"
	"os"
	"strings"
	"path/filepath"
	"mime"

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
	mediatype, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error when parsing header", err)
		return
	}
	if mediatype != "image/jpeg" || mediatype != "image/png" {
		respondWithError(w, http.StatusBadRequest, "Wrong media type", err)
		return
	}
	defer file.Close()

	meta, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to load image metadata", err)
		return
	}
	if meta.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Wrong user ID for given video", err)
		return
	}

	fileExt := strings.Split(mediatype, "/")[1]
	imgPath := filepath.Join(cfg.assetsRoot, videoIDString) + "." + fileExt
	imgFile, err := os.Create(imgPath)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error when creating file", err)
		return
	}
	_, err = io.Copy(imgFile, file)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error when copying file to disk", err)
		return
	}

	newURL := fmt.Sprintf("http://localhost:%s/assets/%s.%s", cfg.port, videoIDString, fileExt)
	meta.ThumbnailURL = &newURL
	err = cfg.db.UpdateVideo(meta)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to update thumbnail URL", err)
		return
	}

	respondWithJSON(w, http.StatusOK, meta)
}
