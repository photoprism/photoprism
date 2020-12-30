package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var photoIconSvg = []byte(`
<svg xmlns="http://www.w3.org/2000/svg" height="24" viewBox="0 0 24 24" width="24"><path d="M0 0h24v24H0V0z" fill="none"/>
<path d="M19 5v14H5V5h14m0-2H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm-4.86 8.86l-3 3.87L9 13.14 6 17h12l-3.86-5.14z"/></svg>`)

var rawIconSvg = []byte(`
<svg xmlns="http://www.w3.org/2000/svg" height="24" viewBox="0 0 24 24" width="24"><circle cx="12" cy="12" r="3.2"/>
<path d="M9 2L7.17 4H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V6c0-1.1-.9-2-2-2h-3.17L15 2H9zm3 15c-2.76 0-5-2.24-5-5s2.24-5 5-5 5 2.24 5 5-2.24 5-5 5z"/>
<path d="M0 0h24v24H0z" fill="none"/></svg>`)

var fileIconSvg = []byte(`<svg xmlns="http://www.w3.org/2000/svg" height="24" viewBox="0 0 24 24" width="24">
<path d="M6 2c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6H6zm7 7V3.5L18.5 9H13z"/><path d="M0 0h24v24H0z" fill="none"/></svg>`)

var videoIconSvg = []byte(`<svg xmlns="http://www.w3.org/2000/svg" height="24" viewBox="0 0 24 24" width="24">
<path d="M0 0h24v24H0z" fill="none"/><path d="M10 8v8l5-4-5-4zm9-5H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm0 16H5V5h14v14z"/></svg>`)

var albumIconSvg = []byte(`<svg id="f394fb39-4024-46e8-9a68-3f5e53f52bfd" data-name="Ebene 1" xmlns="http://www.w3.org/2000/svg"
     viewBox="0 0 276.76 276.76">
    <defs>
        <style>
            .a8bd947d-a1f8-4a14-8069-c9663505e4b1{fill:url(#b905e2c4-4772-4914-b70f-44b4d19aca4b);}.a4ee1f07-8fee-4200-835a-e030a32bca90{fill:none;stroke:#1d1d1b;stroke-miterlimit:10;stroke-width:5px;}
        </style>
        <linearGradient id="b905e2c4-4772-4914-b70f-44b4d19aca4b" x1="45.04" y1="231.72" x2="231.72" y2="45.04"
                        gradientUnits="userSpaceOnUse">
            <stop offset="0" stop-color="#fff"/>
            <stop offset="0" stop-color="#c2fde4"/>
            <stop offset="1" stop-color="#cdc6e9"/>
        </linearGradient>
    </defs>
    <circle class="a8bd947d-a1f8-4a14-8069-c9663505e4b1" cx="138.38" cy="138.38" r="132"/>
    <path id="bfa29dbe-bd65-4058-95bb-bdb844b1d3b1" data-name="Logo Pfad" class="a4ee1f07-8fee-4200-835a-e030a32bca90"
          d="M229.57,181.89l-4,24.19M47.29,182.88l14.81,14M157.86,59.23,62,196.69a.09.09,0,0,0,.07.15l163.41,9.37a.09.09,0,0,0,.09-.13L158,59.25A.1.1,0,0,0,157.86,59.23ZM138.12,45.94,47.18,182.69a.13.13,0,0,0,.11.19l182.18-.8a.12.12,0,0,0,.1-.19L138.33,45.94A.12.12,0,0,0,138.12,45.94Zm.11-.16L158,59.1"/>
</svg>`)

var labelIconSvg = []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24">
<path d="M0 0h24v24H0z" fill="none"/><path d="M17.63 5.84C17.27 5.33 16.67 5 16 5L5 5.01C3.9 5.01 3 5.9 3 7v10c0 1.1.9 1.99 2 1.99L16 19c.67 0 1.27-.33 1.63-.84L22 12l-4.37-6.16z"/></svg>`)

var brokenIconSvg = []byte(`
<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24">
<path fill="none" d="M0 0h24v24H0zm0 0h24v24H0zm21 19c0 1.1-.9 2-2 2H5c-1.1 0-2-.9-2-2V5c0-1.1.9-2 2-2h14c1.1 0 2 .9 2 2"/>
<path fill="none" d="M0 0h24v24H0z"/>
<path d="M21 5v6.59l-3-3.01-4 4.01-4-4-4 4-3-3.01V5c0-1.1.9-2 2-2h14c1.1 0 2 .9 2 2zm-3 6.42l3 3.01V19c0 1.1-.9 2-2 2H5c-1.1 0-2-.9-2-2v-6.58l3 2.99 4-4 4 4 4-3.99z"/></svg>`)

var uncachedIconSvg = []byte(`
<svg xmlns="http://www.w3.org/2000/svg" height="24" viewBox="0 0 24 24" width="24"><path d="M0 0h24v24H0z" fill="none"/>
<path d="M21 19V5c0-1.1-.9-2-2-2H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2zM8.5 13.5l2.5 3.01L14.5 12l4.5 6H5l3.5-4.5z"/></svg>`)

// GET /api/v1/svg/*
func GetSvg(router *gin.RouterGroup) {
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

	router.GET("/svg/album", func(c *gin.Context) {
		c.Data(http.StatusOK, "image/svg+xml", albumIconSvg)
	})

	router.GET("/svg/folder", func(c *gin.Context) {
		c.Data(http.StatusOK, "image/svg+xml", albumIconSvg)
	})

	router.GET("/svg/broken", func(c *gin.Context) {
		c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)
	})

	router.GET("/svg/uncached", func(c *gin.Context) {
		c.Data(http.StatusOK, "image/svg+xml", uncachedIconSvg)
	})
}
