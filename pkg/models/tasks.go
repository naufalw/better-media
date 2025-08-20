package models

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const (
	TaskEncodeVideo = "task:encode_video"
)

type VideoEncodingPayload struct {
	VideoID      string   `json:"video_id"`
	InputFile    string   `json:"input_file"`
	TargetFormat string   `json:"target_format"`
	Resolutions  []string `json:"resolutions"`
}

func NewVideoEncodingTask(data VideoEncodingPayload) (*asynq.Task, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TaskEncodeVideo, payload), nil
}
