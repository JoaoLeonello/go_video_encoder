package services

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type VideoUpload struct {
	Paths        []string
	VideoPath    string
	OutputBucket string
	Errors       []string
	s3Client     *s3.S3
}

func NewVideoUpload() *VideoUpload {
	return &VideoUpload{s3Client: newS3Client()}
}

// newS3Client usa a cadeia padrão de credenciais da AWS (variáveis AWS_*, ~/.aws, IAM role).
func newS3Client() *s3.S3 {
	return s3.New(session.Must(session.NewSession()))
}

func (vu *VideoUpload) UploadObject(objectPath string) error {
	// A chave no S3 é o caminho relativo a localStoragePath, ex.: "<video_id>_dash/manifest.mpd"
	key, err := filepath.Rel(os.Getenv("localStoragePath"), objectPath)
	if err != nil {
		return err
	}

	f, err := os.Open(objectPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = vu.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(vu.OutputBucket),
		Key:                  aws.String(key),
		Body:                 f,
		ServerSideEncryption: aws.String("AES256"),
	})
	return err
}

func (vu *VideoUpload) loadPaths() error {
	return filepath.Walk(vu.VideoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			vu.Paths = append(vu.Paths, path)
		}
		return nil
	})
}

// ProcessUpload envia todos os arquivos de VideoPath para o bucket usando até
// `concurrency` goroutines. Retorna erro se algum upload falhar.
func (vu *VideoUpload) ProcessUpload(concurrency int) error {
	if err := vu.loadPaths(); err != nil {
		return err
	}
	if len(vu.Paths) == 0 {
		return fmt.Errorf("no files to upload in %s", vu.VideoPath)
	}
	if concurrency < 1 {
		concurrency = 1
	}

	in := make(chan string)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range in {
				if err := vu.UploadObject(path); err != nil {
					log.Printf("Error during the upload: %v. Error: %v", path, err)
					mu.Lock()
					vu.Errors = append(vu.Errors, path)
					mu.Unlock()
				}
			}
		}()
	}

	for _, path := range vu.Paths {
		in <- path
	}
	close(in)
	wg.Wait()

	if len(vu.Errors) > 0 {
		return fmt.Errorf("failed to upload %d of %d files", len(vu.Errors), len(vu.Paths))
	}
	return nil
}
