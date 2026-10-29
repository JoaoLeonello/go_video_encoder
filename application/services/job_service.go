package services

import (
	"encoder/application/repositories"
	"encoder/domain"
	"os"
	"strconv"
)

type JobService struct {
	Job           *domain.Job
	JobRepository repositories.JobRepository
	VideoService  VideoService
}

func (j *JobService) Start() error {

	err := j.changeJobStatus("DOWNLOADING")

	if err != nil {
		return j.failJob(err)
	}

	err = j.VideoService.Download(os.Getenv("inputBucketName"))

	if err != nil {
		return j.failJob(err)
	}

	err = j.changeJobStatus("FRAGMENTING")

	if err != nil {
		return j.failJob(err)
	}

	err = j.VideoService.Fragment()

	if err != nil {
		return j.failJob(err)
	}

	err = j.changeJobStatus("ENCODING")

	if err != nil {
		return j.failJob(err)
	}

	err = j.VideoService.Encode()

	if err != nil {
		return j.failJob(err)
	}

	err = j.performUpload()

	if err != nil {
		return j.failJob(err)
	}

	err = j.changeJobStatus("FINISHING")

	if err != nil {
		return j.failJob(err)
	}

	err = j.VideoService.Finish()

	if err != nil {
		return j.failJob(err)
	}

	err = j.changeJobStatus("COMPLETED")

	if err != nil {
		return j.failJob(err)
	}

	return nil
}

func (j *JobService) performUpload() error {

	err := j.changeJobStatus("UPLOADING")

	if err != nil {
		return err
	}

	videoUpload := NewVideoUpload()
	videoUpload.OutputBucket = os.Getenv("outputBucketName")
	videoUpload.VideoPath = j.VideoService.DashPath()
	concurrency, _ := strconv.Atoi(os.Getenv("CONCURRENCY_UPLOAD"))

	return videoUpload.ProcessUpload(concurrency)
}

func (j *JobService) changeJobStatus(status string) error {
	j.Job.Status = status
	_, err := j.JobRepository.Update(j.Job)

	return err
}

func (j *JobService) failJob(error error) error {

	j.Job.Status = "FAILED"
	j.Job.Error = error.Error()

	_, err := j.JobRepository.Update(j.Job)

	if err != nil {
		return err
	}

	return error
}
