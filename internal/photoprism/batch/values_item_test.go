package batch

import (
	"reflect"
	"testing"
)

func TestItemsGetValuesByActions(t *testing.T) {
	t.Parallel()

	items := Items{
		Items: []Item{
			{Value: "uid-add", Action: ActionAdd},
			{Value: "uid-remove", Action: ActionRemove},
			{Value: "", Action: ActionAdd},
			{Value: "uid-update", Action: ActionUpdate},
		},
	}

	want := []string{"uid-add", "uid-remove"}
	got := items.GetValuesByActions([]Action{ActionAdd, ActionRemove})

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetValuesByActions() = %v, want %v", got, want)
	}

	if vals := items.GetValuesByActions(nil); vals != nil {
		t.Fatalf("expected nil slice when no actions provided, got %v", vals)
	}

	var empty *Items
	if vals := empty.GetValuesByActions([]Action{ActionAdd}); vals != nil {
		t.Fatalf("expected nil slice for nil receiver, got %v", vals)
	}
}

func TestItemsGetItemsByActions(t *testing.T) {
	t.Parallel()

	items := Items{
		Items: []Item{
			{Value: "uid-add", Title: "Add", Action: ActionAdd},
			{Value: "uid-remove", Title: "Remove", Action: ActionRemove},
			{Value: "uid-skip", Title: "Skip", Action: ActionNone},
		},
	}

	want := []Item{
		{Value: "uid-add", Title: "Add", Action: ActionAdd},
		{Value: "uid-remove", Title: "Remove", Action: ActionRemove},
	}

	got := items.GetItemsByActions([]Action{ActionAdd, ActionRemove})
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetItemsByActions() = %v, want %v", got, want)
	}

	if filtered := items.GetItemsByActions([]Action{ActionUpdate}); filtered != nil {
		t.Fatalf("expected nil when no items match, got %v", filtered)
	}

	var nilItems *Items
	if filtered := nilItems.GetItemsByActions([]Action{ActionAdd}); filtered != nil {
		t.Fatalf("expected nil for nil receiver, got %v", filtered)
	}
}

func TestItemsResolveValuesByTitle(t *testing.T) {
	t.Parallel()

	items := Items{
		Items: []Item{
			{Title: "Trips", Action: ActionRemove},
			{Title: "Trips", Action: ActionAdd},
			{Title: "Trips", Action: ActionAdd},
			{Title: "Archived", Value: "album-arch", Action: ActionAdd},
			{Title: "", Action: ActionAdd},
			{Title: "Drafts", Action: ActionAdd},
			{Title: "Trips", Action: ActionRemove},
			{Title: "Trips", Action: ActionUpdate},
		},
	}

	var callLog []string
	resolver := func(title, action string) string {
		callLog = append(callLog, title+":"+action)
		if action != ActionAdd {
			return ""
		}

		resolved := map[string]string{
			"Trips":  "album-trips",
			"Drafts": "",
		}[title]
		return resolved
	}

	items.ResolveValuesByTitle(resolver)

	expectedCalls := []string{"Trips:remove", "Trips:add", "Drafts:add", "Trips:update"}
	if !reflect.DeepEqual(callLog, expectedCalls) {
		t.Fatalf("unexpected resolver calls %v, want %v", callLog, expectedCalls)
	}

	if got := items.Items[0].Value; got != "" {
		t.Fatalf("expected remove action to remain empty, got %q", got)
	}

	if got := items.Items[1].Value; got != "album-trips" {
		t.Fatalf("expected first add action to resolve value, got %q", got)
	}

	if got := items.Items[2].Value; got != "album-trips" {
		t.Fatalf("expected duplicate add to reuse cached value, got %q", got)
	}

	if got := items.Items[3].Value; got != "album-arch" {
		t.Fatalf("expected existing value to remain, got %q", got)
	}

	if got := items.Items[4].Value; got != "" {
		t.Fatalf("expected empty title to remain untouched, got %q", got)
	}

	if got := items.Items[5].Value; got != "" {
		t.Fatalf("expected Drafts resolver empty result, got %q", got)
	}

	if got := items.Items[6].Value; got != "" {
		t.Fatalf("expected cached remove action to stay empty, got %q", got)
	}

	if got := items.Items[7].Value; got != "" {
		t.Fatalf("expected update action to stay empty, got %q", got)
	}

	items.ResolveValuesByTitle(nil)
	var nilItems *Items
	nilItems.ResolveValuesByTitle(func(string, string) string { t.Fatal("should not be called"); return "" })
}
