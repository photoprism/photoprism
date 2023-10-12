package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var userIconSvg = []byte(`
<svg xmlns="http://www.w3.org/2000/svg" width="24px" height="24px" viewBox="0 0 24 24" fill="#27282A">
  <path d="M20.474 19.013a8.941 8.941 0 0 0-4.115-4.89 6 6 0 1 0-8.717 0 8.941 8.941 0 0 0-4.115 4.89 11.065 11.065 0 0 0 1.63 1.59 6.965 6.965 0 0 1 4.728-5.275 1 1 0 0 0 .181-1.829 4 4 0 1 1 3.871 0 1 1 0 0 0 .181 1.829 6.965 6.965 0 0 1 4.726 5.272 11.059 11.059 0 0 0 1.63-1.587z"/>
</svg>`)

var faceIconSvg = []byte(`
<svg xmlns="http://www.w3.org/2000/svg" height="48" width="48" fill="#27282A"><path d="M31.3 21.35q1.15 0 1.925-.8.775-.8.775-1.9 0-1.15-.775-1.925-.775-.775-1.925-.775-1.1 0-1.9.775-.8.775-.8 1.925 0 1.1.8 1.9.8.8 1.9.8Zm-14.6 0q1.15 0 1.925-.8.775-.8.775-1.9 0-1.15-.775-1.925-.775-.775-1.925-.775-1.1 0-1.9.775-.8.775-.8 1.925 0 1.1.8 1.9.8.8 1.9.8Zm7.3 13.6q3.3 0 6.075-1.775Q32.85 31.4 34.1 28.35h-2.6q-1.15 2-3.15 3.075-2 1.075-4.3 1.075-2.35 0-4.375-1.05t-3.125-3.1H13.9q1.3 3.05 4.05 4.825Q20.7 34.95 24 34.95ZM24 44q-4.15 0-7.8-1.575-3.65-1.575-6.35-4.275-2.7-2.7-4.275-6.35Q4 28.15 4 24t1.575-7.8Q7.15 12.55 9.85 9.85q2.7-2.7 6.35-4.275Q19.85 4 24 4t7.8 1.575q3.65 1.575 6.35 4.275 2.7 2.7 4.275 6.35Q44 19.85 44 24t-1.575 7.8q-1.575 3.65-4.275 6.35-2.7 2.7-6.35 4.275Q28.15 44 24 44Zm0-20Zm0 17q7.1 0 12.05-4.95Q41 31.1 41 24q0-7.1-4.95-12.05Q31.1 7 24 7q-7.1 0-12.05 4.95Q7 16.9 7 24q0 7.1 4.95 12.05Q16.9 41 24 41Z"/></svg>`)

var cameraIconSvg = []byte(`
<svg xmlns="http://www.w3.org/2000/svg" height="48" width="48" fill="#27282A"><path d="M24 34.7q3.6 0 6.05-2.45 2.45-2.45 2.45-6.05 0-3.65-2.45-6.075Q27.6 17.7 24 17.7q-3.65 0-6.075 2.425Q15.5 22.55 15.5 26.2q0 3.6 2.425 6.05Q20.35 34.7 24 34.7ZM7 42q-1.2 0-2.1-.9Q4 40.2 4 39V13.35q0-1.15.9-2.075.9-.925 2.1-.925h7.35L18 6h12l3.65 4.35H41q1.15 0 2.075.925Q44 12.2 44 13.35V39q0 1.2-.925 2.1-.925.9-2.075.9Z"/></svg>`)

var photoIconSvg = []byte(`
<svg xmlns="http://www.w3.org/2000/svg" height="24" viewBox="0 0 24 24" width="24"><path d="M0 0h24v24H0V0z" fill="none"/>
<path d="M19 5v14H5V5h14m0-2H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm-4.86 8.86l-3 3.87L9 13.14 6 17h12l-3.86-5.14z"/></svg>`)

