package entity

import "time"

type MediaType string

const (
	MediaTypeVoiceMessage MediaType = "voice_message"
)

type MediaAsset struct {
	ID              string
	OwnerID         string
	MediaType       MediaType
	ContentType     string
	DurationSeconds *int
	CreatedAt       time.Time
}

type CreateMediaAsset struct {
	OwnerID         string
	MediaType       MediaType
	ContentType     string
	Payload         []byte
	DurationSeconds *int
}

type MediaContent struct {
	OwnerID     string
	ContentType string
	Payload     []byte
}
