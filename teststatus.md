GORM v2 has introduced foreign keys to the database.  
This has caused a number of the tests to no longer function.  
The fundamental issue is that these tests worked in the past as GORM v1 did not check if the parent record existed before saving the child record.  
eg.  
You could save a PhotoLabel with random numbers as the PhotoID and LabelID without an error being raised.  This is no longer possible.  


**Tests changed so that they work with GORM v2**

All the TestMains that utilise the database have been changed so they have a database hosted MUTEX.  
This is to prevent 2 or more sets of database tests running at the same time.
When this happens then 2nd or subsequent test will truncate all the data as they other test(s) are running causing random test failures.  
In testing up to four separate testing threads have attempted to run against the database at the same time using the makefile.  
The issue is the requirement to clear and refresh the unit test data so each suite of tests work correctly.  


Have to use Create() instead of Save() as save doesn't call BeforeCreate in v2.  
| File | Test |
|----------------------------------------|---------------------------------------------|  
| internal/entity/photo_test.go | TestPhoto_SyncKeywordLabels/Ok |

Have to create records in the database or the tested function will fail due to foreign key violations  
| File | Test |
|----------------------------------------|---------------------------------------------|
| internal/entity/photo_label_test.go | TestPhotoLabel_Save/success |
| internal/entity/photo_label_test.go | TestPhotoLabel_Save/photo not nil and label not nil |
| internal/entity/photo_album_test.go | TestFirstOrCreatePhotoAlbum/not yet existing photo_album |
| internal/entity/photo_album_test.go | TestPhotoAlbum_Save/success |
| internal/entity/file_test.go | TestFile_Create/file already exists |
| internal/entity/file_test.go | TestFile_Update/success |
| internal/entity/file_test.go | TestFile_Delete/permanently |
| internal/entity/file_test.go | TestFile_Delete/not permanently |
| internal/entity/file_sync_test.go | TestNewFileSync |
| internal/entity/file_sync_test.go | TestFirstOrCreateFileSync/not yet existing |
| internal/entity/file_sync_test.go | TestFileSync_Updates/success |
| internal/entity/file_sync_test.go | TestFileSync_Update/success |
| internal/entity/file_sync_test.go | TestFileSync_Save/success |
| internal/entity/file_share_test.go | TestFirstOrCreateFileShare/not yet existing |
| internal/entity/file_share_test.go | TestFirstOrCreateFileShare/existing |
| internal/entity/file_share_test.go | TestFileShare_Updates/success |
| internal/entity/file_share_test.go | TestFileShare_Update/success |
| internal/entity/file_share_test.go | TestFileShare_Save/success |
| internal/entity/auth_user_share_test.go | TestUserShare_Create |
| internal/entity/auth_user_settings_test.go | TestCreateUserSettings/Success |
| internal/entity/auth_user_details.go | TestCreateUserDetails/Success |



Have to populate extra fields or the tested function will fail due to foreign key violations  
| File | Test |
|----------------------------------------|---------------------------------------------|
| internal/entity/keyword_test.go | TestMarker_ClearFace/empty face id |
| internal/entity/file_sync_test.go | TestFirstOrCreateFileSync/existing |


Have to provide an ID value or tested function will fail with Where clause missing  
| File | Test |
|----------------------------------------|---------------------------------------------|
| internal/entity/keyword_test.go | TestKeyword_Update/success |


Specials  
| File | Test | Description |
|----------------------------------------|---------------------------------------------|------------------------------------------------------|
| internal/entity/search/photos_filter_filter_test.go | TestPhotosFilterFilter/CenterPercent | Soft delete a record that was "hidden" due to duplicate ID values in Fixture |
| internal/entity/search/photos_filter_filter_test.go | TestPhotosQueryFilter/CenterPercent | Soft delete a record that was "hidden" due to duplicate ID values in Fixture |
| internal/entity/query/moments_test.go | TestRemoveDuplicateMoments/Ok | sqlite issue on GORMv1 which hasn't shown up on GORMv2 |
| internal/entity/query/files_test.go | TestFilesByUID/error | GORMv1 vs GORMv2 differences |
| internal/entity/query/file_selection_test.go | TestFileSelection/ShareSelectionOriginals | Not sure why MediaType is empty string on GORMv1 as it shouldn't be, so force it to ensure test works.  See Filefixture which sets file name with .jpg and BeforeCreate which sets the MediaType. |

