package api

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// PhotoUnstack removes a file from an existing photo stack.
//
// The request parameters are:
//
//   - uid: string Photo UID as returned by the API
//   - file_uid: string File UID as returned by the API
//
// POST /api/v1/photos/:uid/files/:file_uid/unstack
func PhotoUnstack(router *gin.RouterGroup) {
	router.POST("/photos/:uid/files/:file_uid/unstack", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		conf := get.Config()
		fileUid := clean.UID(c.Param("file_uid"))
		file, err := query.FileByUID(fileUid)

		if err != nil {
			log.Errorf("photo: %s (unstack)", err)
			AbortEntityNotFound(c)
			return
		}

		if file.FilePrimary {
			log.Errorf("photo: cannot unstack primary file")
			AbortBadRequest(c)
			return
		} else if file.FileSidecar {
			log.Errorf("photo: cannot unstack sidecar files")
			AbortBadRequest(c)
			return
		} else if file.FileRoot != entity.RootOriginals {
			log.Errorf("photo: only originals can be unstacked")
			AbortBadRequest(c)
			return
		}

		fileName := photoprism.FileName(file.FileRoot, file.FileName)
		baseName := filepath.Base(fileName)

		unstackFile, err := photoprism.NewMediaFile(fileName)

		if err != nil {
			log.Errorf("photo: %s (unstack %s)", err, clean.Log(baseName))
			AbortEntityNotFound(c)
			return
		} else if file.Photo == nil {
			log.Errorf("photo: cannot find photo for file uid %s (unstack)", fileUid)
			AbortEntityNotFound(c)
			return
		}

		stackPhoto := *file.Photo
		createdBy := stackPhoto.CreatedBy
		stackPrimary, err := stackPhoto.PrimaryFile()

		if err != nil {
			log.Errorf("photo: cannot find primary file for %s (unstack)", clean.Log(baseName))
			AbortUnexpectedError(c)
			return
		}

		// Flag original photo as unstacked / not stackable.
		stackPhoto.SetStack(entity.IsUnstacked)

		related, err := unstackFile.RelatedFiles(false)

		if err != nil {
			log.Errorf("photo: %s (unstack %s)", err, clean.Log(baseName))
			AbortEntityNotFound(c)
			return
		} else if related.Len() == 0 {
			log.Errorf("photo: found no files for %s (unstack)", clean.Log(baseName))
			AbortEntityNotFound(c)
			return
		} else if related.Main == nil {
			log.Errorf("photo: found no main media file for %s (unstack)", clean.Log(baseName))
			AbortEntityNotFound(c)
			return
		}

		var files photoprism.MediaFiles
		unstackSingle := false

		if unstackFile.BasePrefix(false) == stackPhoto.PhotoName {
			if conf.ReadOnly() {
				log.Errorf("photo: cannot rename files in read only mode (unstack %s)", clean.Log(baseName))
				AbortFeatureDisabled(c)
				return
			}

			destName := fmt.Sprintf("%s.%s%s", unstackFile.AbsPrefix(false), unstackFile.Checksum(), unstackFile.Extension())

			if err := unstackFile.Move(destName); err != nil {
				log.Errorf("photo: cannot rename %s to %s (unstack)", clean.Log(unstackFile.BaseName()), clean.Log(filepath.Base(destName)))
				AbortUnexpectedError(c)
				return
			}

			files = append(files, unstackFile)
			unstackSingle = true
		} else {
			files = related.Files
		}

		// Create new photo, also flagged as unstacked / not stackable.
		newPhoto := entity.NewUserPhoto(false, createdBy)
		newPhoto.PhotoPath = unstackFile.RootRelPath()
		newPhoto.PhotoName = unstackFile.BasePrefix(false)

		if err := newPhoto.Create(); err != nil {
			log.Errorf("photo: %s (unstack %s)", err.Error(), clean.Log(baseName))
			AbortSaveFailed(c)
			return
		}

		for _, r := range files {
			relName := r.RootRelName()
			relRoot := r.Root()

			if unstackSingle {
				relName = file.FileName
				relRoot = file.FileRoot
			}

			if err := entity.UnscopedDb().Exec(`UPDATE files 
				SET photo_id = ?, photo_uid = ?, file_name = ?, file_missing = 0
				WHERE file_name = ? AND file_root = ?`,
				newPhoto.ID, newPhoto.PhotoUID, r.RootRelName(),
				relName, relRoot).Error; err != nil {
				// Handle error...
				log.Errorf("photo: %s (unstack %s)", err.Error(), clean.Log(r.BaseName()))

				// Remove new photo from index.
				if _, err := newPhoto.Delete(true); err != nil {
					log.Errorf("photo: %s (unstack %s)", err.Error(), clean.Log(r.BaseName()))
				}

				// Revert file rename.
				if unstackSingle {
					if err := r.Move(photoprism.FileName(relRoot, relName)); err != nil {
						log.Errorf("photo: %s (unstack %s)", err.Error(), clean.Log(r.BaseName()))
					}
				}

				AbortSaveFailed(c)
				return
			}
		}

		ind := get.Index()

		// Index unstacked files.
		if res := ind.FileName(unstackFile.FileName(), photoprism.IndexOptionsSingle()); res.Failed() {
			log.Errorf("photo: %s (unstack %s)", res.Err, clean.Log(baseName))
			AbortSaveFailed(c)
			return
		}

		// Reset type for existing photo stack to image.
		if err := stackPhoto.Update("PhotoType", entity.MediaImage); err != nil {
			log.Errorf("photo: %s (unstack %s)", err, clean.Log(baseName))
			AbortUnexpectedError(c)
			return
		}

		// Re-index existing photo stack.
		if res := ind.FileName(photoprism.FileName(stackPrimary.FileRoot, stackPrimary.FileName), photoprism.IndexOptionsSingle()); res.Failed() {
			log.Errorf("photo: %s (unstack %s)", res.Err, clean.Log(baseName))
			AbortSaveFailed(c)
			return
		}

		// Notify clients by publishing events.
		PublishPhotoEvent(StatusCreated, newPhoto.PhotoUID, c)
		PublishPhotoEvent(StatusUpdated, stackPhoto.PhotoUID, c)

		event.SuccessMsg(i18n.MsgFileUnstacked)

		p, err := query.PhotoPreloadByUID(stackPhoto.PhotoUID)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		c.JSON(http.StatusOK, p)
	})
}
