package media

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/httpx"
	ctrlmw "sourcecraft.dev/lct-hackaton-2026.arena/internal/controller/middleware"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
	pkghttpx "sourcecraft.dev/lct-hackaton-2026.arena/pkg/httpx"
)

type Service interface {
	Upload(ctx context.Context, claims entity.AccessClaims, params entity.CreateMediaAsset) (*entity.MediaAsset, error)
	GetMedia(ctx context.Context, claims entity.AccessClaims, id string) (*entity.MediaAsset, error)
	GetContent(ctx context.Context, claims entity.AccessClaims, id string) (*entity.MediaContent, error)
}

type Controller struct {
	logger   *zap.Logger
	service  Service
	maxBytes int64
}

func NewController(logger *zap.Logger, service Service, maxBytes int64) *Controller {
	return &Controller{logger: logger, service: service, maxBytes: maxBytes}
}

func (c *Controller) InitRoutes(r chi.Router) {
	r.Post("/media", c.Upload)
	r.Get("/media/{media_id}", c.GetMedia)
	r.Get("/media/{media_id}/content", c.GetContent)
}

func (c *Controller) Upload(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFromRequest(w, r)
	if !ok {
		return
	}
	if err := r.ParseMultipartForm(c.maxBytes); err != nil {
		if errors.Is(err, multipart.ErrMessageTooLarge) {
			httpx.WriteError(w, http.StatusRequestEntityTooLarge, httpx.ErrMediaTooLarge)
			return
		}
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return
	}
	defer file.Close()
	payload, err := io.ReadAll(io.LimitReader(file, c.maxBytes+1))
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	duration, ok := optionalInt(w, r.FormValue("duration_seconds"))
	if !ok {
		return
	}
	asset, err := c.service.Upload(r.Context(), claims, entity.CreateMediaAsset{
		MediaType:       entity.MediaType(strings.TrimSpace(r.FormValue("media_type"))),
		ContentType:     contentType(header),
		Payload:         payload,
		DurationSeconds: duration,
	})
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusCreated, mediaToResponse(*asset))
}

func (c *Controller) GetMedia(w http.ResponseWriter, r *http.Request) {
	claims, id, ok := claimsAndMediaID(w, r)
	if !ok {
		return
	}
	asset, err := c.service.GetMedia(r.Context(), claims, id)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	pkghttpx.WriteJSON(w, http.StatusOK, mediaToResponse(*asset))
}

func (c *Controller) GetContent(w http.ResponseWriter, r *http.Request) {
	claims, id, ok := claimsAndMediaID(w, r)
	if !ok {
		return
	}
	content, err := c.service.GetContent(r.Context(), claims, id)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", content.ContentType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content.Payload)
}

func (c *Controller) writeError(w http.ResponseWriter, r *http.Request, err error) {
	code := httpx.ErrorCodeFromError(err)
	status := http.StatusInternalServerError
	switch code {
	case httpx.ErrInvalidRequest:
		status = http.StatusBadRequest
	case httpx.ErrMediaTooLarge:
		status = http.StatusRequestEntityTooLarge
	case httpx.ErrUnsupportedMediaType:
		status = http.StatusUnsupportedMediaType
	case httpx.ErrMediaAccessDenied:
		status = http.StatusForbidden
	case httpx.ErrMediaNotFound:
		status = http.StatusNotFound
	}
	if status == http.StatusInternalServerError {
		c.logger.Error("media request failed", zap.Error(err), zap.String("path", r.URL.Path), zap.String("code", string(code)))
	}
	httpx.WriteError(w, status, code)
}

func claimsFromRequest(w http.ResponseWriter, r *http.Request) (entity.AccessClaims, bool) {
	claims, ok := ctrlmw.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrAccessTokenInvalid)
		return entity.AccessClaims{}, false
	}
	return claims, true
}

func claimsAndMediaID(w http.ResponseWriter, r *http.Request) (entity.AccessClaims, string, bool) {
	claims, ok := claimsFromRequest(w, r)
	if !ok {
		return entity.AccessClaims{}, "", false
	}
	id, err := entity.NormalizeUUID(chi.URLParam(r, "media_id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return entity.AccessClaims{}, "", false
	}
	return claims, id, true
}

func optionalInt(w http.ResponseWriter, raw string) (*int, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, true
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrInvalidRequest)
		return nil, false
	}
	return &value, true
}

func contentType(header *multipart.FileHeader) string {
	if values := header.Header.Values("Content-Type"); len(values) > 0 {
		return values[0]
	}
	return "application/octet-stream"
}

func mediaToResponse(asset entity.MediaAsset) mediaAssetResponse {
	return mediaAssetResponse{
		MediaID:         asset.ID,
		MediaType:       string(asset.MediaType),
		ContentType:     asset.ContentType,
		URL:             "/api/v1/media/" + asset.ID + "/content",
		DurationSeconds: asset.DurationSeconds,
		Transcript:      nil,
		CreatedAt:       asset.CreatedAt,
	}
}
