package rnd

// ContainsUID checks if a slice of strings contains ContainsUID only.
func ContainsUID(s []string, prefix byte) bool {
	if len(s) < 1 {
		return false
	}

	for _, id := range s {
		switch prefix {
		case 0:
			if !IsUnique(id, prefix) {
				return false
			}
		default:
			if !IsUID(id, prefix) {
				return false
			}
		}
	}

	return true
}

// ContainsType checks if a slice of strings contains only random IDs of a given type and returns it.
func ContainsType(ids []string) (idType Type, idPrefix byte) {
	if len(ids) < 1 {
		return TypeEmpty, PrefixNone
	}

	idType = TypeUnknown
	idPrefix = PrefixNone

	for _, id := range ids {
		t, p := IdType(id)

		if t == TypeUnknown {
			return TypeUnknown, PrefixNone
		} else if idType == TypeUnknown {
			idType = t
		} else if idType != t {
			return TypeMixed, PrefixNone
		}

		if t != TypeUID {
			continue
		} else if idPrefix == PrefixNone {
			idPrefix = p
		} else if idPrefix != PrefixMixed && idPrefix != p {
			idPrefix = PrefixMixed
		}
	}

	return idType, idPrefix
}
