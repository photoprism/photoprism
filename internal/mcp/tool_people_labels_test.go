package mcp

import (
	"context"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
)

func TestValidateListInput(t *testing.T) {
	limit, offset, err := validateListInput(0, 3, "x")
	require.NoError(t, err)
	assert.Equal(t, 20, limit)
	assert.Equal(t, 3, offset)

	_, _, err = validateListInput(-1, 0, "")
	assert.Error(t, err)
	_, _, err = validateListInput(0, -1, "")
	assert.Error(t, err)
	_, _, err = validateListInput(0, 0, string(make([]byte, 201)))
	assert.Error(t, err)
}

func TestListPeople(t *testing.T) {
	admin := entity.SessionFixtures.Pointer("alice")

	t.Run("SameAsSearch", func(t *testing.T) {
		expected, err := search.UserSubjects(form.SearchSubjects{Type: entity.SubjPerson, Files: 1, Count: search.MaxResults}, admin)
		require.NoError(t, err)
		require.NotEmpty(t, expected)

		_, out, err := listPeople(context.Background(), sessionRequest(admin), ListPeopleInput{Limit: 50})
		require.NoError(t, err)
		assert.Equal(t, len(expected), out.Total)
		require.Len(t, out.Items, len(expected))

		for i, p := range out.Items {
			assert.Equal(t, expected[i].SubjUID, p.UID)
			assert.Equal(t, expected[i].SubjName, p.Name)
			assert.Equal(t, expected[i].PhotoCount, p.MediaCount)
		}
	})
	t.Run("Paging", func(t *testing.T) {
		_, all, err := listPeople(context.Background(), sessionRequest(admin), ListPeopleInput{Limit: 50})
		require.NoError(t, err)
		require.Greater(t, all.Total, 2)

		_, page, err := listPeople(context.Background(), sessionRequest(admin), ListPeopleInput{Limit: 1, Offset: 1})
		require.NoError(t, err)
		assert.Equal(t, all.Items[1:2], page.Items)
		assert.True(t, page.Truncated)
	})
	t.Run("Hidden", func(t *testing.T) {
		// Hidden people are left out, like on the people page of the web app.
		var uid string
		require.NoError(t, entity.Db().Table("subjects").Select("subj_uid").
			Where("subj_type = ? AND subj_hidden = ? AND subj_private = ? AND file_count > 0 AND deleted_at IS NULL", entity.SubjPerson, false, false).
			Order("subj_uid").Limit(1).Row().Scan(&uid))

		_, before, err := listPeople(context.Background(), sessionRequest(admin), ListPeopleInput{Limit: 50})
		require.NoError(t, err)
		require.Contains(t, personUIDs(before.Items), uid, "listed before it is hidden")
		require.NoError(t, entity.Db().Model(&entity.Subject{}).Where("subj_uid = ?", uid).UpdateColumn("subj_hidden", true).Error)
		t.Cleanup(func() {
			_ = entity.Db().Model(&entity.Subject{}).Where("subj_uid = ?", uid).UpdateColumn("subj_hidden", false).Error
		})

		_, out, err := listPeople(context.Background(), sessionRequest(admin), ListPeopleInput{Limit: 50})
		require.NoError(t, err)
		assert.NotContains(t, personUIDs(out.Items), uid, "hidden people are left out, as on the people page")

		client := &entity.Session{AuthProvider: authn.ProviderClient.String(), AuthScope: "*"}
		_, out, err = listPeople(context.Background(), sessionRequest(client), ListPeopleInput{Limit: 50})
		require.NoError(t, err)
		assert.NotContains(t, personUIDs(out.Items), uid)
	})
	t.Run("Query", func(t *testing.T) {
		_, all, err := listPeople(context.Background(), sessionRequest(admin), ListPeopleInput{Limit: 50})
		require.NoError(t, err)
		require.Greater(t, len(all.Items), 1)

		_, out, err := listPeople(context.Background(), sessionRequest(admin), ListPeopleInput{Query: all.Items[0].Name})
		require.NoError(t, err)
		assert.Contains(t, personUIDs(out.Items), all.Items[0].UID)
		assert.Less(t, out.Total, all.Total)
	})
	t.Run("QueryIsText", func(t *testing.T) {
		var uid string
		require.NoError(t, entity.Db().Table("subjects").Select("subj_uid").
			Where("subj_type = ? AND subj_hidden = ? AND subj_private = ? AND file_count > 0 AND deleted_at IS NULL", entity.SubjPerson, false, false).
			Order("subj_uid").Limit(1).Row().Scan(&uid))

		_, before, err := listPeople(context.Background(), sessionRequest(admin), ListPeopleInput{Limit: 50})
		require.NoError(t, err)
		require.Contains(t, personUIDs(before.Items), uid, "listed before it is hidden")
		require.NoError(t, entity.Db().Model(&entity.Subject{}).Where("subj_uid = ?", uid).UpdateColumn("subj_hidden", true).Error)
		t.Cleanup(func() {
			_ = entity.Db().Model(&entity.Subject{}).Where("subj_uid = ?", uid).UpdateColumn("subj_hidden", false).Error
		})

		_, out, err := listPeople(context.Background(), sessionRequest(admin), ListPeopleInput{Query: "hidden:yes all:true"})
		require.NoError(t, err)
		assert.NotContains(t, personUIDs(out.Items), uid)
	})
	t.Run("WithoutFiles", func(t *testing.T) {
		// People who appear in no file that counts, e.g. only in private pictures, are left out like on the people page.
		_, all, err := listPeople(context.Background(), sessionRequest(admin), ListPeopleInput{Limit: 50})
		require.NoError(t, err)
		require.NotEmpty(t, all.Items)

		var count int
		uid := all.Items[0].UID
		require.NoError(t, entity.Db().Table("subjects").Select("file_count").Where("subj_uid = ?", uid).Row().Scan(&count))
		require.NoError(t, entity.Db().Model(&entity.Subject{}).Where("subj_uid = ?", uid).UpdateColumn("file_count", 0).Error)
		t.Cleanup(func() {
			_ = entity.Db().Model(&entity.Subject{}).Where("subj_uid = ?", uid).UpdateColumn("file_count", count).Error
		})

		_, out, err := listPeople(context.Background(), sessionRequest(admin), ListPeopleInput{Limit: 50})
		require.NoError(t, err)
		assert.NotContains(t, personUIDs(out.Items), uid)
	})
	t.Run("OtherTypes", func(t *testing.T) {
		other := entity.NewSubject("MCP Pet Test", "pet", entity.SrcManual)
		require.NotNil(t, other)
		other.FileCount = 3
		require.NoError(t, other.Create())
		t.Cleanup(func() { _ = entity.UnscopedDb().Delete(&entity.Subject{}, "subj_uid = ?", other.SubjUID).Error })

		_, out, err := listPeople(context.Background(), sessionRequest(admin), ListPeopleInput{Limit: 50})
		require.NoError(t, err)
		assert.NotContains(t, personUIDs(out.Items), other.SubjUID)
	})
	t.Run("Cover", func(t *testing.T) {
		_, all, err := listPeople(context.Background(), sessionRequest(admin), ListPeopleInput{Limit: 50})
		require.NoError(t, err)
		require.NotEmpty(t, all.Items)
		uid := all.Items[0].UID

		var cover struct {
			Hash     string
			PhotoUID string
		}
		require.NoError(t, entity.Db().Table("files").Select("files.file_hash AS hash, photos.photo_uid").
			Joins("JOIN photos ON photos.id = files.photo_id AND photos.deleted_at IS NULL").
			Where("files.file_hash <> '' AND files.deleted_at IS NULL").Order("files.id").Limit(1).Scan(&cover).Error)

		var thumb string
		require.NoError(t, entity.Db().Table("subjects").Select("thumb").Where("subj_uid = ?", uid).Row().Scan(&thumb))
		require.NoError(t, entity.Db().Model(&entity.Subject{}).Where("subj_uid = ?", uid).UpdateColumn("thumb", cover.Hash+"-0910162fd2fd").Error)
		t.Cleanup(func() {
			_ = entity.Db().Model(&entity.Subject{}).Where("subj_uid = ?", uid).UpdateColumn("thumb", thumb).Error
		})

		_, out, err := listPeople(context.Background(), sessionRequest(admin), ListPeopleInput{Limit: 50})
		require.NoError(t, err)
		require.Equal(t, uid, out.Items[0].UID)
		assert.Equal(t, cover.PhotoUID, out.Items[0].Cover)
	})
	t.Run("Invalid", func(t *testing.T) {
		_, _, err := listPeople(context.Background(), sessionRequest(admin), ListPeopleInput{Limit: -1})
		assert.Error(t, err)
		_, _, err = listPeople(context.Background(), &sdkmcp.CallToolRequest{}, ListPeopleInput{})
		assert.Error(t, err)
	})
}

