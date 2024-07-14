package server

import (
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/api"
	"github.com/photoprism/photoprism/internal/config"
)

var APIv1 *gin.RouterGroup
var registerApiDocs func(router *gin.RouterGroup)

// registerRoutes registers the routes for handling HTTP requests with the built-in web server.
func registerRoutes(router *gin.Engine, conf *config.Config) {
	// Enables automatic redirection if the current route cannot be matched but a
	// handler for the path with (without) the trailing slash exists.
	router.RedirectTrailingSlash = true

	// Register static asset and templates routes.
	registerStaticRoutes(router, conf)

	// Register PWA bootstrap and config routes.
	registerPWARoutes(router, conf)

	// Register built-in WebDAV server routes.
	registerWebDAVRoutes(router, conf)

	// Register sharing routes starting with "/s".
	registerSharingRoutes(router, conf)

	// Register ".well-known" service discovery routes.
	registerWellknownRoutes(router, conf)

	// Register JSON REST-API version 1 (APIv1) routes, grouped by functionality.
	// Docs: https://pkg.go.dev/github.com/photoprism/photoprism/internal/api

	// API Documentation.
	if registerApiDocs != nil {
		registerApiDocs(APIv1)
	}

	// User Sessions.
	api.CreateSession(APIv1)
	api.GetSession(APIv1)
	api.DeleteSession(APIv1)

	// OAuth2 Client Endpoints.
	api.OAuthAuthorize(APIv1)
	api.OAuthUserinfo(APIv1)
	api.OAuthToken(APIv1)
	api.OAuthRevoke(APIv1)

	// OIDC Client Endpoints.
	api.OIDCLogin(APIv1)
	api.OIDCRedirect(APIv1)

	// Global Configuration.
	api.GetConfigOptions(APIv1)
	api.SaveConfigOptions(APIv1)
	api.StopServer(APIv1)

	// User Settings.
	api.GetClientConfig(APIv1)
	api.GetSettings(APIv1)
	api.SaveSettings(APIv1)

	// User Profile and Uploads.
	api.UploadUserFiles(APIv1)
	api.ProcessUserUpload(APIv1)
	api.UploadUserAvatar(APIv1)
	api.FindUserSessions(APIv1)
	api.CreateUserPasscode(APIv1)
	api.ConfirmUserPasscode(APIv1)
	api.ActivateUserPasscode(APIv1)
	api.DeactivateUserPasscode(APIv1)
	api.UpdateUserPassword(APIv1)
	api.UpdateUser(APIv1)

	// Service Accounts.
	api.SearchServices(APIv1)
	api.GetService(APIv1)
	api.GetServiceFolders(APIv1)
	api.UploadToService(APIv1)
	api.AddService(APIv1)
	api.DeleteService(APIv1)
	api.UpdateService(APIv1)

	// Thumbnail Images.
	api.GetThumb(APIv1)

	// Video Streaming.
	api.GetVideo(APIv1)

	// Downloads.
	api.GetDownload(APIv1)
	api.ZipCreate(APIv1)
	api.ZipDownload(APIv1)

	// Index and Import.
	api.StartImport(APIv1)
	api.CancelImport(APIv1)
	api.StartIndexing(APIv1)
	api.CancelIndexing(APIv1)

	// Photo Search and Organization.
	api.SearchPhotos(APIv1)
	api.SearchGeo(APIv1)
	api.GetPhoto(APIv1)
	api.GetPhotoYaml(APIv1)
	api.UpdatePhoto(APIv1)
	api.GetPhotoDownload(APIv1)
	// api.GetPhotoLinks(APIv1)
	// api.CreatePhotoLink(APIv1)
	// api.UpdatePhotoLink(APIv1)
	// api.DeletePhotoLink(APIv1)
	api.ApprovePhoto(APIv1)
	api.LikePhoto(APIv1)
	api.DislikePhoto(APIv1)
	api.AddPhotoLabel(APIv1)
	api.RemovePhotoLabel(APIv1)
	api.UpdatePhotoLabel(APIv1)
	api.GetMomentsTime(APIv1)
	api.GetFile(APIv1)
	api.DeleteFile(APIv1)
	api.ChangeFileOrientation(APIv1)
	api.CreateMarker(APIv1)
	api.UpdateMarker(APIv1)
	api.ClearMarkerSubject(APIv1)
	api.PhotoPrimary(APIv1)
	api.PhotoUnstack(APIv1)

	// Photo Albums.
	api.SearchAlbums(APIv1)
	api.GetAlbum(APIv1)
	api.AlbumCover(APIv1)
	api.CreateAlbum(APIv1)
	api.UpdateAlbum(APIv1)
	api.DeleteAlbum(APIv1)
	api.DownloadAlbum(APIv1)
	api.GetAlbumLinks(APIv1)
	api.CreateAlbumLink(APIv1)
	api.UpdateAlbumLink(APIv1)
	api.DeleteAlbumLink(APIv1)
	api.LikeAlbum(APIv1)
	api.DislikeAlbum(APIv1)
	api.CloneAlbums(APIv1)
	api.AddPhotosToAlbum(APIv1)
	api.RemovePhotosFromAlbum(APIv1)

	// Photo Labels.
	api.SearchLabels(APIv1)
	api.LabelCover(APIv1)
	api.UpdateLabel(APIv1)
	// api.GetLabelLinks(APIv1)
	// api.CreateLabelLink(APIv1)
	// api.UpdateLabelLink(APIv1)
	// api.DeleteLabelLink(APIv1)
	api.LikeLabel(APIv1)
	api.DislikeLabel(APIv1)

	// Files and Folders.
	api.SearchFoldersOriginals(APIv1)
	api.SearchFoldersImport(APIv1)
	api.FolderCover(APIv1)

	// People.
	api.SearchSubjects(APIv1)
	api.GetSubject(APIv1)
	api.UpdateSubject(APIv1)
	api.LikeSubject(APIv1)
	api.DislikeSubject(APIv1)

	// Faces.
	api.SearchFaces(APIv1)
	api.GetFace(APIv1)
	api.UpdateFace(APIv1)

	// Batch Operations.
	api.BatchPhotosApprove(APIv1)
	api.BatchPhotosArchive(APIv1)
	api.BatchPhotosRestore(APIv1)
	api.BatchPhotosPrivate(APIv1)
	api.BatchPhotosDelete(APIv1)
	api.BatchAlbumsDelete(APIv1)
	api.BatchLabelsDelete(APIv1)

	// Technical Endpoints.
	api.GetSvg(APIv1)
	api.GetStatus(APIv1)
	api.GetErrors(APIv1)
	api.DeleteErrors(APIv1)
	api.SendFeedback(APIv1)
	api.Connect(APIv1)
	api.WebSocket(APIv1)
	api.GetMetrics(APIv1)
	api.Echo(APIv1)
}
