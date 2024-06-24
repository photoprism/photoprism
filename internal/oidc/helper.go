package oidc

import (
	"errors"
	"strings"
)

type UserInfo interface {
	GetPreferredUsername() string
	GetNickname() string
	GetName() string
	GetEmail() string
	GetClaim(key string) interface{}
}

func UsernameFromUserInfo(userinfo UserInfo) (username string) {
	if len(userinfo.GetPreferredUsername()) >= 4 {
		username = userinfo.GetPreferredUsername()
	} else if len(userinfo.GetNickname()) >= 4 {
		username = userinfo.GetNickname()
	} else if len(userinfo.GetName()) >= 4 {
		username = strings.ReplaceAll(strings.ToLower(userinfo.GetName()), " ", "-")
	} else if len(userinfo.GetEmail()) >= 4 {
		username = userinfo.GetEmail()
	} else {
		log.Error("oidc: no username found")
	}
	return username
}

// HasRoleAdmin searches UserInfo claims for admin role.
// Returns true if role is present or false if claim was found but no role in there.
// Error will be returned if the role claim is not delivered at all.
func HasRoleAdmin(userinfo UserInfo) (bool, error) {
	claim := userinfo.GetClaim(RoleClaim)
	return claimContainsProp(claim, AdminRole)
}

func claimContainsProp(claim interface{}, property string) (bool, error) {
	switch t := claim.(type) {
	case nil:
		return false, errors.New("oidc: claim not found")
	case []interface{}:
		for _, value := range t {
			res, err := claimContainsProp(value, property)
			if err != nil {
				return false, err
			}
			if res {
				return res, nil
			}
		}
		return false, nil
	case interface{}:
		if value, ok := t.(string); ok {
			return value == property, nil
		} else {
			return false, errors.New("oidc: unexpected type")
		}
	default:
		return false, errors.New("oidc: unexpected type")
	}
}
