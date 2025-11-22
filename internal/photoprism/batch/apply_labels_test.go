package batch

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
)

// TestApplyLabels exercises batch action logic.
func TestApplyLabels(t *testing.T) {
	t.Run("AddExistingLabelByUID", func(t *testing.T) {
		photo := entity.PhotoFixtures.Pointer("Photo06")
		labelUID := entity.LabelFixtures.Get("landscape").LabelUID

		// Ensure clean slate.
		entity.Db().Where("photo_id = ? AND label_id = (SELECT id FROM labels WHERE label_uid = ?)", photo.ID, labelUID).Delete(&entity.PhotoLabel{})
		photo.PreloadLabels()

		labels := Items{
			Items: []Item{
				{Action: ActionAdd, Value: labelUID},
			},
			Action: ActionUpdate,
		}

		if err := ApplyLabels(photo, labels); err != nil {
			t.Fatal(err)
		}

		// Verify photo has the label with 100% confidence and batch source.
		var photoLabel entity.PhotoLabel

		if err := entity.Db().Preload("Label").Where("photo_id = ? AND label_id = (SELECT id FROM labels WHERE label_uid = ?)", photo.ID, labelUID).First(&photoLabel).Error; err != nil {
			t.Fatal(err)
		}

		if photoLabel.Uncertainty != 0 {
			t.Errorf("expected uncertainty 0 (100%% confidence), got %d", photoLabel.Uncertainty)
		}

		if photoLabel.LabelSrc != entity.SrcBatch {
			t.Errorf("expected label source %s, got %s", entity.SrcBatch, photoLabel.LabelSrc)
		}
	})
	t.Run("AddNewLabelByTitle", func(t *testing.T) {
		photo := entity.PhotoFixtures.Pointer("Photo07")
		labelTitle := "Test Label for Actions"

		labels := Items{
			Items: []Item{
				{Action: ActionAdd, Title: labelTitle},
			},
			Action: ActionUpdate,
		}

		if err := ApplyLabels(photo, labels); err != nil {
			t.Fatal(err)
		}

		// Verify label was created and added to photo.
		var label entity.Label

		if err := entity.Db().Where("label_name = ?", labelTitle).First(&label).Error; err != nil {
			t.Fatal(err)
		}

		if label.LabelName != labelTitle {
			t.Errorf("expected label name %s, got %s", labelTitle, label.LabelName)
		}

		var photoLabel entity.PhotoLabel

		if err := entity.Db().Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).First(&photoLabel).Error; err != nil {
			t.Fatal(err)
		}

		if photoLabel.Uncertainty != 0 {
			t.Errorf("expected uncertainty 0, got %d", photoLabel.Uncertainty)
		}

		if photoLabel.LabelSrc != entity.SrcBatch {
			t.Errorf("expected label source %s, got %s", entity.SrcBatch, photoLabel.LabelSrc)
		}
	})
	t.Run("RemoveLabelByUID", func(t *testing.T) {
		photo := entity.PhotoFixtures.Pointer("Photo08")
		label := entity.LabelFixtures.Get("flower")

		// First add the label manually.
		photoLabel := entity.NewPhotoLabel(photo.ID, label.ID, 0, entity.SrcBatch)

		if err := entity.Db().Create(&photoLabel).Error; err != nil {
			t.Fatal(err)
		}

		// Verify label is on photo.
		var checkPhotoLabel entity.PhotoLabel

		if err := entity.Db().Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).First(&checkPhotoLabel).Error; err != nil {
			t.Fatal(err)
		}

		// Now remove it
		labels := Items{
			Items: []Item{
				{Action: ActionRemove, Value: label.LabelUID},
			},
			Action: ActionUpdate,
		}

		if err := ApplyLabels(photo, labels); err != nil {
			t.Fatal(err)
		}

		// Verify label was removed (should be deleted from photos_labels).
		var deletedLabel entity.PhotoLabel

		if err := entity.Db().Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).First(&deletedLabel).Error; err == nil {
			t.Error("expected label to be deleted, but it was found")
		}
	})
	t.Run("RemoveLabelWithPreloadedPhoto", func(t *testing.T) {
		photo := entity.PhotoFixtures.Pointer("Photo17")
		label := entity.LabelFixtures.Get("landscape")

		entity.Db().Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).Delete(&entity.PhotoLabel{})
		photo.PreloadLabels()

		if err := entity.Db().Create(entity.NewPhotoLabel(photo.ID, label.ID, 0, entity.SrcManual)).Error; err != nil {
			t.Fatal(err)
		}

		preloaded, err := query.PhotoPreloadByUID(photo.PhotoUID)

		if err != nil {
			t.Fatal(err)
		}

		if len(preloaded.Labels) == 0 {
			t.Fatalf("expected preloaded labels for %s", photo.PhotoUID)
		}

		labels := Items{
			Items:  []Item{{Action: ActionRemove, Value: label.LabelUID}},
			Action: ActionUpdate,
		}

		if errs := ApplyLabels(&preloaded, labels); errs != nil {
			t.Fatal(errs)
		}

		var deleted entity.PhotoLabel

		if err = entity.Db().Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).First(&deleted).Error; err == nil {
			t.Fatal("expected label relation to be removed")
		}
	})
	t.Run("RemoveAutoLabelMarksBlocked", func(t *testing.T) {
		photo := entity.PhotoFixtures.Pointer("Photo09")
		label := entity.LabelFixtures.Get("cake")

		// Add label with auto source (not manual/batch).
		photoLabel := entity.NewPhotoLabel(photo.ID, label.ID, 15, entity.SrcImage)

		if err := entity.Db().Create(&photoLabel).Error; err != nil {
			t.Fatal(err)
		}

		labels := Items{
			Items: []Item{
				{Action: ActionRemove, Value: label.LabelUID},
			},
			Action: ActionUpdate,
		}

		if err := ApplyLabels(photo, labels); err != nil {
			t.Fatal(err)
		}

		var blocked entity.PhotoLabel

		if err := entity.Db().Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).First(&blocked).Error; err != nil {
			t.Fatalf("expected auto label relation to remain, lookup failed: %v", err)
		}

		if blocked.Uncertainty != 100 {
			t.Errorf("expected uncertainty 100 (blocked), got %d", blocked.Uncertainty)
		}

		if blocked.LabelSrc != entity.SrcBatch {
			t.Errorf("expected label source %s, got %s", entity.SrcBatch, blocked.LabelSrc)
		}
	})

	t.Run("RemoveVisionLabelBlocksRelation", func(t *testing.T) {
		photo := entity.PhotoFixtures.Pointer("Photo15")
		label := entity.LabelFixtures.Get("landscape")

		entity.Db().Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).Delete(&entity.PhotoLabel{})
		photo.PreloadLabels()

		if err := entity.Db().Create(entity.NewPhotoLabel(photo.ID, label.ID, 25, entity.SrcVision)).Error; err != nil {
			t.Fatal(err)
		}

		labels := Items{Items: []Item{{Action: ActionRemove, Value: label.LabelUID}}, Action: ActionUpdate}

		preloaded, err := query.PhotoPreloadByUID(photo.PhotoUID)

		if err != nil {
			t.Fatal(err)
		}

		if errs := ApplyLabels(&preloaded, labels); err != nil {
			t.Fatal(errs)
		}

		var updated entity.PhotoLabel

		if err = entity.Db().Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).First(&updated).Error; err != nil {
			t.Fatalf("expected vision label relation to remain, lookup failed: %v", err)
		}

		if updated.Uncertainty != 100 {
			t.Errorf("expected uncertainty 100 (blocked), got %d", updated.Uncertainty)
		}

		if updated.LabelSrc != entity.SrcBatch {
			t.Errorf("expected label source %s, got %s", entity.SrcBatch, updated.LabelSrc)
		}
	})

	t.Run("KeepHigherPriorityLabel", func(t *testing.T) {
		photo := entity.PhotoFixtures.Pointer("Photo16")
		label := entity.LabelFixtures.Get("flower")

		entity.Db().Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).Delete(&entity.PhotoLabel{})
		photo.PreloadLabels()

		if err := entity.Db().Create(entity.NewPhotoLabel(photo.ID, label.ID, 10, entity.SrcAdmin)).Error; err != nil {
			t.Fatal(err)
		}

		labels := Items{Items: []Item{{Action: ActionRemove, Value: label.LabelUID}}, Action: ActionUpdate}

		if errs := ApplyLabels(photo, labels); errs != nil {
			t.Fatal(errs)
		}

		var updated entity.PhotoLabel

		if err := entity.Db().Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).First(&updated).Error; err != nil {
			t.Fatal(err)
		}

		if updated.Uncertainty != 10 {
			t.Errorf("expected uncertainty to remain 10, got %d", updated.Uncertainty)
		}

		if updated.LabelSrc != entity.SrcAdmin {
			t.Errorf("expected label source %s, got %s", entity.SrcAdmin, updated.LabelSrc)
		}
	})
	t.Run("KeepManualLabelWithZeroProbability", func(t *testing.T) {
		photo := entity.PhotoFixtures.Pointer("Photo12")
		label := entity.LabelFixtures.Get("landscape")

		entity.Db().Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).Delete(&entity.PhotoLabel{})
		photo.PreloadLabels()

		pl := entity.NewPhotoLabel(photo.ID, label.ID, 100, entity.SrcManual)

		if err := entity.Db().Create(pl).Error; err != nil {
			t.Fatal(err)
		}

		var before entity.PhotoLabel
		if err := entity.Db().Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).First(&before).Error; err != nil {
			t.Fatal(err)
		}

		labels := Items{Items: []Item{{Action: ActionRemove, Value: label.LabelUID}}, Action: ActionUpdate}

		preloaded, err := query.PhotoPreloadByUID(photo.PhotoUID)
		if err != nil {
			t.Fatal(err)
		}
		if err := ApplyLabels(&preloaded, labels); err != nil {
			t.Fatal(err)
		}

		var persisted entity.PhotoLabel
		if err := entity.Db().Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).First(&persisted).Error; err != nil {
			t.Fatal(err)
		}

		if persisted.Uncertainty != 100 {
			t.Errorf("expected uncertainty to remain 100, got %d", persisted.Uncertainty)
		}

		if persisted.LabelSrc != entity.SrcManual {
			t.Errorf("expected label source %s, got %s", entity.SrcManual, persisted.LabelSrc)
		}
	})
	t.Run("UpdateExistingLabelConfidence", func(t *testing.T) {
		photo := entity.PhotoFixtures.Pointer("Photo10")
		label := entity.LabelFixtures.Get("landscape")

		// First, delete any existing photo-label to ensure clean start.
		if existing, err := entity.FindPhotoLabel(photo.ID, label.ID, true); err == nil && existing != nil {
			assert.NoError(t, existing.Delete())
		}

		// Load labels from database.
		photo.PreloadLabels()

		// Add label with some uncertainty using FirstOrCreatePhotoLabel.
		photoLabel := entity.FirstOrCreatePhotoLabel(entity.NewPhotoLabel(photo.ID, label.ID, 50, entity.SrcImage))

		if photoLabel == nil {
			t.Fatal("failed to create photo label")
		}

		// Finally, delete the added photo label to ensure a clean state.
		t.Cleanup(func() {
			_ = photoLabel.Delete()
		})

		// Verify initial state
		if photoLabel.Uncertainty != 50 {
			t.Errorf("expected uncertainty 50, got %d", photoLabel.Uncertainty)
		}

		if photoLabel.LabelSrc != entity.SrcImage {
			t.Errorf("expected label source %s, got %s", entity.SrcImage, photoLabel.LabelSrc)
		}

		// Re-add same label via batch (should update to 100% confidence).
		labels := Items{
			Items: []Item{
				{Action: ActionAdd, Value: label.LabelUID},
			},
			Action: ActionUpdate,
		}

		if err := ApplyLabels(photo, labels); err != nil {
			t.Fatal(err)
		}

		// Verify label confidence was updated.
		var updatedLabel entity.PhotoLabel

		if err := entity.Db().Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).First(&updatedLabel).Error; err != nil {
			t.Fatal(err)
		}

		if updatedLabel.Uncertainty != 0 {
			t.Errorf("expected uncertainty 0 (100%% confidence), got %d", updatedLabel.Uncertainty)
		}

		if updatedLabel.LabelSrc != entity.SrcBatch {
			t.Errorf("expected label source %s, got %s", entity.SrcBatch, updatedLabel.LabelSrc)
		}
	})
	t.Run("AddExistingBatchLabelZeroConfidence", func(t *testing.T) {
		// Load labels from database.
		photo := entity.PhotoFixtures.Pointer("Photo10").PreloadLabels()
		label := entity.LabelFixtures.Get("landscape")

		// Add label with some uncertainty using FirstOrCreatePhotoLabel.
		photoLabel := entity.FirstOrCreatePhotoLabel(entity.NewPhotoLabel(photo.ID, label.ID, 100, entity.SrcBatch))

		if photoLabel == nil {
			t.Fatal("failed to create photo label")
		}

		// Finally, delete the added photo label to ensure a clean state.
		t.Cleanup(func() {
			_ = photoLabel.Delete()
		})

		// Verify initial state
		if photoLabel.Uncertainty != 100 {
			t.Errorf("expected uncertainty 100, got %d", photoLabel.Uncertainty)
		}

		if photoLabel.LabelSrc != entity.SrcBatch {
			t.Errorf("expected label source %s, got %s", entity.SrcBatch, photoLabel.LabelSrc)
		}

		// Re-add same label via batch (should update to 100% confidence).
		labels := Items{
			Items: []Item{
				{Action: ActionAdd, Value: label.LabelUID},
			},
			Action: ActionUpdate,
		}

		if err := ApplyLabels(photo, labels); err != nil {
			t.Fatal(err)
		}

		// Verify label confidence was updated.
		var updatedLabel entity.PhotoLabel

		if err := entity.Db().Where("photo_id = ? AND label_id = ?", photo.ID, label.ID).First(&updatedLabel).Error; err != nil {
			t.Fatal(err)
		}

		if updatedLabel.Uncertainty != 0 {
			t.Errorf("expected uncertainty 0 (100%% confidence), got %d", updatedLabel.Uncertainty)
		}

		if updatedLabel.LabelSrc != entity.SrcBatch {
			t.Errorf("expected label source %s, got %s", entity.SrcBatch, updatedLabel.LabelSrc)
		}
	})
	t.Run("InvalidPhotoReturnsError", func(t *testing.T) {
		labels := Items{
			Items: []Item{
				{Action: ActionAdd, Value: "some-uid"},
			},
			Action: ActionUpdate,
		}

		err := ApplyLabels(nil, labels)

		if err == nil {
			t.Error("expected error for nil photo")
		}

		emptyPhoto := &entity.Photo{}
		err = ApplyLabels(emptyPhoto, labels)

		if err == nil {
			t.Error("expected error for empty photo")
		}
	})
	// Additional error cases
	t.Run("AddNonExistingLabelByUID", func(t *testing.T) {
		photo := entity.PhotoFixtures.Pointer("Photo11")
		nonExistingLabelUID := "lt9lxuqxpoaaaaaa" // Invalid/non-existing UID

		labels := Items{
			Items: []Item{
				{Action: ActionAdd, Value: nonExistingLabelUID},
			},
			Action: ActionUpdate,
		}

		err := ApplyLabels(photo, labels)

		if err == nil {
			t.Error("expected error when adding non-existing label by UID, but got none")
		}
	})
	t.Run("AddLabelWithInvalidUID", func(t *testing.T) {
		photo := entity.PhotoFixtures.Pointer("Photo12")
		invalidUID := "invalid-label-uid-format" // Invalid UID format

		labels := Items{
			Items: []Item{
				{Action: ActionAdd, Value: invalidUID},
			},
			Action: ActionUpdate,
		}

		err := ApplyLabels(photo, labels)

		if err == nil {
			t.Error("expected error when adding label with invalid UID, but got none")
		}
	})
	t.Run("RemoveNonExistingLabelByUID", func(t *testing.T) {
		photo := entity.PhotoFixtures.Pointer("Photo13")
		nonExistingLabelUID := "lt9lxuqxpobbbbbb" // Non-existing UID

		labels := Items{
			Items: []Item{
				{Action: ActionRemove, Value: nonExistingLabelUID},
			},
			Action: ActionUpdate,
		}

		err := ApplyLabels(photo, labels)
		if err == nil {
			t.Error("expected error when removing non-existing label, but got none")
		}
	})
	t.Run("InvalidActionOnLabel", func(t *testing.T) {
		photo := entity.PhotoFixtures.Pointer("Photo14")
		labelUID := entity.LabelFixtures.Get("landscape").LabelUID

		labels := Items{
			Items: []Item{
				{Action: "invalid-action", Value: labelUID}, // Invalid action
			},
			Action: ActionUpdate,
		}

		err := ApplyLabels(photo, labels)
		if err == nil {
			t.Error("expected error for invalid action, but got none")
		}
	})
	t.Run("EmptyLabelItems", func(t *testing.T) {
		photo := entity.PhotoFixtures.Pointer("Photo15")

		labels := Items{
			Items: []Item{}, // Empty items
		}

		// This should not error, but should be a no-op
		err := ApplyLabels(photo, labels)

		if err != nil {
			t.Errorf("expected no error for empty label items, but got: %v", err)
		}
	})
	t.Run("AddLabelWithEmptyValueAndTitle", func(t *testing.T) {
		photo := entity.PhotoFixtures.Pointer("Photo16")

		labels := Items{
			Items: []Item{
				{Action: ActionAdd, Value: "", Title: ""}, // Both empty
			},
			Action: ActionUpdate,
		}

		err := ApplyLabels(photo, labels)

		if err == nil {
			t.Error("expected error when both Value and Title are empty, but got none")
		}
	})
	t.Run("RemoveLabelNotAssignedToPhoto", func(t *testing.T) {
		photo := entity.PhotoFixtures.Pointer("Photo17")
		labelUID := entity.LabelFixtures.Get("bird").LabelUID

		// Ensure the label is not assigned to this photo
		entity.Db().Where("photo_id = ? AND label_id = (SELECT id FROM labels WHERE label_uid = ?)", photo.ID, labelUID).Delete(&entity.PhotoLabel{})
		photo.PreloadLabels()

		labels := Items{
			Items: []Item{
				{Action: ActionRemove, Value: labelUID},
			},
			Action: ActionUpdate,
		}

		err := ApplyLabels(photo, labels)

		if err == nil {
			t.Error("expected error when removing label not assigned to photo, but got none")
		}
	})
}

