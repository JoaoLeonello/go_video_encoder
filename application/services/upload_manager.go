package services

import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

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
	// Inicializa a sessão da AWS e o cliente S3
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")), // Certifique-se de definir AWS_REGION no .env
	}))
	return &VideoUpload{s3Client: s3.New(sess)}
}

func (vu *VideoUpload) UploadObject(objectPath string) error {
	// Divide o caminho local para criar a chave do objeto no S3
	path := strings.Split(objectPath, os.Getenv("localStoragePath")+"/")

	f, err := os.Open(objectPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Lê o conteúdo do arquivo
	fileContent, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	// Configura o input para upload no S3
	input := &s3.PutObjectInput{
		Bucket: aws.String(vu.OutputBucket),
		Key:    aws.String(path[1]),
		Body:   strings.NewReader(string(fileContent)),
		ACL:    aws.String("public-read"), // Permissão equivalente ao ACL do GCS
	}

	// Faz o upload para o S3
	_, err = vu.s3Client.PutObject(input)
	if err != nil {
		return err
	}

	return nil
}

func (vu *VideoUpload) loadPaths() error {
	err := filepath.Walk(vu.VideoPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			vu.Paths = append(vu.Paths, path)
		}
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func (vu *VideoUpload) ProcessUpload(concurrency int, doneUpload chan string) error {
	in := make(chan int, runtime.NumCPU()) // Índice dos arquivos no slice Paths
	returnChannel := make(chan string)

	err := vu.loadPaths()
	if err != nil {
		return err
	}

	for process := 0; process < concurrency; process++ {
		go vu.uploadWorker(in, returnChannel)
	}

	go func() {
		for x := 0; x < len(vu.Paths); x++ {
			in <- x
		}
	}()

	countDoneWorker := 0
	for r := range returnChannel {
		countDoneWorker++

		if r != "" {
			doneUpload <- r
			break
		}

		if countDoneWorker == len(vu.Paths) {
			close(in)
		}
	}

	return nil
}

func (vu *VideoUpload) uploadWorker(in chan int, returnChan chan string) {
	for x := range in {
		err := vu.UploadObject(vu.Paths[x])

		if err != nil {
			vu.Errors = append(vu.Errors, vu.Paths[x])
			log.Printf("Error during the upload: %v. Error: %v", vu.Paths[x], err)
			returnChan <- err.Error()
		}

		returnChan <- ""
	}

	returnChan <- "upload completed"
}
