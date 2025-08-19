package models

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const (
	TaskEncodeVideo = "task:encode_video"
)

type VideoEncodingPayload struct {
	VideoID string `json:"video_id"`
}

func NewVideoEncodingTask(id string) (*asynq.Task, error) {
	payload, err := json.Marshal(VideoEncodingPayload{VideoID: id})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TaskEncodeVideo, payload), nil
}
