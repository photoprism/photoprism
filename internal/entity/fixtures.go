package entity

// CreateTestFixtures inserts all known entities into the database for testing.
func CreateTestFixtures() {
	if err := Admin.SetPassword("photoprism"); err != nil {
		log.Error(err)
	}

	CreateLabelFixtures()
	CreateCameraFixtures()
	CreateLensFixtures()
	CreateCountryFixtures()
	CreateCellFixtures()
	CreatePlaceFixtures()
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
	CreateFileShareFixtures()
	CreateFileSyncFixtures()
	CreateSubjectFixtures()
	CreateMarkerFixtures()
	CreateFaceFixtures()
	CreateUserFixtures()
	CreateSessionFixtures()
	CreateClientFixtures()
	CreateReactionFixtures()
	CreatePasscodeFixtures()
	CreatePasswordFixtures()
	CreateUserShareFixtures()
}
