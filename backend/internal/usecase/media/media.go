package media

import (
	"context"
	"strings"

	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
)

type repository interface {
	CreateMedia(ctx context.Context, params entity.CreateMediaAsset) (*entity.MediaAsset, error)
	GetMedia(ctx context.Context, id string) (*entity.MediaAsset, error)
	GetMediaContent(ctx context.Context, id string) (*entity.MediaContent, error)
}

type Service struct {
	repository repository
	maxBytes   int64
}

func NewService(repository repository, maxBytes int64) *Service {
	return &Service{repository: repository, maxBytes: maxBytes}
}

func (s *Service) Upload(ctx context.Context, claims entity.AccessClaims, params entity.CreateMediaAsset) (*entity.MediaAsset, error) {
	if int64(len(params.Payload)) > s.maxBytes {
		return nil, entity.ErrMediaTooLarge
	}
	if params.MediaType != entity.MediaTypeVoiceMessage {
		return nil, entity.ErrUnsupportedMediaType
	}
	if !supportedContentType(params.ContentType) {
		return nil, entity.ErrUnsupportedMediaType
	}
	params.OwnerID = claims.UserID
	return s.repository.CreateMedia(ctx, params)
}

func (s *Service) GetMedia(ctx context.Context, claims entity.AccessClaims, id string) (*entity.MediaAsset, error) {
	asset, err := s.repository.GetMedia(ctx, id)
	if err != nil {
		return nil, err
	}
	if asset.OwnerID != claims.UserID {
		return nil, entity.ErrMediaAccessDenied
	}
	return asset, nil
}

func (s *Service) GetContent(ctx context.Context, claims entity.AccessClaims, id string) (*entity.MediaContent, error) {
	content, err := s.repository.GetMediaContent(ctx, id)
	if err != nil {
		return nil, err
	}
	if content.OwnerID != claims.UserID {
		return nil, entity.ErrMediaAccessDenied
	}
	return content, nil
}

func supportedContentType(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "audio/mpeg" || value == "audio/mp4" || value == "audio/wav" || value == "audio/webm" || value == "audio/ogg" || value == "application/octet-stream"
}
