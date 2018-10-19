package api

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	filetype "gopkg.in/h2non/filetype.v1"

	"github.com/photoprism/photoprism/internal/forms"
	"github.com/photoprism/photoprism/internal/photoprism"
)

type PhotosResponse {
	message: interface{}
}

func GetPhotos(router *gin.RouterGroup, conf *photoprism.Config) {
	router.GET("/photos", func(c *gin.Context) {
		var form forms.PhotoSearchForm

		search := photoprism.NewSearch(conf.OriginalsPath, conf.GetDb())

		c.MustBindWith(&form, binding.Form)

		result, err := search.Photos(form)

		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		}

		c.Header("x-result-count", strconv.Itoa(form.Count))
		c.Header("x-result-offset", strconv.Itoa(form.Offset))

		c.JSON(http.StatusOK, result)
	})

	router.POST("/photos", func(c *gin.Context) {
		file, _, err := c.Request.FormFile("file")
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("unable to accept file: %s", err.Error()))
			return
		}
		defer file.Close()

		buff := bytes.NewBuffer(nil)
		if _, err := io.Copy(buff, file); err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("unable to copy into memory: %s", err.Error()))
			return
		}

		if !filetype.IsImage(buff.Bytes()) {
			c.String(http.StatusBadRequest, fmt.Sprintf("invalid file"))
			return
		}

		name, err := uuid.NewUUID()
		if err != nil {
			c.String(http.StatusInternalServerError, "unable to generate uuid")
			return
		}

		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			c.String(http.StatusInternalServerError, "unable to file path")
			return
		}

		filePath := path.Join(dir, "assets", "photos", "temp", name.String())
		if err := ioutil.WriteFile(filePath, buff.Bytes(), 0644); err != nil {
			c.String(http.StatusInternalServerError, "unable to store file %s", err)
			return
		}
	})
}
