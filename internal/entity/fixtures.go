package entity

// CreateTestFixtures inserts all known entities into the database for testing.
func CreateTestFixtures() {
	CreateLabelFixtures()
	CreateCameraFixtures()
	CreateCountryFixtures()
	CreatePhotoFixtures()
	CreateAlbumFixtures()
	CreateAccountFixtures()
	CreateLinkFixtures()
	CreatePhotoAlbumFixtures()
	CreateFolderFixtures()
	CreateFileFixtures()
	CreateKeywordFixtures()
	CreatePhotoKeywordFixtures()
	CreateCategoryFixtures()
	CreateLocationFixtures()
	CreatePlaceFixtures()
	CreateFileShareFixtures()
	CreateFileSyncFixtures()
	CreateLensFixtures()
}
