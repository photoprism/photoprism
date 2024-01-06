package rnd

import (
	"strings"
)

const (
	TypeEmpty     Type = "empty"
	TypeMixed     Type = "mixed"
	TypeUUID      Type = "UUID"
	TypeUID       Type = "UID"
	TypeRefID     Type = "RID"
	TypeSessionID Type = "SID"
	TypeCrcToken  Type = "CRC"
	TypeMD5       Type = "MD5"
	TypeSHA1      Type = "SHA1"
	TypeSHA224    Type = "SHA224"
	TypeSHA256    Type = "SHA256"
	TypeSHA384    Type = "SHA384"
	TypeSHA512    Type = "SHA512"
	TypeUnknown   Type = "unknown"
)

// IdType checks what kind of random ID a string contains
// and returns it along with the id prefix, if any.
func IdType(id string) (Type, byte) {
	if l := len(id); l == 0 {
		return TypeEmpty, PrefixNone
	} else if l < 14 || l > 128 {
		return TypeUnknown, PrefixNone
	}

	switch {
	case IsUID(id, 0):
		return TypeUID, id[0]
	case IsUUID(id):
		return TypeUUID, PrefixNone
	case IsSHA1(id):
		return TypeSHA1, PrefixNone
	case IsRefID(id):
		return TypeRefID, PrefixNone
	case IsAuthToken(id):
		return TypeSessionID, PrefixNone
	case ValidateCrcToken(id):
		return TypeCrcToken, PrefixNone
	case IsSHA224(id):
		return TypeSHA224, PrefixNone
	case IsSHA256(id):
		return TypeSHA256, PrefixNone
	case IsSHA384(id):
		return TypeSHA384, PrefixNone
	case IsSHA512(id):
		return TypeSHA512, PrefixNone
	case IsMD5(id):
		return TypeMD5, PrefixNone
	default:
		return TypeUnknown, PrefixNone
	}
}

// Type represents a random id type.
type Type string

// String returns the type as string.
func (t Type) String() string {
	return string(t)
}

// Equal checks if the type matches.
func (t Type) Equal(s string) bool {
	return strings.EqualFold(s, t.String())
}

// NotEqual checks if the type is different.
func (t Type) NotEqual(s string) bool {
	return !t.Equal(s)
}

func (t Type) EntityID() bool {
	switch t {
	case TypeUID, TypeUUID, TypeRefID, TypeCrcToken, TypeSessionID:
		return true
	default:
		return false
	}
}

func (t Type) SessionID() bool {
	return t == TypeSessionID
}

func (t Type) CrcToken() bool {
	switch t {
	case TypeCrcToken:
		return true
	default:
		return false
	}
}

func (t Type) Hash() bool {
	switch t {
	case TypeMD5:
		return true
	default:
		return t.SHA()
	}
}

func (t Type) SHA() bool {
	return t.SHA1() || t.SHA2()
}

func (t Type) SHA1() bool {
	switch t {
	case TypeSHA1:
		return true
	default:
		return false
	}
}

func (t Type) SHA2() bool {
	switch t {
	case TypeSHA224, TypeSHA256, TypeSHA384, TypeSHA512:
		return true
	default:
		return false
	}
}

// Unknown checks if the type is unknown.
func (t Type) Unknown() bool {
	return t == TypeUnknown
}
