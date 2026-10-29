package services

import (
	"context"
	"encoder/application/repositories"
	"encoder/domain"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

type VideoService struct {
	Video           *domain.Video
	VideoRepository repositories.VideoRepository
	s3Client        *s3.S3
}

func NewVideoService() VideoService {
	return VideoService{s3Client: newS3Client()}
}

// DashPath é o diretório onde o Encode grava a saída MPEG-DASH.
func (v *VideoService) DashPath() string {
	return os.Getenv("localStoragePath") + "/" + v.Video.ID + "_dash"
}

func (v *VideoService) Download(bucketName string) error {
	// Configura o contexto
	ctx := context.Background()

	// Faz o download do arquivo do S3
	objInput := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(v.Video.FilePath),
	}

	obj, err := v.s3Client.GetObjectWithContext(ctx, objInput)
	if err != nil {
		return fmt.Errorf("error getting object from S3: %w", err)
	}
	defer obj.Body.Close()

	// Salva o arquivo localmente, em streaming, sem carregar o vídeo inteiro em memória
	filePath := os.Getenv("localStoragePath") + "/" + v.Video.ID + ".mp4"
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating local file: %w", err)
	}
	defer f.Close()

	_, err = io.Copy(f, obj.Body)
	if err != nil {
		return fmt.Errorf("error writing to local file: %w", err)
	}

	log.Printf("video %v has been stored", v.Video.ID)

	return nil
}

func (v *VideoService) Fragment() error {

	source := os.Getenv("localStoragePath") + "/" + v.Video.ID + ".mp4"
	target := os.Getenv("localStoragePath") + "/" + v.Video.ID + ".frag"

	cmd := exec.Command("mp4fragment", source, target)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mp4fragment failed: %w: %s", err, output)
	}

	printOutput(output)

	return nil
}

func (v *VideoService) Encode() error {
	// Caminho do arquivo de entrada (arquivo .frag gerado antes)
	inputPath := os.Getenv("localStoragePath") + "/" + v.Video.ID + ".frag"

	// Em vez de gerar outro .mp4, vamos criar um diretório de saída DASH
	dashOutputPath := v.DashPath()

	// Definir o array de argumentos para o `mp4-dash.py`
	cmdArgs := []string{
		"/opt/bento4/Source/Python/utils/mp4-dash.py", // Caminho para o script
		inputPath,                // Arquivo .frag
		"--output",               // Flag para o caminho de saída
		dashOutputPath,           // Diretório de saída
		"--use-segment-timeline", // Argumentos adicionais
		"-f",                     // Força overwrite
	}

	// Executa o script Python
	cmd := exec.Command("python3", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mp4-dash failed: %w: %s", err, output)
	}

	printOutput(output)

	return nil
}

func (v *VideoService) Finish() error {

	err := os.Remove(os.Getenv("localStoragePath") + "/" + v.Video.ID + ".mp4")
	if err != nil {
		log.Println("error removing mp4 ", v.Video.ID, ".mp4")
		return err
	}

	err = os.Remove(os.Getenv("localStoragePath") + "/" + v.Video.ID + ".frag")
	if err != nil {
		log.Println("error removing frag ", v.Video.ID, ".frag")
		return err
	}

	err = os.RemoveAll(v.DashPath())
	if err != nil {
		log.Println("error removing dash dir ", v.DashPath())
		return err
	}

	log.Println("files have been removed: ", v.Video.ID)

	return nil

}

func (v *VideoService) InsertVideo() error {
	_, err := v.VideoRepository.Insert(v.Video)

	if err != nil {
		return err
	}

	return nil
}

func printOutput(out []byte) {
	if len(out) > 0 {
		log.Printf("=====> Output: %s\n", string(out))
	}
}
