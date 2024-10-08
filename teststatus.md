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

The TestMains have an additional check for error records in the Errors table, and will mark the test suite as failed if any new records are reported.  
This is to ensure that the checking that has been implemented in Photo.Save (and will be implemented in other Save's as appropriate) hasn't found any scenarios where mismatches between the in Struct ID field and the sub Struct's ID field are detected.  


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



Have to populate extra fields or the tested function will fail (or pass as the errors aren't checked in the test) due to foreign key violations  
| File | Test |
|----------------------------------------|---------------------------------------------|
| internal/entity/keyword_test.go | TestMarker_ClearFace/empty face id |
| internal/entity/file_sync_test.go | TestFirstOrCreateFileSync/existing |
| internal/entity/auth_user_details_test.go | TestUserDetails_Updates |
| internal/entity/auth_user_settings_test.go | TestUserSettings_Updates |
| internal/entity/file_test.go | TestFile_Undelete/success |
| internal/entity/file_test.go | TestFile_Undelete/file not missing |



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
| internal/entity/entity_update_test.go | TestEntitiy_Update/Photo01 | add checking that the Camera isn't removed |



Fixed Tests
| File | Test | Description |
|----------------------------------------|---------------------------------------------|------------------------------------------------------|
| internal/entity/auth_user_details_test.go | TestUserDetails_Updates | Validate that no errors are returned |
| internal/entity/auth_user_settings_test.go | TestUserSettings_Updates | Validate that no errors are returned |
| internal/entity/photo_test.go | TestPhoto_ClassifyLabels/NewPhoto | Use a new photo struct, and load data correctly.  (Name of this test is misleading) |
| internal/entity/photo_test.go | TestPhoto_ClassifyLabels/ExistingPhoto | Use a new photo struct, and load data correctly. (Name of this test is misleading) |



New Tests
| File | Test | Description |
|----------------------------------------|---------------------------------------------|------------------------------------------------------|
| internal/entity/keyword_test.go | TestKeyword_UpdateNoID/success | Validates "id value required but not provided" error for Update |
| internal/entity/keyword_test.go | TestKeyword_Updates/success ID on keyword | Validate that Update saves a Keyword when the ID is in the struct |
| internal/entity/keyword_test.go | TestKeyword_Updates/failure | Validate that Update fails when an ID is not in the struct or the request |
| internal/entity/marker_test.go | TestMarker_ClearFace/missing markeruid | Validates "markeruid required but not provided" error for Update |
| internal/entity/marker_test.go | TestMarker_Matched/missing markeruid | Validates "markeruid required but not provided" error for Update |
| internal/entity/auth_user_test.go | TestUser_ValidatePreload/* | Validates that Preload is used to get child attributes |
| internal/entity/query/files_test.go | TestFilesByUID/Negative limit with offset | Validates limits and offsets |
| internal/entity/query/files_test.go | TestFilesByUID/offset and limit | Validates limits and offsets |
| internal/entity/dbtest/dbtest_init_test.go | TestInit/* | checks that the number of records in a fresh database is correct |
| internal/entity/dbtest/dbtest_foreignkey_test.go | TestDbtestForeignKey_Validate/Photos_CameraID | makes sure that the database throws a foreign key error |
| internal/entity/dbtest/dbtest_foreignkey_test.go | TestDbtestForeignKey_Validate/Photos_LensID | makes sure that the database throws a foreign key error |
| internal/entity/dbtest/dbtest_fieldlength_test.go | TestInitDBLengths/PhotoMaxVarLengths | makes sure that the database can hold specified maximum length of each column in a Photo |
| internal/entity/dbtest/dbtest_fieldlength_test.go | TestInitDBLengths/PhotoExceedMax* | makes sure that the database throws an error when the maximum length is exceeded by 1 character in a Photo |
| internal/entity/dbtest/dbtest_blocking_test.go | TestEntity_UpdateDBErrors | verifies that entity.Update detects and returns database level errors |
| internal/entity/dbtest/dbtest_blocking_test.go | TestEntity_SaveDBErrors | verifies that entity.Save detects and returns database level errors |
| internal/entity/dbtest/dbtest_validatecreatesave_test.go | * | A set of tests that compare Gorm version functionality.  They can show you what to expect for a variety of scenarios.  The Entity.Save test no longer fails as Entity.Save has been uplifted to work like Gorm V1. |
| internal/entity/camera_test.go | TestCamera_ScopedSearchFirst/* | Validate that ScopedSearchFirstCamera returns the expected results/errors |
| internal/entity/entitiy_save_test.go | TestSave/NewParentPhotoWithNewChildDetails | Validate that FK violations do not happen when saving Details with a new Photo |
| internal/entity/entity_update_test.go | TestEntitiy_Update/InconsistentCameraVSCameraID | Validates that an inconsistent CameraID and Camera.ID is handled |
| internal/entity/entity_values_test.go | TestModelValuesStructOption/NoInterface | Validates that ModelValuesStructOption handles lack of an Interface |
| internal/entity/entity_values_test.go | TestModelValuesStructOption/NewPhoto | Validates that ModelValuesStructOption handles a new and empty Struct |
| internal/entity/entity_values_test.go | TestModelValuesStructOption/ExistingPhoto | Validates that ModelValuesStructOption handles a populated Struct and does/doesn't remove appropriate attributes from the Struct |
| internal/entity/entity_values_test.go | TestModelValuesStructOption/NewFace | Validates that ModelValuesStructOption handles a new and empty Struct |
| internal/entity/entity_values_test.go | TestModelValuesStructOption/ExistingFace | Validates that ModelValuesStructOption handles a populated Struct and does/doesn't remove appropriate attributes from the Struct |
| internal/entity/entity_values_test.go | TestModelValuesStructOption/AllTypes | Validates that ModelValuesStructOption handles a populated Struct and does/doesn't remove appropriate attributes from the Struct.  **This test actions all the known types in PhotoPrism.** |
| internal/entity/file_test.go | TestFile_MissingPhotoID/No PhotoID or Photo | Validate that an error is raised when attempting to create a File without a Photo |
| internal/entity/file_test.go | TestFile_MissingPhotoID/No PhotoID and Photo.ID = 0 | Validate that an error is raised when attempting to create a File with a Photo that hasn't been created |
| internal/entity/file_test.go | TestFile_MissingPhotoID/PhotoID = 0 and Photo.ID = 0 | Validate that an error is raised when attempting to create a File without a PhotoID and a Photo that hasn't been created|
| internal/entity/lens_test.go | TestLens_ScopedSearchFirst/* | Validate that ScopedSearchFirstLens returns the expected results/errors |
| internal/entity/photo_label_test.go | TestFirstOrCreatePhotoLabel/success path 1 | Validate that an existing Label is added |
| internal/entity/photo_label_test.go | TestFirstOrCreatePhotoLabel/success path 2 | Validate that a new Label is added |
| internal/entity/photo_quality_test.go | TestPhoto_QualityScore/digikam test | Test scenario where a new Photo is created with and saved correctly, then a 2nd file is added and saved.  Ensure that the QualityScore is not incorrect.  This replicates a front end acceptance test that was failing for GormV2 |
| internal/entity/photo_test.go | TestSavePhotoForm/BadCamera | Validate that when a bad CameraID is passed from a form it is replaced with UnknownCameraID |
| internal/entity/photo_test.go | TestSavePhotoForm/BadLens | Validate that when a bad LensID is passed from a form it is replaced with UnknownLensID |
| internal/entity/photo_test.go | TestPhoto_Save/BadCameraID | Validate that when a mismatch between CameraID and Camera.ID is saved, the CameraID wins and an Error is added to database |
| internal/entity/photo_test.go | TestPhoto_Save/BadCellID | Validate that when a mismatch between CellID and Cell.ID is saved, the CellID wins and an Error is added to database |
| internal/entity/photo_test.go | TestPhoto_Save/BadLensID | Validate that when a mismatch between LensID and Lens.ID is saved, the LensID wins and an Error is added to database |
| internal/entity/photo_test.go | TestPhoto_Save/BadPlaceID | Validate that when a mismatch between PlaceID and Place.ID is saved, the PlaceID wins and an Error is added to database |
| internal/entity/photo_test.go | TestPhoto_UnscopedSearch/* | Validate that UnscopedSearchPhoto returns that expected results/errors |
| internal/entity/photo_test.go | TestPhoto_ScopedSearch/* | Validate that ScopedSearchPhoto returns that expected results/errors |
| internal/entity/photos_test.go | TestPhotos_UnscopedSearch/* | Validate that UnscopedSearchPhotos returns that expected results/errors |
| internal/entity/photos_test.go | TestPhotos_ScopedSearch/* | Validate that ScopedSearchPhotos returns that expected results/errors |
| internal/photoprism/index_mediafile_test.go | TestIndex_MediaFile/twoFiles | Test scenario where 2 files are indexed (Primary and Json) that it is done correctly.  This replicates a front end acceptance test that was failing for GormV2 |



**Please note that the tests in internal/entity/dbtest all MUST use the dbtestMutex as they must run synchronous due to the database blocking tests.  Failure to include the dbtestMutex will cause unexpected failure of the test.**  


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
| SKIP | github.com/photoprism/photoprism/internal/testextras	[no test files] |
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
| SKIP | github.com/photoprism/photoprism/internal/testextras	[no test files] |
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


The following is current state of acceptance testing against sqlite:  
```
> photoprism@1 testcafe
> testcafe chrome --headless=new --test-grep ^(Common|Core)\:* --test-meta mode=auth --config-file ./testcaferc.json tests/acceptance
```
 Running tests in:  
 - Chrome 128.0.0.0 / Ubuntu 24.04  

 Test authentication  
 ✓ Common: Login and Logout  
 ✓ Common: Login with wrong credentials  
 ✓ Common: Change password  
 ✓ Common: Delete Clipboard on logout  
  
 Test components  
 ✓ Common: Mobile Toolbar  
  
 Test account settings  
 ✓ Core: Sign in with recovery code  
 ✓ Core: Create App Password  
 ✓ Core: Check App Password has limited permissions and last updated is set  
 ✓ Core: Try to login with invalid credentials/insufficient scope  
 ✓ Core: Delete App Password  
 ✓ Common: Try to activate 2FA with wrong password/passcode  
  
 Test general settings  
 ✓ Common: Disable delete  
 ✓ Common: Change language  
 ✓ Common: Disable pages: import, originals, logs, moments, places, library  
 ✓ Common: Disable people and labels  
 ✓ Common: Disable private, archive and quality filter  
 ✓ Common: Disable upload, download, edit and share  
  
 Test link sharing  
 ✓ Common: Create, view, delete shared albums  
 ✓ Common: Verify visitor role has limited permissions  
  
  
 19 passed (29m 11s)  

 Warnings (2):  
- TestCafe cannot interact with the <div title="Logout" class="v-list__tile__action">...</div> element because another element obstructs it.
- TestCafe cannot interact with the <button type="button" class="action-clear v-btn v-btn--floating v-btn--small theme--dark accent">...</button> element because another element obstructs it.
  

Running public-mode tests in Chrome...
```
(cd frontend &&	npm run testcafe -- "chrome --headless=new" --test-grep "^(Common|Core)\:*" --test-meta mode=public --config-file ./testcaferc.json "tests/acceptance")

> photoprism@1 testcafe
> testcafe chrome --headless=new --test-grep ^(Common|Core)\:* --test-meta mode=public --config-file ./testcaferc.json tests/acceptance
```
 Running tests in:  
 - Chrome 128.0.0.0 / Ubuntu 24.04  

 Test albums  
 ✓ Common: Create/delete album on /albums  
 ✓ Common: Create/delete album during add to album  
 ✓ Common: Update album details  
 ✓ Common: Add/Remove Photos to/from album  
 ✓ Common: Use album search and filters  
 ✓ Common: Test album autocomplete  
 ✓ Common: Create, Edit, delete sharing link  
 ✓ Common: Verify album sort options  
  
 Test calendar  
 ✓ Common: View calendar  
 ✓ Common: Update calendar details  
 ✓ Common: Create, Edit, delete sharing link for calendar  
 ✓ Common: Create/delete album-clone from calendar  
 ✓ Common: Verify calendar sort options  
  
 Test components  
 ✓ Common: Test filter options  
 ✓ Common: Fullscreen mode  
 ✓ Common: Mosaic view  
 ✓ Common: List view  
 ✓ Common: Card view  
 ✓ Common: Mobile Toolbar  
  
 Test folders  
 ✓ Common: View folders  
 ✓ Common: Update folder details  
 ✓ Common: Create, Edit, delete sharing link  
 ✓ Common: Create/delete album-clone from folder  
 ✓ Common: Verify folder sort options  
  
 Test labels  
 ✓ Common: Remove/Activate Add/Delete Label from photo  
 ✓ Common: Toggle between important and all labels  
 ✓ Common: Rename Label  
 ✓ Common: Add label to album  
 ✓ Common: Delete label  
  
 Import file from folder  
 ✓ Common: Import files from folder using copy  
  
 Test index  
 ✓ Common: Index files from folder  
  
 Test moments  
 ✓ Common: Update moment details  
 ✓ Common: Create, Edit, delete sharing link for moment  
 ✓ Common: Create/delete album-clone from moment  
 ✓ Common: Verify moment sort options  
  
 Test files  
 ✓ Common: Navigate in originals  
 ✓ Common: Add original files to album  
 ✓ Common: Download available in originals  
  
 Test people  
 ✓ Common: Add name to new face and rename subject  
 ✓ Common: Add + Reject name on people tab  
 ✓ Common: Test mark subject as favorite  
 ✓ Common: Test new face autocomplete  
 ✓ Common: Remove face  
 ✓ Common: Hide face  
 ✓ Common: Hide person  
  
 Test photos archive and private functionalities  
 ✓ Common: Private/unprivate photo/video using clipboard and list  
 ✓ Common: Archive/restore video, photos, private photos and review photos using clipboard  
 ✓ Common: Check that archived files are not shown in monochrome/panoramas/stacks/scans/review/albums/favorites/private/videos/calendar/moments/states/labels/folders/originals  
 ✓ Common: Check that private files are not shown in monochrome/panoramas/stacks/scans/review/albums/favorites/archive/videos/calendar/moments/states/labels/folders/originals  
 ✓ Common: Check delete all dialog  
  
 Does not work in container and we have no content-disposition header anymore  
 ø Common: Test download jpg file from context menu and fullscreen  
 ø Common: Test download video from context menu  
 ø Common: Test download multiple jpg files from context menu  
 ø Common: Test raw file from context menu and fullscreen mode  
  
 Test photos upload and delete  
 ✓ Core: Upload + Delete jpg/json  
 ✓ Core: Upload + Delete video  
 ✓ Core: Upload to existing Album + Delete  
 ✓ Core: Upload jpg to new Album + Delete  
 ✓ Core: Try uploading nsfw file  
 ✓ Core: Try uploading txt file  
  
 Test photos  
 ✖ Common: Scroll to top (also fails on GormV1)  
 ✓ Common: Download single photo/video using clipboard and fullscreen mode  
 ✓ Common: Approve photo using approve and by adding location  
 ✓ Common: Like/dislike photo/video  
 ✓ Common: Edit photo/video  
 ø Common: Navigate from card view to place  
 ✓ Common: Mark photos/videos as panorama/scan  
 ✓ Common: Navigate from card view to photos taken at the same date  
  
 Search and open photo from places  
 ✓ Common: Test places  
  
 Test about  
 ✓ Core: About page is displayed with all links  
 ✓ Core: License page is displayed with all links  
  
 Test stacks  
desktop  
 ✓ Common: View all files of a stack  
 ✓ Common: Change primary file  
 ✓ Common: Ungroup files  
 ✓ Common: Delete non primary file  
  
 Test states  
 ✓ Common: Update state details  
 ✓ Common: Create, Edit, delete sharing link for state  
 ✓ Common: Create/delete album-clone from state  
  
  
 1/73 failed (1h 03m 27s)  
 5 skipped  
