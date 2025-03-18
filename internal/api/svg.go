package api

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed embed/user.svg
var userIconSvg []byte

//go:embed embed/face.svg
var faceIconSvg []byte

//go:embed embed/camera.svg
var cameraIconSvg []byte

//go:embed embed/photo.svg
var photoIconSvg []byte

//go:embed embed/raw.svg
var rawIconSvg []byte

//go:embed embed/file.svg
var fileIconSvg []byte

//go:embed embed/video.svg
var videoIconSvg []byte

//go:embed embed/folder.svg
var folderIconSvg []byte

//go:embed embed/album.svg
var albumIconSvg []byte

//go:embed embed/label.svg
var labelIconSvg []byte

//go:embed embed/portrait.svg
var portraitIconSvg []byte

//go:embed embed/broken.svg
var brokenIconSvg []byte

//go:embed embed/uncached.svg
var uncachedIconSvg []byte

// GetSvg returns SVG placeholder symbols.
//
// GET /api/v1/svg/*
func GetSvg(router *gin.RouterGroup) {
	router.GET("/svg/user", func(c *gin.Context) {
		c.Data(http.StatusOK, "image/svg+xml", userIconSvg)
	})

	router.GET("/svg/face", func(c *gin.Context) {
		c.Data(http.StatusOK, "image/svg+xml", faceIconSvg)
	})

	router.GET("/svg/camera", func(c *gin.Context) {
		c.Data(http.StatusOK, "image/svg+xml", cameraIconSvg)
	})

	router.GET("/svg/photo", func(c *gin.Context) {
		c.Data(http.StatusOK, "image/svg+xml", photoIconSvg)
	})

	router.GET("/svg/raw", func(c *gin.Context) {
		c.Data(http.StatusOK, "image/svg+xml", rawIconSvg)
	})

	router.GET("/svg/file", func(c *gin.Context) {
		c.Data(http.StatusOK, "image/svg+xml", fileIconSvg)
	})

	router.GET("/svg/video", func(c *gin.Context) {
		c.Data(http.StatusOK, "image/svg+xml", videoIconSvg)
	})

	router.GET("/svg/label", func(c *gin.Context) {
		c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)
	})

	router.GET("/svg/portrait", func(c *gin.Context) {
		c.Data(http.StatusOK, "image/svg+xml", portraitIconSvg)
	})

	router.GET("/svg/folder", func(c *gin.Context) {
		c.Data(http.StatusOK, "image/svg+xml", folderIconSvg)
	})

	router.GET("/svg/album", func(c *gin.Context) {
		c.Data(http.StatusOK, "image/svg+xml", albumIconSvg)
	})

	router.GET("/svg/broken", func(c *gin.Context) {
		c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)
	})

	router.GET("/svg/uncached", func(c *gin.Context) {
		c.Data(http.StatusOK, "image/svg+xml", uncachedIconSvg)
	})
}
