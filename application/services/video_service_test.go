package services_test

import (
	"encoder/application/repositories"
	"encoder/application/services"
	"encoder/domain"
	"encoder/framework/database"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load("../../.env")
}

// prepare pula o teste quando não há credenciais AWS: estes testes dependem de S3 real e do Bento4.
func prepare(t *testing.T) (*domain.Video, repositories.VideoRepositoryDb) {
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" {
		t.Skip("AWS_ACCESS_KEY_ID not set, skipping integration test")
	}

	db := database.NewDbTest()
	t.Cleanup(func() { db.Close() })

	video := domain.NewVideo()
	video.ID = uuid.NewV4().String()
	video.FilePath = "video.mp4"
	video.CreatedAt = time.Now()

	repo := repositories.VideoRepositoryDb{Db: db}

	return video, repo
}

func TestVideoServiceDownload(t *testing.T) {

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

	err = videoService.Finish()
	require.Nil(t, err)
}
