package api

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/txt"
)

// GET /api/v1/:fileHash/duplicates
//
// Parameters:
//   fileHash: string Duplicate File Hash
func GetDuplicates(router *gin.RouterGroup) {
	router.GET("/duplicates/:fileHash", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceLabels, acl.ActionSearch)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		result, err := query.GetDuplicatesByHash(c.Param("fileHash"))

		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		AddTokenHeaders(c)

		c.JSON(http.StatusOK, result)
	})
}

// DELETE /api/v1/:fileName/duplicates
//
// Parameters:
//   fileName: string Duplicate File Name
func DeleteDuplicate(router *gin.RouterGroup) {
	router.DELETE("/duplicates/:fileName", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceLabels, acl.ActionUpdate)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		fileName := c.Param("fileName")
		baseName := filepath.Base(fileName)
		dup, err := query.GetDuplicateByName(fileName)
		
		if err != nil {
			log.Errorf("duplicate: %s (load %s from duplicates)", err, txt.Quote(baseName))
			AbortEntityNotFound(c)
			return
		}

		if err := dup.Purge(); err != nil {
			log.Errorf("duplicate: %s (delete %s from duplicates)", err, txt.Quote(baseName))
			AbortDeleteFailed(c)
			return
		}

		// PublishDuplicateEvent(EntityUpdated, name, c)

		event.SuccessMsg(i18n.MsgFileDeleted)

		c.JSON(http.StatusOK, http.Response{})
	})
}
