package api

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/i18n"
)

const (
	UploadPath = "/upload"
)

// StartImport imports media files from a directory and converts/indexes them as needed.
//
// POST /api/v1/import*
func StartImport(router *gin.RouterGroup) {
	router.POST("/import/*path", func(c *gin.Context) {
		s := AuthAny(c, acl.ResourceFiles, acl.Permissions{acl.ActionManage, acl.ActionUpload})

		if s.Abort(c) {
			return
		}

		conf := get.Config()

		if conf.ReadOnly() || !conf.Settings().Features.Import {
			AbortFeatureDisabled(c)
			return
		}

		start := time.Now()

		var f form.ImportOptions

		// Assign and validate request form values.
		if err := c.BindJSON(&f); err != nil {
			AbortBadRequest(c)
			return
		}

		srcFolder := ""
		importPath := conf.ImportPath()

		// Import from subfolder?
		if srcFolder = c.Param("path"); srcFolder != "" && srcFolder != "/" {
			srcFolder = clean.UserPath(srcFolder)
		} else if f.Path != "" {
			srcFolder = clean.UserPath(f.Path)
		}

		// To avoid conflicts, uploads are imported from "import_path/upload/session_ref/timestamp".
		if token := path.Base(srcFolder); token != "" && path.Dir(srcFolder) == UploadPath {
			srcFolder = path.Join(UploadPath, s.RefID+token)
			event.AuditInfo([]string{ClientIP(c), "session %s", "import uploads from %s as %s", authn.Granted}, s.RefID, clean.Log(srcFolder), s.UserRole().String())
		} else if acl.Rules.Deny(acl.ResourceFiles, s.UserRole(), acl.ActionManage) {
			event.AuditErr([]string{ClientIP(c), "session %s", "import files from %s as %s", authn.Denied}, s.RefID, clean.Log(srcFolder), s.UserRole().String())
			AbortForbidden(c)
			return
		}

		importPath = path.Join(importPath, srcFolder)

		imp := get.Import()

		RemoveFromFolderCache(entity.RootImport)

		// Get destination folder.
		var destFolder string
		if destFolder = s.User().GetUploadPath(); destFolder == "" {
			destFolder = conf.ImportDest()
		}

		var opt photoprism.ImportOptions

		// Copy or move files to the destination folder?
		if f.Move {
			event.InfoMsg(i18n.MsgMovingFilesFrom, clean.Log(filepath.Base(importPath)))
			opt = photoprism.ImportOptionsMove(importPath, destFolder)
		} else {
			event.InfoMsg(i18n.MsgCopyingFilesFrom, clean.Log(filepath.Base(importPath)))
			opt = photoprism.ImportOptionsCopy(importPath, destFolder)
		}

		// Add imported files to albums if allowed.
		if len(f.Albums) > 0 &&
			acl.Rules.AllowAny(acl.ResourceAlbums, s.UserRole(), acl.Permissions{acl.ActionCreate, acl.ActionUpload}) {
			log.Debugf("import: adding files to album %s", clean.Log(strings.Join(f.Albums, " and ")))
			opt.Albums = f.Albums
		}

		// Set user UID if known.
		if s.UserUID != "" {
			opt.UID = s.UserUID
		}

		// Start import.
		imported := imp.Start(opt)

		// Delete empty import directory.
		if srcFolder != "" && importPath != conf.ImportPath() && fs.DirIsEmpty(importPath) {
			if err := os.Remove(importPath); err != nil {
				log.Errorf("import: failed to delete empty folder %s: %s", clean.Log(importPath), err)
			} else {
				log.Infof("import: deleted empty folder %s", clean.Log(importPath))
			}
		}

		// Update moments if files have been imported.
		if n := len(imported); n == 0 {
			log.Infof("import: found no new files to import from %s", clean.Log(importPath))
		} else {
			log.Infof("import: imported %s", english.Plural(n, "file", "files"))
			if moments := get.Moments(); moments == nil {
				log.Warnf("import: moments service not set - you may have found a bug")
			} else if err := moments.Start(); err != nil {
				log.Warnf("moments: %s", err)
			}
		}

		elapsed := int(time.Since(start).Seconds())

		// Show success message.
		msg := i18n.Msg(i18n.MsgImportCompletedIn, elapsed)

		event.Success(msg)

		eventData := event.Data{
			"uid":     opt.UID,
			"action":  opt.Action,
			"path":    importPath,
			"seconds": elapsed,
		}

		event.Publish("import.completed", eventData)
		event.Publish("index.completed", eventData)

		for _, uid := range f.Albums {
			PublishAlbumEvent(StatusUpdated, uid, c)
		}

		// Update the user interface.
		UpdateClientConfig()

		// Update album, label, and subject cover thumbs.
		if err := query.UpdateCovers(); err != nil {
			log.Warnf("index: %s (update covers)", err)
		}

		c.JSON(http.StatusOK, i18n.Response{Code: http.StatusOK, Msg: msg})
	})
}

// CancelImport stops the current import operation.
//
// DELETE /api/v1/import
func CancelImport(router *gin.RouterGroup) {
	router.DELETE("/import", func(c *gin.Context) {
		s := Auth(c, acl.ResourceFiles, acl.ActionManage)

		if s.Abort(c) {
			return
		}

		conf := get.Config()

		if conf.ReadOnly() || !conf.Settings().Features.Import {
			AbortFeatureDisabled(c)
			return
		}

		imp := get.Import()

		imp.Cancel()

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgImportCanceled))
	})
}