// TestIndexPhotoLabels exercises batch action logic.
func TestIndexPhotoLabels(t *testing.T) {
	labels := entity.PhotoLabels{
		{LabelID: 11},
		{LabelID: 0},
		{LabelID: 22},
	}

	idx := indexPhotoLabels(labels)

	if len(idx) != 2 {
		t.Fatalf("expected 2 labels in index, got %d", len(idx))
	}

	if idx[11] == nil || idx[22] == nil {
		t.Fatal("expected indexed labels to be present")
	}
}

// TestDetermineLabelRemovalAction enumerates the matrix documented in
// internal/photoprism/batch/README.md under "Rules for Deleting Photo Labels"
// so we can guarantee the implementation stays aligned with the published
// expectations.
func TestDetermineLabelRemovalAction(t *testing.T) {
	t.Parallel()

	makeLabel := func(src string, uncertainty int) *entity.PhotoLabel {
		return &entity.PhotoLabel{
			LabelSrc:    src,
			Uncertainty: uncertainty,
		}
	}

	tests := []struct {
		name string
		pl   *entity.PhotoLabel
		want labelRemovalAction
	}{
		{name: "NilLabelDefaultsToKeep", pl: nil, want: labelRemovalKeep},

		// image, openai, ollama (priority < batch) => block regardless of confidence.
		{name: "ImagePriority8Uncertainty0", pl: makeLabel(entity.SrcImage, 0), want: labelRemovalBlock},
		{name: "OpenAI_Uncertainty50", pl: makeLabel(entity.SrcOpenAI, 50), want: labelRemovalBlock},
		{name: "Ollama_Uncertainty100", pl: makeLabel(entity.SrcOllama, 100), want: labelRemovalBlock},

		// Generic sources with priority < 64 => block.
		{name: "FilePriority2_Uncertainty0", pl: makeLabel(entity.SrcFile, 0), want: labelRemovalBlock},
		{name: "AutoPriority1_Uncertainty35", pl: makeLabel(entity.SrcAuto, 35), want: labelRemovalBlock},
		{name: "UnknownSrcDefaultsToPriority0", pl: makeLabel("", 100), want: labelRemovalBlock},

		// manual source (priority == batch) => delete unless already blocked (uncertainty 100).
		{name: "Manual_Uncertainty0", pl: makeLabel(entity.SrcManual, 0), want: labelRemovalDelete},
		{name: "Manual_Uncertainty42", pl: makeLabel(entity.SrcManual, 42), want: labelRemovalDelete},
		{name: "Manual_Uncertainty100", pl: makeLabel(entity.SrcManual, 100), want: labelRemovalKeep},

		// vision source (priority == batch) => block unless already blocked.
		{name: "Vision_Uncertainty0", pl: makeLabel(entity.SrcVision, 0), want: labelRemovalBlock},
		{name: "Vision_Uncertainty60", pl: makeLabel(entity.SrcVision, 60), want: labelRemovalBlock},
		{name: "Vision_Uncertainty100", pl: makeLabel(entity.SrcVision, 100), want: labelRemovalKeep},

		// batch source mirrors manual (delete until explicitly blocked).
		{name: "Batch_Uncertainty0", pl: makeLabel(entity.SrcBatch, 0), want: labelRemovalDelete},
		{name: "Batch_Uncertainty15", pl: makeLabel(entity.SrcBatch, 15), want: labelRemovalDelete},
		{name: "Batch_Uncertainty100", pl: makeLabel(entity.SrcBatch, 100), want: labelRemovalKeep},

		// admin (higher priority) => always keep.
		{name: "Admin_Uncertainty0", pl: makeLabel(entity.SrcAdmin, 0), want: labelRemovalKeep},
		{name: "Admin_Uncertainty55", pl: makeLabel(entity.SrcAdmin, 55), want: labelRemovalKeep},
		{name: "Admin_Uncertainty100", pl: makeLabel(entity.SrcAdmin, 100), want: labelRemovalKeep},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := determineLabelRemovalAction(tc.pl)
			if got != tc.want {
				src := "<nil>"
				uncertainty := 0
				if tc.pl != nil {
					src = tc.pl.LabelSrc
					uncertainty = tc.pl.Uncertainty
				}

				t.Fatalf("determineLabelRemovalAction(src=%q uncertainty=%d) = %s, want %s",
					src, uncertainty, got.String(), tc.want.String())
			}
		})
	}
}

func TestLabelRemovalActionString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		act  labelRemovalAction
		want string
	}{
		{name: "Keep", act: labelRemovalKeep, want: "keep"},
		{name: "Block", act: labelRemovalBlock, want: "block"},
		{name: "Delete", act: labelRemovalDelete, want: "delete"},
		{name: "Unknown", act: labelRemovalAction(99), want: "unknown"},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := tc.act.String(); got != tc.want {
				t.Fatalf("labelRemovalAction(%d).String() = %q, want %q", tc.act, got, tc.want)
			}
		})
	}
}