// personUIDs returns the UIDs of the people.
func personUIDs(people []Person) (uids []string) {
	for _, p := range people {
		uids = append(uids, p.UID)
	}

	return uids
}

func TestListLabels(t *testing.T) {
	admin := entity.SessionFixtures.Pointer("alice")

	t.Run("SameAsSearch", func(t *testing.T) {
		require.NoError(t, entity.UpdateLabelCounts())

		for _, all := range []bool{false, true} {
			expected, err := search.UserLabels(form.SearchLabels{All: all, Count: search.MaxResults}, admin)
			require.NoError(t, err)
			require.NotEmpty(t, expected)

			_, out, err := listLabels(context.Background(), sessionRequest(admin), ListLabelsInput{All: all, Limit: 50})
			require.NoError(t, err)
			assert.Equal(t, len(expected), out.Total)
			assert.Len(t, out.Items, min(len(expected), 50))

			for i, l := range out.Items {
				assert.Equal(t, expected[i].LabelUID, l.UID)
				assert.Equal(t, expected[i].LabelName, l.Name)
				assert.Equal(t, expected[i].PhotoCount, l.MediaCount)
				assert.NotEmpty(t, l.Slug)
			}
		}
	})
	t.Run("PrivateLabels", func(t *testing.T) {
		var label entity.Label
		require.NoError(t, entity.Db().Where("photo_count > 1 AND label_nsfw = ? AND deleted_at IS NULL", false).Order("id").First(&label).Error)
		require.NoError(t, entity.Db().Model(&entity.Label{}).Where("id = ?", label.ID).UpdateColumn("label_nsfw", true).Error)
		t.Cleanup(func() {
			_ = entity.Db().Model(&entity.Label{}).Where("id = ?", label.ID).UpdateColumn("label_nsfw", false).Error
		})

		_, out, err := listLabels(context.Background(), sessionRequest(admin), ListLabelsInput{All: true, Limit: 50})
		require.NoError(t, err)
		assert.Contains(t, labelUIDs(out.Items), label.LabelUID)

		client := &entity.Session{AuthProvider: authn.ProviderClient.String(), AuthScope: "*"}
		_, out, err = listLabels(context.Background(), sessionRequest(client), ListLabelsInput{All: true, Limit: 50})
		require.NoError(t, err)
		assert.NotContains(t, labelUIDs(out.Items), label.LabelUID)
	})
	t.Run("QueryIsText", func(t *testing.T) {
		// Filter syntax in the query cannot turn off the public-only rule.
		var label entity.Label
		require.NoError(t, entity.Db().Where("photo_count > 1 AND label_nsfw = ? AND deleted_at IS NULL", false).Order("id").First(&label).Error)
		require.NoError(t, entity.Db().Model(&entity.Label{}).Where("id = ?", label.ID).UpdateColumn("label_nsfw", true).Error)
		t.Cleanup(func() {
			_ = entity.Db().Model(&entity.Label{}).Where("id = ?", label.ID).UpdateColumn("label_nsfw", false).Error
		})

		client := &entity.Session{AuthProvider: authn.ProviderClient.String(), AuthScope: "*"}
		for _, query := range []string{"public:false", "nsfw:true", "public:false nsfw:true all:true"} {
			_, out, err := listLabels(context.Background(), sessionRequest(client), ListLabelsInput{Query: query, All: true, Limit: 50})
			require.NoError(t, err, query)
			assert.NotContains(t, labelUIDs(out.Items), label.LabelUID, query)
		}
	})
	t.Run("Query", func(t *testing.T) {
		_, all, err := listLabels(context.Background(), sessionRequest(admin), ListLabelsInput{All: true, Limit: 50})
		require.NoError(t, err)
		_, out, err := listLabels(context.Background(), sessionRequest(admin), ListLabelsInput{Query: "cake", All: true})
		require.NoError(t, err)
		require.NotEmpty(t, out.Items)
		assert.Less(t, out.Total, all.Total)
	})
	t.Run("CustomSlug", func(t *testing.T) {
		_, out, err := listLabels(context.Background(), sessionRequest(admin), ListLabelsInput{Query: "cake", All: true})
		require.NoError(t, err)

		found := false
		for _, l := range out.Items {
			if l.Name == "Cake" {
				found = true
				assert.Equal(t, "kuchen", l.Slug)
			}
		}
		assert.True(t, found)
	})
	t.Run("Categories", func(t *testing.T) {
		// Private categories are left out for sessions that see public labels only.
		var category entity.Category
		require.NoError(t, entity.Db().First(&category).Error)

		var child, parent entity.Label
		require.NoError(t, entity.Db().Where("id = ?", category.LabelID).First(&child).Error)
		require.NoError(t, entity.Db().Where("id = ?", category.CategoryID).First(&parent).Error)
		require.NoError(t, entity.Db().Model(&entity.Label{}).Where("id = ?", parent.ID).UpdateColumn("label_nsfw", true).Error)
		t.Cleanup(func() {
			_ = entity.Db().Model(&entity.Label{}).Where("id = ?", parent.ID).UpdateColumn("label_nsfw", parent.LabelNSFW).Error
		})

		categoriesOf := func(s *entity.Session) []string {
			_, out, err := listLabels(context.Background(), sessionRequest(s), ListLabelsInput{All: true, Limit: 50, Query: child.LabelName})
			require.NoError(t, err)
			for _, l := range out.Items {
				if l.UID == child.LabelUID {
					return l.Categories
				}
			}
			t.Fatalf("label %s not listed", child.LabelName)
			return nil
		}

		assert.Contains(t, categoriesOf(admin), parent.LabelName)
		assert.NotContains(t, categoriesOf(&entity.Session{AuthProvider: authn.ProviderClient.String(), AuthScope: "*"}), parent.LabelName)
	})
	t.Run("Invalid", func(t *testing.T) {
		_, _, err := listLabels(context.Background(), sessionRequest(admin), ListLabelsInput{Offset: -1})
		assert.Error(t, err)
		_, _, err = listLabels(context.Background(), &sdkmcp.CallToolRequest{}, ListLabelsInput{})
		assert.Error(t, err)
	})
}

