package models

// The payload to be sent to the external transcoder
type TranscodeRequest struct {
	JobID       string `json:"job_id"`
	VideoID     string `json:"video_id"`
	DownloadURL string `json:"download_url"` // Presigned S3 GET URL
	CallbackURL string `json:"callback_url"` // Where worker reports back. This is needed to generate presigned upload
	Resolutions []int  `json:"resolutions"`
	Transcribe  bool   `json:"transcribe"`
}
