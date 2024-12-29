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
    video, repo := prepare()

    videoService := services.NewVideoService()
    videoService.Video = video
    videoService.VideoRepository = repo

    err := videoService.Download("video-encoder-test")
    require.Nil(t, err)

    err = videoService.Fragment()
    require.Nil(t, err)

    err = videoService.Encode()
    require.Nil(t, err)

    log.Println("Preparing VideoUpload****************************")
    videoUpload := services.NewVideoUpload()
    videoUpload.OutputBucket = "video-encoder-test-output"
    videoUpload.VideoPath = os.Getenv("localStoragePath") + "/" + video.ID + "_dash"

    log.Println("Starting ProcessUpload------")
    doneUpload := make(chan string)
    go videoUpload.ProcessUpload(50, doneUpload)

    result := <-doneUpload
    require.Equal(t, result, "upload completed")

    log.Println("Finishing")
    err = videoService.Finish()
    require.Nil(t, err)
}