// labelUIDs returns the UIDs of the labels.
func labelUIDs(labels []Label) (uids []string) {
	for _, l := range labels {
		uids = append(uids, l.UID)
	}

	return uids
}

func TestLabelCategories(t *testing.T) {
	var category entity.Category
	require.NoError(t, entity.Db().First(&category).Error)

	var parent entity.Label
	require.NoError(t, entity.Db().Where("id = ?", category.CategoryID).First(&parent).Error)

	assert.Contains(t, labelCategories([]uint{category.LabelID}, false)[category.LabelID], parent.LabelName)
	assert.Empty(t, labelCategories(nil, false))

	// Deleted categories are left out.
	require.NoError(t, entity.Db().Model(&entity.Label{}).Where("id = ?", parent.ID).UpdateColumn("deleted_at", entity.Now()).Error)
	t.Cleanup(func() {
		_ = entity.UnscopedDb().Model(&entity.Label{}).Where("id = ?", parent.ID).UpdateColumn("deleted_at", nil).Error
	})
	assert.NotContains(t, labelCategories([]uint{category.LabelID}, false)[category.LabelID], parent.LabelName)
	require.NoError(t, entity.UnscopedDb().Model(&entity.Label{}).Where("id = ?", parent.ID).UpdateColumn("deleted_at", nil).Error)

	require.NoError(t, entity.Db().Model(&entity.Label{}).Where("id = ?", parent.ID).UpdateColumn("label_nsfw", true).Error)
	t.Cleanup(func() {
		_ = entity.Db().Model(&entity.Label{}).Where("id = ?", parent.ID).UpdateColumn("label_nsfw", parent.LabelNSFW).Error
	})

	assert.Contains(t, labelCategories([]uint{category.LabelID}, false)[category.LabelID], parent.LabelName)
	assert.NotContains(t, labelCategories([]uint{category.LabelID}, true)[category.LabelID], parent.LabelName)
}

func TestRegisterPeopleAndLabelTools(t *testing.T) {
	server := sdkmcp.NewServer(&sdkmcp.Implementation{Name: "test"}, nil)
	registerPeopleAndLabelTools(server)

	access, ok := ToolDeclaration("list_people")
	assert.True(t, ok)
	assert.Equal(t, ToolAccess{Resource: acl.ResourcePeople, Action: acl.ActionSearch}, access)

	access, ok = ToolDeclaration("list_labels")
	assert.True(t, ok)
	assert.Equal(t, ToolAccess{Resource: acl.ResourceLabels, Action: acl.ActionSearch}, access)

	assert.ElementsMatch(t, []string{"list_people", "list_labels"}, toolNames(t, connectServer(t, server)))
}
