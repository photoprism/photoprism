The following is the status of unit testing.
| Status | Path/Test |
| ------ | --------------------------------------------------------------------- |
| PASS | github.com/photoprism/photoprism/pkg/authn |
| PASS | github.com/photoprism/photoprism/pkg/capture |
| PASS | github.com/photoprism/photoprism/pkg/checksum |
| PASS | github.com/photoprism/photoprism/pkg/clean |
| PASS | github.com/photoprism/photoprism/pkg/clusters |
| PASS | github.com/photoprism/photoprism/pkg/env |
| PASS | github.com/photoprism/photoprism/pkg/fs |
| PASS | github.com/photoprism/photoprism/pkg/fs/fastwalk |
| PASS | github.com/photoprism/photoprism/pkg/geo |
| PASS | github.com/photoprism/photoprism/pkg/geo/pluscode |
| PASS | github.com/photoprism/photoprism/pkg/geo/s2 |
| PASS | github.com/photoprism/photoprism/pkg/header |
| PASS | github.com/photoprism/photoprism/pkg/i18n |
| PASS | github.com/photoprism/photoprism/pkg/list |
| PASS | github.com/photoprism/photoprism/pkg/log/dummy |
| PASS | github.com/photoprism/photoprism/pkg/log/level |
| PASS | github.com/photoprism/photoprism/pkg/media |
| PASS | github.com/photoprism/photoprism/pkg/media/colors |
| PASS | github.com/photoprism/photoprism/pkg/media/projection |
| PASS | github.com/photoprism/photoprism/pkg/media/video |
| PASS | github.com/photoprism/photoprism/pkg/react |
| PASS | github.com/photoprism/photoprism/pkg/rnd |
| PASS | github.com/photoprism/photoprism/pkg/time/unix |
| PASS | github.com/photoprism/photoprism/pkg/txt |
| PASS | github.com/photoprism/photoprism/pkg/txt/clip |
| PASS | github.com/photoprism/photoprism/pkg/txt/report |
| PASS | github.com/photoprism/photoprism/internal/ai/classify |
| PASS | github.com/photoprism/photoprism/internal/ai/face |
| PASS | github.com/photoprism/photoprism/internal/ai/nsfw |
| PASS | github.com/photoprism/photoprism/internal/auth/acl |
| PASS | github.com/photoprism/photoprism/internal/auth/oidc |
| PASS | github.com/photoprism/photoprism/internal/auth/session |
| PASS | github.com/photoprism/photoprism/internal/config/customize |
| PASS | github.com/photoprism/photoprism/internal/config/pwa |
| PASS | github.com/photoprism/photoprism/internal/config/ttl |
| PASS | github.com/photoprism/photoprism/internal/entity/dbtest |
| PASS | github.com/photoprism/photoprism/internal/entity/migrate |
| PASS | github.com/photoprism/photoprism/internal/entity/search/viewer |
| PASS | github.com/photoprism/photoprism/internal/entity/sortby |
| PASS | github.com/photoprism/photoprism/internal/event |
| PASS | github.com/photoprism/photoprism/internal/ffmpeg |
| PASS | github.com/photoprism/photoprism/internal/form |
| PASS | github.com/photoprism/photoprism/internal/functions |
| PASS | github.com/photoprism/photoprism/internal/meta |
| PASS | github.com/photoprism/photoprism/internal/mutex |
| PASS | github.com/photoprism/photoprism/internal/photoprism/get |
| PASS | github.com/photoprism/photoprism/internal/server |
| PASS | github.com/photoprism/photoprism/internal/server/limiter |
| PASS | github.com/photoprism/photoprism/internal/server/wellknown |
| PASS | github.com/photoprism/photoprism/internal/service |
| PASS | github.com/photoprism/photoprism/internal/service/hub |
| PASS | github.com/photoprism/photoprism/internal/service/hub/places |
| PASS | github.com/photoprism/photoprism/internal/service/maps |
| PASS | github.com/photoprism/photoprism/internal/service/webdav |
| PASS | github.com/photoprism/photoprism/internal/thumb |
| PASS | github.com/photoprism/photoprism/internal/thumb/avatar |
| PASS | github.com/photoprism/photoprism/internal/thumb/crop |
| PASS | github.com/photoprism/photoprism/internal/thumb/frame |
| PASS | github.com/photoprism/photoprism/internal/workers |
| PASS | github.com/photoprism/photoprism/internal/workers/auto |


