package uploader

import "context"

// Handles file uploads and status notifications
// This abstraction allows the same pipeline to work in both local and distributed modes
// so we can easily switch for getting PUT url and db update from a function to an RPC
type Uploader interface {

	// Upload uploads a local file to storage
	Upload(ctx context.Context, localPath, objectKey string) error

	// NotifyPlayable signals that the video can start playing
	NotifyPlayable() error

	// NotifyComplete signals that all encoding is done
	NotifyComplete() error

	// NotifyFailed signals that encoding failed
	NotifyFailed(errMsg string) error
}
