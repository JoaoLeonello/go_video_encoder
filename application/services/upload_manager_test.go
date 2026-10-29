package services_test

import (
	"encoder/application/services"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
)

func TestVideoServiceUpload(t *testing.T) {
	video, repo := prepare(t)

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
	videoUpload.VideoPath = videoService.DashPath()

	log.Println("Starting ProcessUpload------")
	err = videoUpload.ProcessUpload(50)
	require.Nil(t, err)

	log.Println("Finishing")
	err = videoService.Finish()
	require.Nil(t, err)
}

func TestProcessUploadFailsWithoutFiles(t *testing.T) {
	videoUpload := services.NewVideoUpload()

	videoUpload.VideoPath = t.TempDir()
	require.Error(t, videoUpload.ProcessUpload(4))

	videoUpload.VideoPath = t.TempDir() + "/missing"
	require.Error(t, videoUpload.ProcessUpload(4))
}
