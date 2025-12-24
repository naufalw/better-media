package dispatcher

import (
	"better-media/pkg/models"
	"context"
)

// We are going to support multiple mode
// Video Encoding Job -> Encode Request -> Either Process it locally as a monolith or send it to external transcoder
// There are two types
// - LocalDispatcher: runs in a goroutine
// - HTTPDIspatcher: POST to external worker.
type JobDispatcher interface {
	Dispatch(ctx context.Context, payload models.VideoEncodingPayload) (jobID string, err error)
}
