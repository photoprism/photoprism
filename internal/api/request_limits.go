package api

import (
	"errors"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/pkg/i18n"
)

const (
	// MaxAuthRequestBytes bounds small authentication and credential-change payloads.
	MaxAuthRequestBytes int64 = 64 * 1024
	// MaxMutationRequestBytes bounds general JSON mutation payloads.
	MaxMutationRequestBytes int64 = 256 * 1024
	// MaxSelectionRequestBytes bounds selection-heavy batch mutation payloads.
	MaxSelectionRequestBytes int64 = 1024 * 1024
	// MaxVisionRequestBytes bounds Vision API payloads while still allowing supported data URLs.
	MaxVisionRequestBytes int64 = 32 * 1024 * 1024
	// MaxSessionRequestBytes bounds login request payloads.
	MaxSessionRequestBytes int64 = MaxAuthRequestBytes
	// MaxAlbumRequestBytes bounds album create and update payloads.
	MaxAlbumRequestBytes int64 = MaxMutationRequestBytes
	// MaxSettingsRequestBytes bounds settings and config option payloads.
	MaxSettingsRequestBytes int64 = MaxMutationRequestBytes
	// MaxClusterRegisterBytes bounds cluster registration and patch payloads.
	MaxClusterRegisterBytes int64 = MaxMutationRequestBytes
	// MaxUploadOptionsRequestBytes bounds upload-processing payloads.
	MaxUploadOptionsRequestBytes int64 = MaxMutationRequestBytes
	// MaxBatchPhotosEditBytes bounds batch photo edit payloads.
	MaxBatchPhotosEditBytes int64 = MaxSelectionRequestBytes
	// MaxMultipartOverheadBytes reserves room for multipart framing overhead.
	MaxMultipartOverheadBytes int64 = 1024 * 1024
	// MaxAvatarUploadBytes bounds avatar uploads including multipart overhead.
	MaxAvatarUploadBytes int64 = 20000000 + MaxMultipartOverheadBytes
)

// LimitRequestBodyBytes caps the readable request body size for the current handler.
func LimitRequestBodyBytes(c *gin.Context, limit int64) {
	if c == nil || c.Request == nil || c.Request.Body == nil || limit <= 0 {
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
}

// IsRequestBodyTooLarge reports whether the parsing error was caused by a body-size limit.
func IsRequestBodyTooLarge(err error) bool {
	if err == nil {
		return false
	}

	var maxBytesErr *http.MaxBytesError

	return errors.As(err, &maxBytesErr) || errors.Is(err, multipart.ErrMessageTooLarge)
}

// AbortRequestTooLarge writes a localized 413 response and stops the current request.
func AbortRequestTooLarge(c *gin.Context, id i18n.Message) {
	Abort(c, http.StatusRequestEntityTooLarge, id)
}
