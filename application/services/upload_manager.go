package services

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"encoding/json" // Para manipular JSON
	"github.com/aws/aws-sdk-go/aws/credentials" // Para gerenciar credenciais AWS
)

type VideoUpload struct {
	Paths        []string
	VideoPath    string
	OutputBucket string
	Errors       []string
	s3Client     *s3.S3
}

func NewVideoUpload() *VideoUpload {
	// Carregar o caminho do arquivo de credenciais do ambiente
	credFile := os.Getenv("AWS_APPLICATION_CREDENTIALS")
	if credFile == "" {
		log.Fatalf("AWS_APPLICATION_CREDENTIALS not set")
	}

	// Ler o arquivo JSON de credenciais
	data, err := os.ReadFile(credFile)
	if err != nil {
		log.Fatalf("Failed to read AWS credentials file: %v", err)
	}

	// Estrutura para armazenar credenciais
	type AWSCredentials struct {
		AccessKeyId     string `json:"AccessKeyId"`
		SecretAccessKey string `json:"SecretAccessKey"`
		Region          string `json:"Region"`
	}

	var creds AWSCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		log.Fatalf("Failed to unmarshal AWS credentials: %v", err)
	}

	// Criar uma sessão AWS com as credenciais carregadas
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(creds.Region),
		Credentials: credentials.NewStaticCredentials(
			creds.AccessKeyId,
			creds.SecretAccessKey,
			"",
		),
	})
	if err != nil {
		log.Fatalf("Failed to create AWS session: %v", err)
	}

	// Inicializa o cliente S3
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
