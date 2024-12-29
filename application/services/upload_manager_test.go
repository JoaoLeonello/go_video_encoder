package services_test

import (
	"encoder/application/services"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"log"
	"os"
	"testing"
)

func init() {
	// Carregar variáveis do .env no início do pacote de testes
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func TestVideoServiceUpload(t *testing.T) {
    log.Println("Starting TestVideoServiceUpload")
    video, repo := prepare()

    log.Println("Initializing VideoService")
    videoService := services.NewVideoService()
    videoService.Video = video
    videoService.VideoRepository = repo

    log.Println("Starting Download")
    err := videoService.Download("video-encoder-test-output")
    require.Nil(t, err)

    log.Println("Starting Fragment")
    err = videoService.Fragment()
    require.Nil(t, err)

    log.Println("Starting Encode")
    err = videoService.Encode()
    require.Nil(t, err)

    log.Println("Preparing VideoUpload")
    videoUpload := services.NewVideoUpload()
    videoUpload.OutputBucket = "video-encoder-test-output"
    videoUpload.VideoPath = os.Getenv("localStoragePath") + "/" + video.ID

    log.Println("Starting ProcessUpload")
    doneUpload := make(chan string)
    go videoUpload.ProcessUpload(50, doneUpload)

    result := <-doneUpload
    require.Equal(t, result, "upload completed")

    log.Println("Finishing")
    err = videoService.Finish()
    require.Nil(t, err)
}


