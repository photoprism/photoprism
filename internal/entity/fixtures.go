package entity

// CreateTestFixtures inserts all known entities into the database for testing.
func CreateTestFixtures() {
	if err := Admin.SetPassword("photoprism"); err != nil {
		log.Error(err)
	}

	CreateLabelFixtures()
	CreateCameraFixtures()
	CreateCountryFixtures()
	CreatePhotoFixtures()
	CreateAlbumFixtures()
	CreateServiceFixtures()
	CreateLinkFixtures()
	CreatePhotoAlbumFixtures()
	CreateFolderFixtures()
	CreateFileFixtures()
	CreateKeywordFixtures()
	CreatePhotoKeywordFixtures()
	CreateCategoryFixtures()
	CreateCellFixtures()
	CreatePlaceFixtures()
	CreateFileShareFixtures()
	CreateFileSyncFixtures()
	CreateLensFixtures()
	CreateSubjectFixtures()
	CreateMarkerFixtures()
	CreateFaceFixtures()
	CreateUserFixtures()
	CreateSessionFixtures()
	CreateClientFixtures()
	CreateReactionFixtures()
	CreatePasswordFixtures()
	CreateUserShareFixtures()
}
