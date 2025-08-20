package models

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const (
	TaskEncodeVideo = "task:encode_video"
)

type VideoEncodingPayload struct {
	VideoID      string `json:"video_id" binding:"required"`
	InputFile    string `json:"input_file" binding:"required"`
	TargetFormat string `json:"target_format" binding:"required"`
	Resolutions  []int  `json:"resolutions" binding:"required"`
}

func NewVideoEncodingTask(data VideoEncodingPayload) (*asynq.Task, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TaskEncodeVideo, payload), nil
}