New Tests
| File | Test | Description |
|----------------------------------------|---------------------------------------------|------------------------------------------------------|
| internal/entity/keyword_test.go | TestKeyword_UpdateNoID/success | Validates missing PK error for Update |
| internal/entity/marker_test.go | TestMarker_ClearFace/missing markeruid | Validates missing PK error for Update |
| internal/entity/auth_user_test.go | TestUser_ValidatePreload/* | Validates that Preload is used to get child attributes |
| internal/entity/query/files_test.go | TestFilesByUID/Negative limit with offset | Validates limits and offsets |
| internal/entity/query/files_test.go | TestFilesByUID/offset and limit | Validates limits and offsets |
| internal/entity/dbtest/dbtest_init_test.go | TestInit/* | checks that the number of records in a fresh database is correct |
| internal/entity/dbtest/dbtest_foreignkey_test.go | TestDbtestForeignKey_Validate/Photos_CameraID | makes sure that the database throws a foreign key error |
| internal/entity/dbtest/dbtest_foreignkey_test.go | TestDbtestForeignKey_Validate/Photos_LensID | makes sure that the database throws a foreign key error |
| internal/entity/dbtest/dbtest_fieldlength_test.go | TestInitDBLengths/PhotoMaxVarLengths | makes sure that the database can hold specified maximum length of each column in a Photo |
| internal/entity/dbtest/dbtest_fieldlength_test.go | TestInitDBLengths/PhotoExceedMax* | makes sure that the database throws an error when the maximum length is exceeded by 1 character in a Photo |



The following is the status of unit testing against sqlite.  

Removing test database files...  
find ./internal -type f -name ".test.*" -delete  
Running all Go tests...  
richgo test -parallel 1 -count 1 -cpu 1 -tags slow -timeout 20m ./pkg/... ./internal/...  

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
| SKIP | github.com/photoprism/photoprism/internal/entity/legacy	[no test files] |
| PASS | github.com/photoprism/photoprism/internal/api |
| PASS | github.com/photoprism/photoprism/internal/auth/acl |
| PASS | github.com/photoprism/photoprism/internal/auth/oidc |
| PASS | github.com/photoprism/photoprism/internal/auth/session |
| PASS | github.com/photoprism/photoprism/internal/commands |
| PASS | github.com/photoprism/photoprism/internal/config |
| PASS | github.com/photoprism/photoprism/internal/config/customize |
| PASS | github.com/photoprism/photoprism/internal/config/pwa |
| PASS | github.com/photoprism/photoprism/internal/config/ttl |
| PASS | github.com/photoprism/photoprism/internal/entity |
| PASS | github.com/photoprism/photoprism/internal/entity/dbtest |
| PASS | github.com/photoprism/photoprism/internal/entity/migrate |
| PASS | github.com/photoprism/photoprism/internal/entity/query |
| PASS | github.com/photoprism/photoprism/internal/entity/search |
| PASS | github.com/photoprism/photoprism/internal/entity/search/viewer |
| PASS | github.com/photoprism/photoprism/internal/entity/sortby |
| PASS | github.com/photoprism/photoprism/internal/event |
| PASS | github.com/photoprism/photoprism/internal/ffmpeg |
| PASS | github.com/photoprism/photoprism/internal/form |
| PASS | github.com/photoprism/photoprism/internal/functions |
| PASS | github.com/photoprism/photoprism/internal/meta |
| PASS | github.com/photoprism/photoprism/internal/mutex |
| SKIP | github.com/photoprism/photoprism/internal/testextras	[no test files] |
| PASS | github.com/photoprism/photoprism/internal/photoprism |
| PASS | github.com/photoprism/photoprism/internal/photoprism/backup |
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



The following is the status of unit testing against mariadb, which drops the database as part of the init.  
Resetting acceptance database...  
mysql < scripts/sql/reset-acceptance.sql  
Running all Go tests on MariaDB...
PHOTOPRISM_TEST_DRIVER="mysql" PHOTOPRISM_TEST_DSN="root:photoprism@tcp(mariadb:4001)/acceptance?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true" richgo test -parallel 1 -count 1 -cpu 1 -tags slow -timeout 20m ./pkg/... ./internal/...  
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
| PASS | github.com/photoprism/photoprism/internal/api |
| PASS | github.com/photoprism/photoprism/internal/auth/acl |
| PASS | github.com/photoprism/photoprism/internal/auth/oidc |
| PASS | github.com/photoprism/photoprism/internal/auth/session |
| SKIP | github.com/photoprism/photoprism/internal/entity/legacy	[no test files] |
| PASS | github.com/photoprism/photoprism/internal/commands |
| PASS | github.com/photoprism/photoprism/internal/config |
| PASS | github.com/photoprism/photoprism/internal/config/customize |
| PASS | github.com/photoprism/photoprism/internal/config/pwa |
| PASS | github.com/photoprism/photoprism/internal/config/ttl |
| PASS | github.com/photoprism/photoprism/internal/entity |
| PASS | github.com/photoprism/photoprism/internal/entity/dbtest |
| PASS | github.com/photoprism/photoprism/internal/entity/migrate |
| SKIP | github.com/photoprism/photoprism/internal/testextras	[no test files] |
| PASS | github.com/photoprism/photoprism/internal/entity/query |
| PASS | github.com/photoprism/photoprism/internal/entity/search |
| PASS | github.com/photoprism/photoprism/internal/entity/search/viewer |
| PASS | github.com/photoprism/photoprism/internal/entity/sortby |
| PASS | github.com/photoprism/photoprism/internal/event |
| PASS | github.com/photoprism/photoprism/internal/ffmpeg |
| PASS | github.com/photoprism/photoprism/internal/form |
| PASS | github.com/photoprism/photoprism/internal/functions |
| PASS | github.com/photoprism/photoprism/internal/meta |
| PASS | github.com/photoprism/photoprism/internal/mutex |
| PASS | github.com/photoprism/photoprism/internal/photoprism |
| PASS | github.com/photoprism/photoprism/internal/photoprism/backup |
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

