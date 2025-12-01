package batch

// Item represents batch edit value item.
type Item struct {
	Value  string `json:"value"`
	Title  string `json:"title"`
	Mixed  bool   `json:"mixed"`
	Action Action `json:"action"`
}

// Items represents batch edit value items.
type Items struct {
	Items  []Item `json:"items"`
	Mixed  bool   `json:"mixed"`
	Action Action `json:"action"`
}

// ResolveValuesByTitle replaces empty values with resolver results so callers
// can pre-create referenced entities (for example albums) once per unique
// title instead of repeating the same lookup for every photo.
func (it *Items) ResolveValuesByTitle(resolver func(title, action string) string) {
	if it == nil || resolver == nil || len(it.Items) == 0 {
		return
	}

	type cacheKey struct {
		title  string
		action Action
	}

	cache := make(map[cacheKey]string)

	for i := range it.Items {
		item := &it.Items[i]
		if item.Value != "" || item.Title == "" {
			continue
		}

		key := cacheKey{title: item.Title, action: item.Action}

		if val, ok := cache[key]; ok {
			if val != "" {
				item.Value = val
			}
			continue
		}

		resolved := resolver(item.Title, item.Action)
		cache[key] = resolved
		if resolved != "" {
			item.Value = resolved
		}
	}
}

// GetValuesByActions returns the non-empty values for items whose action is
// included in the provided filter. The original ordering is preserved so the
// caller can correlate results with the source payload.
func (it *Items) GetValuesByActions(actions []Action) []string {
	if it == nil || len(it.Items) == 0 {
		return nil
	}

	actionFilter := actionSet(actions)
	if len(actionFilter) == 0 {
		return nil
	}

	values := make([]string, 0, len(it.Items))

	for _, item := range it.Items {
		if _, ok := actionFilter[item.Action]; !ok {
			continue
		}

		if item.Value == "" {
			continue
		}

		values = append(values, item.Value)
	}

	if len(values) == 0 {
		return nil
	}

	return values
}

// GetItemsByActions returns all items whose action matches any entry in the
// provided filter, preserving their original order.
func (it *Items) GetItemsByActions(actions []Action) []Item {
	if it == nil || len(it.Items) == 0 {
		return nil
	}

	actionFilter := actionSet(actions)
	if len(actionFilter) == 0 {
		return nil
	}

	filtered := make([]Item, 0, len(it.Items))

	for _, item := range it.Items {
		if _, ok := actionFilter[item.Action]; ok {
			filtered = append(filtered, item)
		}
	}

	if len(filtered) == 0 {
		return nil
	}

	return filtered
}

func actionSet(actions []Action) map[Action]struct{} {
	if len(actions) == 0 {
		return nil
	}

	m := make(map[Action]struct{}, len(actions))
	for _, act := range actions {
		m[act] = struct{}{}
	}

	return m
}
