package mcp

import (
	"context"
	"errors"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/form"
)

// Declarations of the people and label tools.
var (
	listPeopleAccess = ToolAccess{Resource: acl.ResourcePeople, Action: acl.ActionSearch}
	listLabelsAccess = ToolAccess{Resource: acl.ResourceLabels, Action: acl.ActionSearch}
)

// Person is a person the face recognition or a user has named.
type Person struct {
	UID        string `json:"uid"`
	Name       string `json:"name"`
	MediaCount int    `json:"media_count"`
	Cover      string `json:"cover,omitempty"`
}

// Label is a label of media, as returned by list_labels.
type Label struct {
	UID        string   `json:"uid"`
	Name       string   `json:"name"`
	Slug       string   `json:"slug"`
	MediaCount int      `json:"media_count"`
	Categories []string `json:"categories,omitempty"`
}

// ListPeopleInput defines the inputs of the list_people tool.
type ListPeopleInput struct {
	Query  string `json:"query,omitempty" jsonschema:"optional text to search in names and aliases"`
	Limit  int    `json:"limit,omitempty" jsonschema:"optional maximum number of people to return (default 20, max 50)"`
	Offset int    `json:"offset,omitempty" jsonschema:"optional number of people to skip"`
}

// ListPeopleOutput defines the structured output of the list_people tool.
type ListPeopleOutput struct {
	Items     []Person `json:"items"`
	Total     int      `json:"total"`
	Limit     int      `json:"limit"`
	Offset    int      `json:"offset"`
	Truncated bool     `json:"truncated"`
	Warnings  []string `json:"warnings,omitempty"`
}

// ListLabelsInput defines the inputs of the list_labels tool.
type ListLabelsInput struct {
	Query  string `json:"query,omitempty" jsonschema:"optional search in label names"`
	All    bool   `json:"all,omitempty" jsonschema:"optional; include labels with low priority or a single media"`
	Limit  int    `json:"limit,omitempty" jsonschema:"optional maximum number of labels to return (default 20, max 50)"`
	Offset int    `json:"offset,omitempty" jsonschema:"optional number of labels to skip"`
}

// ListLabelsOutput defines the structured output of the list_labels tool.
type ListLabelsOutput struct {
	Items     []Label  `json:"items"`
	Total     int      `json:"total"`
	Limit     int      `json:"limit"`
	Offset    int      `json:"offset"`
	Truncated bool     `json:"truncated"`
	Warnings  []string `json:"warnings,omitempty"`
}

// validateListInput validates the paging and query inputs of a list tool.
func validateListInput(limit, offset int, query string) (int, int, error) {
	limit, err := validateLimit(limit)

	if err != nil {
		return 0, 0, err
	}

	if offset, err = validateOffset(offset); err != nil {
		return 0, 0, err
	}

	if _, err = validateString("query", query); err != nil {
		return 0, 0, err
	}

	return limit, offset, nil
}

// listPeople returns the people the session may see, like GET /api/v1/subjects.
func listPeople(_ context.Context, req *sdkmcp.CallToolRequest, input ListPeopleInput) (*sdkmcp.CallToolResult, ListPeopleOutput, error) {
	s := RequestSession(req)

	if s == nil {
		return nil, ListPeopleOutput{}, errors.New("list_people requires an authenticated session")
	}

	limit, offset, err := validateListInput(input.Limit, input.Offset, input.Query)

	if err != nil {
		return nil, ListPeopleOutput{}, err
	}

	// Like the people page, list people who appear in at least one file.
	results, err := search.UserSubjects(form.SearchSubjects{Query: plainQuery(input.Query), Type: entity.SubjPerson, Files: 1, Count: search.MaxResults}, s)

	if err != nil {
		return nil, ListPeopleOutput{}, errors.New("people could not be listed")
	}

	page, truncated := pageOf(results, limit, offset)
	hashes := make([]string, len(page))

	for i, p := range page {
		hashes[i] = p.Thumb
	}

	covers := coverUIDs(hashes, s)
	items := make([]Person, len(page))

	for i, p := range page {
		items[i] = Person{UID: p.SubjUID, Name: p.SubjName, MediaCount: p.PhotoCount, Cover: covers[p.Thumb]}
	}

	return nil, ListPeopleOutput{
		Items:     items,
		Total:     len(results),
		Limit:     limit,
		Offset:    offset,
		Truncated: truncated,
		Warnings:  pageWarnings(len(items), len(results), offset, truncated),
	}, nil
}