| Status | Path/Test |
| ------ | --------------------------------------------------------------------- |
| FAIL | BatchPhotosArchive|
| FAIL |   BatchPhotosArchive/successful_request|
| FAIL | BatchPhotosRestore|
| FAIL |   BatchPhotosRestore/successful_request|
| FAIL | BatchAlbumsDelete|
| FAIL |   BatchAlbumsDelete/successful_request|
| FAIL | BatchPhotosPrivate|
| FAIL |   BatchPhotosPrivate/successful_request|
| FAIL | BatchPhotosApprove|
| FAIL |   BatchPhotosApprove/successful_request|
| FAIL | RemovePhotoLabel|
| FAIL |   RemovePhotoLabel/photo_with_label|
| FAIL |   RemovePhotoLabel/remove_manually_added_label|
| FAIL | UpdatePhotoLabel|
| FAIL |   UpdatePhotoLabel/successful_request|
| FAIL |   UpdatePhotoLabel/bad_request|
| FAIL | GetSubject|
| FAIL |   GetSubject/Ok|
| FAIL | LikeSubject|
| FAIL |   LikeSubject/ExistingSubject|
| FAIL | 	github.com/photoprism/photoprism/internal/api|
| FAIL | UsersCommand|
| FAIL |   UsersCommand/AddModifyAndRemoveJohn|
| FAIL | 	github.com/photoprism/photoprism/internal/commands|
| FAIL | Config_DatabaseDsn|
| FAIL | Config_SqliteBin|
| FAIL | 	github.com/photoprism/photoprism/internal/config|
| FAIL | FirstOrCreateUser|
| FAIL |   FirstOrCreateUser/Existing|
| FAIL | Count|
| FAIL | LabelCounts|
| FAIL | UpdateCounts|
| FAIL | Keyword_Updates|
| FAIL |   Keyword_Updates/success|
| FAIL | Keyword_Update|
| FAIL |   Keyword_Update/success|
| FAIL | Marker_ClearFace|
| FAIL |   Marker_ClearFace/empty_face_id|
| FAIL | PhotoAlbum_Save|
| FAIL |   PhotoAlbum_Save/success|
| FAIL | Photo_UpdateQuality|
| FAIL |   Photo_UpdateQuality/low_quality_expected|
| FAIL | Photo_AddLabels|
| FAIL |   Photo_AddLabels/Add|
| FAIL |   Photo_AddLabels/Update|
| FAIL | 	github.com/photoprism/photoprism/internal/entity|
| FAIL | UpdateAlbumDefaultCovers|
| FAIL | UpdateAlbumFolderCovers|
| FAIL | UpdateAlbumMonthCovers|
| FAIL | UpdateAlbumCovers|
| FAIL | UpdateLabelCovers|
| FAIL | UpdateSubjectCovers|
| FAIL | UpdateCovers|
| FAIL | FileSelection|
| FAIL |   FileSelection/ShareSelectionOriginals|
| FAIL |   FileSelection/ShareFolders|
| FAIL | FilesByUID|
| FAIL |   FilesByUID/error|
| FAIL | LabelThumbBySlug|
| FAIL |   LabelThumbBySlug/file_found|
| FAIL | LabelThumbByUID|
| FAIL |   LabelThumbByUID/file_found|
| FAIL | PhotoLabel|
| FAIL |   PhotoLabel/photo_label_found|
| FAIL | MomentsCategories|
| FAIL |   MomentsCategories/PublicOnly|
| FAIL |   MomentsCategories/IncludePrivate|
| FAIL | CountUsers|
| FAIL |   CountUsers/All|
| FAIL |   CountUsers/Registered|
| FAIL |   CountUsers/Active|
| FAIL |   CountUsers/RegisteredActive|
| FAIL |   CountUsers/NoAdmins|
| FAIL | 	github.com/photoprism/photoprism/internal/entity/query|
| FAIL | PhotosFilterFilter|
| FAIL |   PhotosFilterFilter/CenterPercent|
| FAIL | PhotosQueryFilter|
| FAIL |   PhotosQueryFilter/CenterPercent|
| FAIL | PhotosFilterLabel|
| FAIL |   PhotosFilterLabel/flower|
| FAIL |   PhotosFilterLabel/cake|
| FAIL |   PhotosFilterLabel/cake_pipe_flower|
| FAIL |   PhotosFilterLabel/cake_whitespace_pipe_whitespace_flower|
| FAIL |   PhotosFilterLabel/StartsWithNumber|
| FAIL |   PhotosFilterLabel/CenterNumber|
| FAIL |   PhotosFilterLabel/EndsWithNumber|
| FAIL |   PhotosFilterLabel/OrSearch|
| FAIL | PhotosQueryLabel|
| FAIL |   PhotosQueryLabel/flower|
| FAIL |   PhotosQueryLabel/cake|
| FAIL |   PhotosQueryLabel/cake_pipe_flower|
| FAIL |   PhotosQueryLabel/cake_whitespace_pipe_whitespace_flower|
| FAIL |   PhotosQueryLabel/StartsWithNumber|
| FAIL |   PhotosQueryLabel/CenterNumber|
| FAIL |   PhotosQueryLabel/EndsWithNumber|
| FAIL |   PhotosQueryLabel/OrSearch|
| FAIL | Photos|
| FAIL |   Photos/label_query_landscape|
| FAIL |   Photos/search_for_label_in_query|
| FAIL |   Photos/search_for_labels|
| FAIL |   Photos/search_for_landscape|
| FAIL | 	github.com/photoprism/photoprism/internal/entity/search|
| FAIL | Places|
| FAIL |   Places/Unresolved|
| FAIL |   Places/Force|
| FAIL | 	github.com/photoprism/photoprism/internal/photoprism|
| FAIL | Albums|
| FAIL | 	github.com/photoprism/photoprism/internal/photoprism/backup|


Issues found in tests that don't cause a failure:
