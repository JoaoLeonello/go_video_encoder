package services

import (
	"context"
	"encoder/application/repositories"
	"encoder/domain"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials" // Importação adicionada
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type VideoService struct {
	Video           *domain.Video
	VideoRepository repositories.VideoRepository
	s3Client        *s3.S3
}

func NewVideoService() VideoService {
	log.Println("Initializing AWS session")
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"",
		),
	}))
	s3Client := s3.New(sess)

	return VideoService{s3Client: s3Client}
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

	// Lê o conteúdo do objeto
	body, err := ioutil.ReadAll(obj.Body)
	if err != nil {
		return fmt.Errorf("error reading object body: %w", err)
	}

	// Salva o arquivo localmente
	filePath := os.Getenv("localStoragePath") + "/" + v.Video.ID + ".mp4"
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating local file: %w", err)
	}
	defer f.Close()

	_, err = f.Write(body)
	if err != nil {
		return fmt.Errorf("error writing to local file: %w", err)
	}

	log.Printf("video %v has been stored", v.Video.ID)

	return nil
}

func (v *VideoService) Fragment() error {

	err := os.Mkdir(os.Getenv("localStoragePath")+"/"+v.Video.ID, os.ModePerm)
	if err != nil {
		return err
	}

	source := os.Getenv("localStoragePath") + "/" + v.Video.ID + ".mp4"
	target := os.Getenv("localStoragePath") + "/" + v.Video.ID + ".frag"

	cmd := exec.Command("mp4fragment", source, target)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	printOutput(output)

	return nil
}

func (v *VideoService) Encode() error {
    // Caminho do arquivo de entrada (arquivo .frag gerado antes)
    inputPath := os.Getenv("localStoragePath") + "/" + v.Video.ID + ".frag"
    
    // Em vez de gerar outro .mp4, vamos criar um diretório de saída DASH
    dashOutputPath := os.Getenv("localStoragePath") + "/" + v.Video.ID + "_dash"

    // Definir o array de argumentos para o `mp4-dash.py`
    cmdArgs := []string{
        "/opt/bento4/Source/Python/utils/mp4-dash.py", // Caminho para o script
        inputPath,                                     // Arquivo .frag
        "--output",                                    // Flag para o caminho de saída
        dashOutputPath,                                // Diretório de saída
        "--use-segment-timeline",                      // Argumentos adicionais
        "-f",                                          // Força overwrite
    }

    // Executa o script Python
    cmd := exec.Command("python3", cmdArgs...)
    output, err := cmd.CombinedOutput()
    if err != nil {
        fmt.Printf("Command failed with error: %v\nOutput: %s\n", err, string(output))
        return err
    }

    fmt.Printf("Command succeeded. Output: %s\n", string(output))

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

	err = os.RemoveAll(os.Getenv("localStoragePath") + "/" + v.Video.ID)
	if err != nil {
		log.Println("error removing mp4 ", v.Video.ID, ".mp4")
		return err
	}

	log.Println("files have been removed: ", v.Video.ID)

	return nil

}

func (v *VideoService) InsertVideo() error  {
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