// listLabels returns the labels the session may see, like GET /api/v1/labels.
func listLabels(_ context.Context, req *sdkmcp.CallToolRequest, input ListLabelsInput) (*sdkmcp.CallToolResult, ListLabelsOutput, error) {
	s := RequestSession(req)

	if s == nil {
		return nil, ListLabelsOutput{}, errors.New("list_labels requires an authenticated session")
	}

	limit, offset, err := validateListInput(input.Limit, input.Offset, input.Query)

	if err != nil {
		return nil, ListLabelsOutput{}, err
	}

	// Update the picture counts of the labels if needed, as the labels endpoint does.
	if countErr := entity.UpdateLabelCountsIfNeeded(); countErr != nil {
		log.Warnf("mcp: could not update label counts (%s)", countErr)
	}

	results, err := search.UserLabels(form.SearchLabels{Query: plainQuery(input.Query), All: input.All, Count: search.MaxResults}, s)

	if err != nil {
		return nil, ListLabelsOutput{}, errors.New("labels could not be listed")
	}

	page, truncated := pageOf(results, limit, offset)
	ids := make([]uint, len(page))

	for i, l := range page {
		ids[i] = l.ID
	}

	categories := labelCategories(ids, search.LabelsPublicOnly(s))
	items := make([]Label, len(page))

	for i, l := range page {
		slug := l.CustomSlug

		if slug == "" {
			slug = l.LabelSlug
		}

		items[i] = Label{UID: l.LabelUID, Name: l.LabelName, Slug: slug, MediaCount: l.PhotoCount, Categories: categories[l.ID]}
	}

	return nil, ListLabelsOutput{
		Items:     items,
		Total:     len(results),
		Limit:     limit,
		Offset:    offset,
		Truncated: truncated,
		Warnings:  pageWarnings(len(items), len(results), offset, truncated),
	}, nil
}

// labelCategories returns the names of the categories of the labels, without private ones if publicOnly is set.
func labelCategories(ids []uint, publicOnly bool) map[uint][]string {
	result := make(map[uint][]string, len(ids))

	if len(ids) == 0 {
		return result
	}

	var rows []struct {
		LabelID uint
		Name    string
		NSFW    bool
	}

	entity.Log("mcp", "find label categories", entity.Db().Table("categories c").
		Select("c.label_id, l.label_name AS name, l.label_nsfw AS nsfw").
		Joins("JOIN labels l ON l.id = c.category_id AND l.deleted_at IS NULL").
		Where("c.label_id IN (?)", ids).Order("c.label_id, l.label_name").Scan(&rows).Error)

	for _, r := range rows {
		if !publicOnly || !r.NSFW {
			result[r.LabelID] = append(result[r.LabelID], r.Name)
		}
	}

	return result
}

// registerPeopleAndLabelTools registers the list_people and list_labels tools.
func registerPeopleAndLabelTools(server *sdkmcp.Server) {
	AddTool(server, listPeopleAccess, &sdkmcp.Tool{
		Name: "list_people",
		Description: "Lists the people the account may see, with the number of media each appears in and a cover " +
			"media UID. Pass a person UID to search_media as person to find their media. " +
			"Returns at most limit items (default 20, max 50) starting at offset.",
		Annotations: readOnly(),
	}, listPeople)

	AddTool(server, listLabelsAccess, &sdkmcp.Tool{
		Name: "list_labels",
		Description: "Lists the labels the account may see, with their slug, the number of media, and categories. " +
			"Search media by a label with the label:<slug> filter of search_media. Set all to include labels " +
			"with low priority or a single media. Returns at most limit items (default 20, max 50) starting at offset.",
		Annotations: readOnly(),
	}, listLabels)
}
