package media

import "time"

type mediaAssetResponse struct {
	MediaID         string    `json:"media_id"`
	MediaType       string    `json:"media_type"`
	ContentType     string    `json:"content_type"`
	URL             string    `json:"url"`
	DurationSeconds *int      `json:"duration_seconds"`
	Transcript      *string   `json:"transcript"`
	CreatedAt       time.Time `json:"created_at"`
}