var rawIconSvg = []byte(`
<svg xmlns="http://www.w3.org/2000/svg" height="24" viewBox="0 0 24 24" width="24" fill="#27282A"><circle cx="12" cy="12" r="3.2"/>
<path d="M9 2L7.17 4H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V6c0-1.1-.9-2-2-2h-3.17L15 2H9zm3 15c-2.76 0-5-2.24-5-5s2.24-5 5-5 5 2.24 5 5-2.24 5-5 5z"/>
<path d="M0 0h24v24H0z" fill="none"/></svg>`)

var fileIconSvg = []byte(`<svg xmlns="http://www.w3.org/2000/svg" height="24" viewBox="0 0 24 24" width="24" fill="#27282A">
<path d="M6 2c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6H6zm7 7V3.5L18.5 9H13z"/><path d="M0 0h24v24H0z" fill="none"/></svg>`)

var videoIconSvg = []byte(`<svg xmlns="http://www.w3.org/2000/svg" height="24" viewBox="0 0 24 24" width="24" fill="#27282A">
<path d="M0 0h24v24H0z" fill="none"/><path d="M10 8v8l5-4-5-4zm9-5H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm0 16H5V5h14v14z"/></svg>`)

var folderIconSvg = []byte(`<svg xmlns="http://www.w3.org/2000/svg" height="24" viewBox="0 0 24 24" width="24" fill="#27282A"><path d="M0 0h24v24H0z" fill="none"/><path d="M10 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z"/></svg>`)

var albumIconSvg = folderIconSvg

var labelIconSvg = []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="#27282A">
<path d="M0 0h24v24H0z" fill="none"/><path d="M17.63 5.84C17.27 5.33 16.67 5 16 5L5 5.01C3.9 5.01 3 5.9 3 7v10c0 1.1.9 1.99 2 1.99L16 19c.67 0 1.27-.33 1.63-.84L22 12l-4.37-6.16z"/></svg>`)

var portraitIconSvg = []byte(`
<svg xmlns="http://www.w3.org/2000/svg" height="24px" viewBox="0 0 24 24" width="24px" fill="#27282A"><path d="M0 0h24v24H0V0z" fill="none"/><path d="M12 12c1.65 0 3-1.35 3-3s-1.35-3-3-3-3 1.35-3 3 1.35 3 3 3zm0-4c.55 0 1 .45 1 1s-.45 1-1 1-1-.45-1-1 .45-1 1-1zm6 8.58c0-2.5-3.97-3.58-6-3.58s-6 1.08-6 3.58V18h12v-1.42zM8.48 16c.74-.51 2.23-1 3.52-1s2.78.49 3.52 1H8.48zM19 3H5c-1.11 0-2 .9-2 2v14c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm0 16H5V5h14v14z"/></svg>`)

var brokenIconSvg = []byte(`
<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="#27282A">
<path fill="none" d="M0 0h24v24H0zm0 0h24v24H0zm21 19c0 1.1-.9 2-2 2H5c-1.1 0-2-.9-2-2V5c0-1.1.9-2 2-2h14c1.1 0 2 .9 2 2"/>
<path fill="none" d="M0 0h24v24H0z"/>
<path d="M21 5v6.59l-3-3.01-4 4.01-4-4-4 4-3-3.01V5c0-1.1.9-2 2-2h14c1.1 0 2 .9 2 2zm-3 6.42l3 3.01V19c0 1.1-.9 2-2 2H5c-1.1 0-2-.9-2-2v-6.58l3 2.99 4-4 4 4 4-3.99z"/></svg>`)

var uncachedIconSvg = []byte(`
<svg xmlns="http://www.w3.org/2000/svg" height="24" viewBox="0 0 24 24" width="24" fill="#27282A"><path d="M0 0h24v24H0z" fill="none"/>
<path d="M21 19V5c0-1.1-.9-2-2-2H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2zM8.5 13.5l2.5 3.01L14.5 12l4.5 6H5l3.5-4.5z"/></svg>`)

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
