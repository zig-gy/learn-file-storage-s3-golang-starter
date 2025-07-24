package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
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
		respondWithError(w, http.StatusUnauthorized, "Coudln't validate JWT", err)
		return
	}

	fmt.Println("uploading video file for video", videoID, "by user", userID)

	const maxMemory = 1 << 30
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	videoMetadata, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Video not found", err)
		return
	}

	if userID != videoMetadata.UserID {
		respondWithError(w, http.StatusUnauthorized, "User does not match with video author", fmt.Errorf("logged in user does not match video author"))
		return
	}

	mediaType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not parse media type", err)
		return
	}

	if mediaType != "video/mp4" {
		respondWithError(w, http.StatusUnsupportedMediaType, "Wrong media type", fmt.Errorf("only upload mp4"))
		return
	}

	tmpFile, err := os.CreateTemp("", "tubely-tmp.mp4")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create file", err)
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, file); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error copying file", err)
		return
	}

	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error changing pointer", err)
		return
	}

	var prefix string
	aspectRatio, err := getVideoAspectRatio(tmpFile.Name())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get aspect ratio", err)
		return
	}

	switch aspectRatio {
	case "16:9":
		prefix = "landscape"
	case "9:16":
		prefix = "portrait"
	default:
		prefix = "other"
	}

	processedFilePath, err := processVideoForFastStart(tmpFile.Name())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not process video", err)
		return
	}

	processedFile, err := os.Open(processedFilePath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not open processed file", err)
		return
	}
	defer os.Remove(processedFilePath)
	defer processedFile.Close()

	fileExt := strings.Split(mediaType, "/")[1]
	filenameBytes := make([]byte, 32)
	rand.Read(filenameBytes)
	filename := base64.RawURLEncoding.EncodeToString(filenameBytes)
	filenameWithExt := fmt.Sprintf("%s/%s.%s", prefix, filename, fileExt)

	_, err = cfg.s3Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      &cfg.s3Bucket,
		Key:         &filenameWithExt,
		Body:        processedFile,
		ContentType: &mediaType,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not store video in bucket", err)
		return
	}

	bucketUrl := fmt.Sprintf("%s,%s", cfg.s3Bucket, filenameWithExt)
	videoMetadata.VideoURL = &bucketUrl
	videoMetadata.UpdatedAt = time.Now()

	if err := cfg.db.UpdateVideo(videoMetadata); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not update video", err)
		return
	}

	signedVideo, err := cfg.dbVideoToSignedVideo(videoMetadata)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create signed video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, signedVideo)
}
