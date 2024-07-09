package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestNewUser(t *testing.T) {
	m := NewUser()

	assert.True(t, rnd.IsRefID(m.RefID))
	assert.True(t, rnd.IsUID(m.UserUID, UserUID))
}

func TestOidcUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.Name = "Jane Doe"
		info.GivenName = "Jane"
		info.FamilyName = "Doe"
		info.Email = "jane@doe.com"
		info.EmailVerified = true
		info.Subject = "e3a9f4a6-9d60-47cb-9bf5-02bd15b0c68d"
		info.PreferredUsername = "Jane Doe"

		m := OidcUser(info, "", "jane.doe")

		assert.Equal(t, "oidc", m.AuthProvider)
		assert.Equal(t, "", m.AuthIssuer)
		assert.Equal(t, "e3a9f4a6-9d60-47cb-9bf5-02bd15b0c68d", m.AuthID)
		assert.Equal(t, "jane@doe.com", m.UserEmail)
		assert.Equal(t, "jane.doe", m.UserName)
		assert.Equal(t, "Jane Doe", m.DisplayName)
	})
	t.Run("NoUsername", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.Name = "Jane Doe"
		info.GivenName = "Jane"
		info.FamilyName = "Doe"
		info.Email = "jane@doe.com"
		info.EmailVerified = true
		info.Subject = "e3a9f4a6-9d60-47cb-9bf5-02bd15b0c68d"
		info.PreferredUsername = "Jane Doe"

		m := OidcUser(info, "https://accounts.google.com", "")

		assert.Equal(t, "oidc", m.AuthProvider)
		assert.Equal(t, "https://accounts.google.com", m.AuthIssuer)
		assert.Equal(t, "e3a9f4a6-9d60-47cb-9bf5-02bd15b0c68d", m.AuthID)
		assert.Equal(t, "jane@doe.com", m.UserEmail)
		assert.Equal(t, "", m.UserName)
		assert.Equal(t, "Jane Doe", m.DisplayName)
	})
	t.Run("NoSubject", func(t *testing.T) {
		info := &oidc.UserInfo{}
		info.Name = "Jane Doe"
		info.GivenName = "Jane"
		info.FamilyName = "Doe"
		info.Nickname = "Jens Mander"
		info.Email = "jane@doe.com"
		info.EmailVerified = true
		info.Subject = ""

		m := OidcUser(info, "https://accounts.google.com", "jane.doe")

		assert.Equal(t, "", m.AuthProvider)
		assert.Equal(t, "", m.AuthIssuer)
		assert.Equal(t, "", m.AuthID)
		assert.Equal(t, "", m.UserEmail)
		assert.Equal(t, "", m.UserName)
		assert.Equal(t, "", m.DisplayName)
	})
}

func TestLdapUser(t *testing.T) {
	m := LdapUser("user-ldap", "ldap@test.com")
	assert.Equal(t, "ldap", m.AuthProvider)
}

func TestFindLocalUser(t *testing.T) {
	t.Run("Admin", func(t *testing.T) {
		m := FindLocalUser("admin")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, 1, m.ID)
		assert.NotEmpty(t, m.UserUID)
		assert.Equal(t, "admin", m.UserName)
		assert.Equal(t, "admin", m.Username())
		m.UserName = "Admin "
		assert.Equal(t, "admin", m.Username())
		assert.Equal(t, "Admin ", m.UserName)
		assert.Equal(t, "Admin", m.DisplayName)
		assert.Equal(t, acl.RoleAdmin, m.AclRole())
		assert.Equal(t, "", m.Attr())
		assert.False(t, m.IsVisitor())
		assert.False(t, m.HasSharedAccessOnly(acl.ResourcePhotos))
		assert.False(t, m.HasSharedAccessOnly(acl.ResourceAlbums))
		assert.True(t, m.SuperAdmin)
		assert.True(t, m.CanLogin)
		assert.True(t, m.CanInvite)
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})

	t.Run("Alice", func(t *testing.T) {
		m := FindLocalUser("alice")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, 5, m.ID)
		assert.Equal(t, "uqxetse3cy5eo9z2", m.UserUID)
		assert.Equal(t, "alice", m.UserName)
		assert.Equal(t, "Alice", m.DisplayName)
		assert.Equal(t, "alice@example.com", m.UserEmail)
		assert.True(t, m.SuperAdmin)
		assert.Equal(t, acl.RoleAdmin, m.AclRole())
		assert.NotEqual(t, acl.RoleVisitor, m.AclRole())
		assert.False(t, m.IsVisitor())
		assert.False(t, m.HasSharedAccessOnly(acl.ResourcePhotos))
		assert.False(t, m.HasSharedAccessOnly(acl.ResourceAlbums))
		assert.True(t, m.CanLogin)
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})

	t.Run("Bob", func(t *testing.T) {
		m := FindLocalUser("bob")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, 7, m.ID)
		assert.Equal(t, "uqxc08w3d0ej2283", m.UserUID)
		assert.Equal(t, "bob", m.UserName)
		assert.Equal(t, "Robert Rich", m.DisplayName)
		assert.Equal(t, "bob@example.com", m.UserEmail)
		assert.False(t, m.SuperAdmin)
		assert.False(t, m.IsVisitor())
		assert.False(t, m.HasSharedAccessOnly(acl.ResourcePhotos))
		assert.False(t, m.HasSharedAccessOnly(acl.ResourceAlbums))
		assert.True(t, m.CanLogin)
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})

	t.Run("Unknown", func(t *testing.T) {
		m := FindLocalUser("")

		if m != nil {
			t.Fatal("result should be nil")
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		m := FindLocalUser("xxx")

		if m != nil {
			t.Fatal("result should be nil")
		}
	})
}

func TestFindUserByName(t *testing.T) {
	t.Run("Admin", func(t *testing.T) {
		m := FindUserByName("admin")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, 1, m.ID)
		assert.NotEmpty(t, m.UserUID)
		assert.Equal(t, "admin", m.UserName)
		assert.Equal(t, "admin", m.Username())
		m.UserName = "Admin "
		assert.Equal(t, "admin", m.Username())
		assert.Equal(t, "Admin ", m.UserName)
		assert.Equal(t, "Admin", m.DisplayName)
		assert.Equal(t, acl.RoleAdmin, m.AclRole())
		assert.Equal(t, "", m.Attr())
		assert.False(t, m.IsVisitor())
		assert.True(t, m.SuperAdmin)
		assert.True(t, m.CanLogin)
		assert.True(t, m.CanInvite)
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})

	t.Run("Alice", func(t *testing.T) {
		m := FindUserByName("alice")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, 5, m.ID)
		assert.Equal(t, "uqxetse3cy5eo9z2", m.UserUID)
		assert.Equal(t, "alice", m.UserName)
		assert.Equal(t, "Alice", m.DisplayName)
		assert.Equal(t, "alice@example.com", m.UserEmail)
		assert.True(t, m.SuperAdmin)
		assert.Equal(t, acl.RoleAdmin, m.AclRole())
		assert.NotEqual(t, acl.RoleVisitor, m.AclRole())
		assert.False(t, m.IsVisitor())
		assert.True(t, m.CanLogin)
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})

	t.Run("Bob", func(t *testing.T) {
		m := FindUserByName("bob")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, 7, m.ID)
		assert.Equal(t, "uqxc08w3d0ej2283", m.UserUID)
		assert.Equal(t, "bob", m.UserName)
		assert.Equal(t, "Robert Rich", m.DisplayName)
		assert.Equal(t, "bob@example.com", m.UserEmail)
		assert.False(t, m.SuperAdmin)
		assert.False(t, m.IsVisitor())
		assert.True(t, m.CanLogin)
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})

	t.Run("Unknown", func(t *testing.T) {
		m := FindUserByName("")

		if m != nil {
			t.Fatal("result should be nil")
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		m := FindUserByName("xxx")

		if m != nil {
			t.Fatal("result should be nil")
		}
	})
}

func TestUser_Create(t *testing.T) {
	t.Run("Slug", func(t *testing.T) {
		var m = User{
			UserName:    "example",
			UserRole:    acl.RoleAdmin.String(),
			DisplayName: "Example",
			SuperAdmin:  false,
			CanLogin:    true,
		}

		if err := m.Create(); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "example", m.Username())
		assert.Equal(t, "example", m.UserName)

		if err := m.UpdateUsername("example-editor"); err == nil {
			t.Fatal("error expected")
		}
	})
	t.Run("NewUser", func(t *testing.T) {
		if err := NewUser().Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestUser_UpdateUsername(t *testing.T) {
	t.Run("Exists", func(t *testing.T) {
		var m = User{
			ID:          2,
			UserName:    "foo",
			UserRole:    acl.RoleAdmin.String(),
			DisplayName: "Admin",
			SuperAdmin:  false,
			CanLogin:    true,
		}

		if err := m.UpdateUsername("admin"); err == nil {
			t.Fatal("error expected")
		} else {
			t.Logf("expected error: %s", err)
		}
	})
	t.Run("Success", func(t *testing.T) {
		var m = User{
			ID:          5000,
			UserName:    "",
			UserRole:    "user",
			DisplayName: "Foo",
			SuperAdmin:  false,
			CanLogin:    true,
		}

		err := m.Save()

		if err != nil {
			t.Fatal(err)
		}

		assert.Nil(t, FindUserByName("bar"))

		err2 := m.UpdateUsername("bar")

		if err2 != nil {
			t.Fatal(err2)
		}

		assert.NotNil(t, FindUserByName("bar"))

	})
}

func TestUser_SetUsername(t *testing.T) {
	t.Run("photoprism", func(t *testing.T) {
		m := FindUserByName("admin")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, "admin", m.Username())
		assert.Equal(t, "admin", m.UserName)

		if err := m.SetUsername("photoprism"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "photoprism", m.Username())
		assert.Equal(t, "photoprism", m.UserName)
	})
	t.Run("system users cannot be modified", func(t *testing.T) {
		assert.Equal(t, "system users cannot be modified", Visitor.SetUsername("newname").Error())
	})
	t.Run("username is empty", func(t *testing.T) {
		assert.Equal(t, "username is empty", Admin.SetUsername("").Error())
	})
	t.Run("same name", func(t *testing.T) {
		assert.Nil(t, Admin.SetUsername("admin"))
	})
}

func TestUser_InvalidPassword(t *testing.T) {
	t.Run("admin", func(t *testing.T) {
		m := FindUserByName("admin")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.False(t, m.InvalidPassword("photoprism"))
	})
	t.Run("admin invalid password", func(t *testing.T) {
		m := FindUserByName("admin")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.True(t, m.InvalidPassword("wrong-password"))
	})
	t.Run("no password existing", func(t *testing.T) {
		p := User{UserUID: "u000000000000010", UserName: "Hans", DisplayName: ""}
		err := p.Save()
		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, p.InvalidPassword("abcdef"))

	})
	t.Run("NotRegistered", func(t *testing.T) {
		p := User{UserUID: "u12", UserName: "", DisplayName: ""}
		assert.True(t, p.InvalidPassword("abcdef"))
	})
	t.Run("PasswordEmpty", func(t *testing.T) {
		p := User{UserUID: "u000000000000011", UserName: "User", DisplayName: ""}
		assert.True(t, p.InvalidPassword(""))
	})
}

func TestUser_Save(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		p := User{}

		err := p.Save()

		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("NewUser", func(t *testing.T) {
		if err := NewUser().Save(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestFirstOrCreateUser(t *testing.T) {
	t.Run("NotExisting", func(t *testing.T) {
		p := &User{ID: 555}

		result := FirstOrCreateUser(p)
		if result == nil {
			t.Fatal("result should not be nil")
		}

		assert.NotEmpty(t, result.ID)

	})
	t.Run("Existing", func(t *testing.T) {
		p := &User{ID: 1234}
		err := p.Save()

		if err != nil {
			t.Fatal(err)
		}

		result := FirstOrCreateUser(p)

		if result == nil {
			t.Fatal("result should not be nil")
		}
		assert.NotEmpty(t, result.ID)
	})
}

func TestFindUser(t *testing.T) {
	t.Run("ID", func(t *testing.T) {
		m := FindUser(User{ID: 1})

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, 1, m.ID)
		assert.NotEmpty(t, m.UserUID)
		assert.Equal(t, "admin", m.UserName)
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})
	t.Run("Admin", func(t *testing.T) {
		m := FindUser(User{ID: 2, UserName: "admin"})

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, 1, m.ID)
		assert.NotEmpty(t, m.UserUID)
		assert.Equal(t, "admin", m.UserName)
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})
	t.Run("UserUID", func(t *testing.T) {
		m := FindUser(User{UserUID: "u000000000000002"})

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, -2, m.ID)
		assert.NotEmpty(t, m.UserUID)
		assert.Equal(t, "", m.UserName)
		assert.Equal(t, "Visitor", m.DisplayName)
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})
	t.Run("UserName", func(t *testing.T) {
		m := FindUser(User{UserName: "admin"})

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, 1, m.ID)
		assert.NotEmpty(t, m.UserUID)
		assert.Equal(t, "admin", m.UserName)
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})
	t.Run("Unknown", func(t *testing.T) {
		m := FindUser(User{})

		if m != nil {
			t.Fatal("result should be nil")
		}
	})
	t.Run("NotFound", func(t *testing.T) {
		m := FindUser(User{UserUID: "xxx"})

		if m != nil {
			t.Fatal("result should be nil")
		}
	})
}

func TestFindUserByUID(t *testing.T) {
	t.Run("Visitor", func(t *testing.T) {
		m := FindUserByUID("u000000000000002")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, -2, m.ID)
		assert.NotEmpty(t, m.UserUID)
		assert.Equal(t, "", m.UserName)
		assert.Equal(t, "Visitor", m.DisplayName)
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})
	t.Run("Unknown", func(t *testing.T) {
		m := FindUserByUID("")

		if m != nil {
			t.Fatal("result should be nil")
		}
	})
	t.Run("NotFound", func(t *testing.T) {
		m := FindUserByUID("xxx")

		if m != nil {
			t.Fatal("result should be nil")
		}
	})
	t.Run("Alice", func(t *testing.T) {
		m := FindUserByUID("uqxetse3cy5eo9z2")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, 5, m.ID)
		assert.Equal(t, "uqxetse3cy5eo9z2", m.UserUID)
		assert.Equal(t, "alice", m.Username())
		assert.Equal(t, "Alice", m.DisplayName)
		assert.Equal(t, "alice@example.com", m.UserEmail)
		assert.True(t, m.SuperAdmin)
		assert.True(t, m.IsAdmin())
		assert.True(t, m.IsSuperAdmin())
		assert.False(t, m.IsVisitor())
		assert.True(t, m.CanLogin)
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})
	t.Run("Bob", func(t *testing.T) {
		m := FindUserByUID("uqxc08w3d0ej2283")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, 7, m.ID)
		assert.Equal(t, "uqxc08w3d0ej2283", m.UserUID)
		assert.Equal(t, "bob", m.UserName)
		assert.Equal(t, "Robert Rich", m.DisplayName)
		assert.Equal(t, "bob@example.com", m.UserEmail)
		assert.False(t, m.SuperAdmin)
		assert.True(t, m.IsAdmin())
		assert.False(t, m.IsSuperAdmin())
		assert.False(t, m.IsVisitor())
		assert.True(t, m.CanLogin)
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})
	t.Run("Friend", func(t *testing.T) {
		m := FindUserByUID("uqxqg7i1kperxvu7")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, 8, m.ID)
		assert.Equal(t, "uqxqg7i1kperxvu7", m.UserUID)
		assert.False(t, m.SuperAdmin)
		assert.True(t, m.IsAdmin())
		assert.False(t, m.IsVisitor())
		assert.True(t, m.CanLogin)
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
		assert.True(t, m.HasUID())
		assert.False(t, m.InvalidUID())
	})
}

func TestUser_SameUID(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		m := FindUserByUID("uqxc08w3d0ej2283")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.True(t, m.SameUID("uqxc08w3d0ej2283"))
		assert.True(t, m.HasUID())
		assert.False(t, m.InvalidUID())
	})
	t.Run("False", func(t *testing.T) {
		m := FindUserByUID("uqxc08w3d0ej2283")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.False(t, m.SameUID("uqxc08w3d0ej2276"))
	})
	t.Run("Invalid", func(t *testing.T) {
		m := FindUserByUID("uqxc08w3d0ej2283")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.False(t, m.SameUID("xxx"))
	})
}

func TestUser_String(t *testing.T) {
	t.Run("UID", func(t *testing.T) {
		p := User{UserUID: "abc123", UserName: "", DisplayName: ""}
		assert.Equal(t, "abc123", p.String())
	})
	t.Run("FullName", func(t *testing.T) {
		p := User{UserUID: "abc123", UserName: "", DisplayName: "Test"}
		assert.Equal(t, "'Test'", p.String())
	})
	t.Run("UserName", func(t *testing.T) {
		p := User{UserUID: "abc123", UserName: "Super-User ", DisplayName: "Test"}
		assert.Equal(t, "'super-user'", p.String())
	})
}

func TestUser_Admin(t *testing.T) {
	t.Run("SuperAdmin", func(t *testing.T) {
		p := User{ID: 8, UserUID: "u000000000000008", UserName: "Hanna", DisplayName: "", SuperAdmin: true}
		assert.True(t, p.IsAdmin())
	})
	t.Run("RoleAdmin", func(t *testing.T) {
		p := User{ID: 8, UserUID: "u000000000000008", UserName: "Hanna", DisplayName: "", UserRole: acl.RoleAdmin.String()}
		assert.True(t, p.IsAdmin())
	})
	t.Run("NoID", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", UserName: "Hanna", DisplayName: "", UserRole: acl.RoleAdmin.String()}
		assert.False(t, p.IsAdmin())
	})
	t.Run("False", func(t *testing.T) {
		p := User{ID: 8, UserUID: "u000000000000008", UserName: "Hanna", DisplayName: "", SuperAdmin: false, UserRole: ""}
		assert.False(t, p.IsAdmin())
	})
}

func TestUser_IsUnknown(t *testing.T) {
	t.Run("ID", func(t *testing.T) {
		p := User{ID: UnknownUser.ID, UserUID: "u000000000000008", UserName: "", DisplayName: "", SuperAdmin: false, UserRole: acl.RoleAdmin.String()}
		assert.True(t, p.IsUnknown())
	})
	t.Run("UID", func(t *testing.T) {
		p := User{ID: 123, UserUID: "", UserName: "Hanna", DisplayName: "", SuperAdmin: false, UserRole: acl.RoleAdmin.String()}
		assert.True(t, p.IsUnknown())
	})
	t.Run("Name", func(t *testing.T) {
		p := User{ID: 123, UserUID: "u000000000000008", UserName: "", DisplayName: "", SuperAdmin: false, UserRole: acl.RoleAdmin.String()}
		assert.False(t, p.IsUnknown())
	})
	t.Run("Role", func(t *testing.T) {
		p := User{ID: 123, UserUID: "u000000000000008", UserName: "Hanna", DisplayName: "", SuperAdmin: false, UserRole: ""}
		assert.True(t, p.IsUnknown())
	})
	t.Run("Admin", func(t *testing.T) {
		p := User{ID: 123, UserUID: "u000000000000008", UserName: "Hanna", DisplayName: "", SuperAdmin: false, UserRole: acl.RoleAdmin.String()}
		assert.False(t, p.IsUnknown())
	})
	t.Run("SuperAdmin", func(t *testing.T) {
		p := User{ID: 123, UserUID: "u000000000000008", UserName: "Hanna", DisplayName: "", SuperAdmin: true, UserRole: ""}
		assert.False(t, p.IsUnknown())
	})
}

func TestUser_IsVisitor(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", UserName: "Hanna", DisplayName: "", UserRole: acl.RoleVisitor.String()}
		assert.True(t, p.IsVisitor())
	})
	t.Run("Unknown", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", UserName: "Hanna", DisplayName: ""}
		assert.False(t, p.IsVisitor())
	})
	t.Run("Admin", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", UserName: "Hanna", DisplayName: "", UserRole: acl.RoleAdmin.String()}
		assert.False(t, p.IsVisitor())
	})
}

func TestUser_SetPassword(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		m := User{ID: 8, UserUID: "u000000000000008", UserName: "Hanna", DisplayName: "", UserRole: acl.RoleAdmin.String()}
		assert.NoError(t, m.SetPassword("insecure"))
		assert.NoError(t, m.DeletePassword())
		assert.NoError(t, m.SetPassword("insecure"))
	})
	t.Run("NotRegistered", func(t *testing.T) {
		m := User{ID: 0, UserUID: "", UserName: "Hanna", DisplayName: "", UserRole: acl.RoleAdmin.String()}
		assert.Error(t, m.SetPassword("insecure"))
	})
	t.Run("PasswordTooShort", func(t *testing.T) {
		m := User{ID: 8, UserUID: "u000000000000008", UserName: "Hanna", DisplayName: "", UserRole: acl.RoleAdmin.String()}
		assert.Error(t, m.SetPassword("cat"))
	})
	t.Run("PasswordTooLong", func(t *testing.T) {
		m := User{ID: 8, UserUID: "u000000000000008", UserName: "Hanna", DisplayName: "", UserRole: acl.RoleAdmin.String()}
		assert.Error(t, m.SetPassword("hfnoehurhgfoeuro7584othgiyruifh85hglhiryhgbbyeirygbubgirgtheuogfugfkhsbdgiyerbgeuigbdtiyrgehbik"))
	})
}

func TestUser_InitLogin(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		p := User{UserUID: "u000000000000009", UserName: "Hanna", DisplayName: ""}
		assert.Nil(t, FindPassword("u000000000000009"))
		p.InitAccount("admin", "insecure")
		m := FindPassword("u000000000000009")

		if m == nil {
			t.Fatal("result should not be nil")
		}
	})
	t.Run("AlreadyExists", func(t *testing.T) {
		p := User{ID: 10, UserUID: "u000000000000010", UserName: "Hans", DisplayName: "", UserRole: acl.RoleAdmin.String()}

		if err := p.Save(); err != nil {
			t.Logf("failed to create user: %s", err)
		}

		if err := p.SetPassword("insecure"); err != nil {
			t.Fatal(err)
		}

		assert.NotNil(t, FindPassword("u000000000000010"))
		p.InitAccount("admin", "insecure")
		m := FindPassword("u000000000000010")

		if m == nil {
			t.Fatal("result should not be nil")
		}
	})
	t.Run("NotRegistered", func(t *testing.T) {
		p := User{UserUID: "u12", UserName: "", DisplayName: ""}
		assert.Nil(t, FindPassword("u12"))
		p.InitAccount("admin", "insecure")
		assert.Nil(t, FindPassword("u12"))
	})
	t.Run("EmptyPassword", func(t *testing.T) {
		p := User{UserUID: "u000000000000011", UserName: "User", DisplayName: ""}
		assert.Nil(t, FindPassword("u000000000000011"))
		p.InitAccount("admin", "")
		assert.Nil(t, FindPassword("u000000000000011"))
	})
}

func TestUser_AclRole(t *testing.T) {
	t.Run("SuperAdmin", func(t *testing.T) {
		p := User{ID: 8, UserUID: "u000000000000008", UserName: "Hanna", DisplayName: "", SuperAdmin: true, UserRole: ""}
		assert.Equal(t, acl.RoleAdmin, p.AclRole())
		assert.True(t, p.IsAdmin())
		assert.False(t, p.IsVisitor())
	})
	t.Run("RoleAdmin", func(t *testing.T) {
		p := User{ID: 8, UserUID: "u000000000000008", UserName: "Hanna", DisplayName: "", SuperAdmin: false, UserRole: acl.RoleAdmin.String()}
		assert.Equal(t, acl.RoleAdmin, p.AclRole())
		assert.True(t, p.IsAdmin())
		assert.False(t, p.IsVisitor())
	})
	t.Run("NoName", func(t *testing.T) {
		p := User{ID: 8, UserUID: "u000000000000008", UserName: "", DisplayName: "", UserRole: acl.RoleAdmin.String()}
		assert.Equal(t, acl.RoleVisitor, p.AclRole())
		assert.False(t, p.IsAdmin())
		assert.True(t, p.IsVisitor())
	})
	t.Run("NoID", func(t *testing.T) {
		p := User{ID: 0, UserUID: "u000000000000008", UserName: "Hanna", DisplayName: "", SuperAdmin: false, UserRole: acl.RoleAdmin.String()}
		assert.Equal(t, acl.RoleAdmin, p.AclRole())
		assert.False(t, p.IsAdmin())
		assert.False(t, p.IsVisitor())
	})
	t.Run("Unauthorized", func(t *testing.T) {
		p := User{ID: 8, UserUID: "u000000000000008", UserName: "Hanna", DisplayName: ""}
		assert.Equal(t, acl.RoleNone, p.AclRole())
		assert.False(t, p.IsAdmin())
		assert.False(t, p.IsVisitor())
	})
}

func TestUser_Validate(t *testing.T) {
	t.Run("NameValid", func(t *testing.T) {
		u := &User{
			UserName:    "validate",
			DisplayName: "Validate",
			UserEmail:   "validate@example.com",
			UserRole:    acl.RoleAdmin.String(),
		}

		assert.NoError(t, u.Validate())
	})
	t.Run("NameEmpty", func(t *testing.T) {
		u := &User{
			UserName:    "",
			DisplayName: "Validate",
			UserEmail:   "validate@example.com",
			UserRole:    acl.RoleAdmin.String(),
		}

		assert.Error(t, u.Validate())
	})
	t.Run("NameNotUnique", func(t *testing.T) {
		FirstOrCreateUser(&User{
			UserName: "notunique1",
		})

		u := &User{
			UserName:    "notunique1",
			DisplayName: "Not Unique",
			UserEmail:   "notunique1@example.com",
			UserRole:    acl.RoleAdmin.String(),
		}

		assert.Error(t, u.Validate())
	})
	t.Run("EmailNotUnique", func(t *testing.T) {
		FirstOrCreateUser(&User{
			UserEmail: "notunique2@example.com",
		})

		u := &User{
			UserName:    "notunique2",
			UserEmail:   "notunique2@example.com",
			DisplayName: "Not Unique",
			UserRole:    acl.RoleAdmin.String(),
		}

		assert.Error(t, u.Validate())
	})
	t.Run("EmailNotUnique", func(t *testing.T) {
		FirstOrCreateUser(&User{
			UserName:    "notunique3",
			UserEmail:   "notunique3@example.com",
			DisplayName: "Not Unique",
			UserRole:    acl.RoleAdmin.String(),
		})

		u := FirstOrCreateUser(&User{
			UserName:    "notunique30",
			UserEmail:   "notunique3@example.com",
			DisplayName: "Not Unique",
			UserRole:    acl.RoleAdmin.String(),
		})

		u.UserName = "notunique3"

		assert.Error(t, u.Validate())
	})
	t.Run("EmailInvalid", func(t *testing.T) {
		u := &User{
			UserName:    "12",
			DisplayName: "Validate",
			UserEmail:   "validate",
			UserRole:    acl.RoleAdmin.String(),
		}

		assert.Error(t, u.Validate())
	})
	t.Run("EmailInvalid2", func(t *testing.T) {
		u := &User{
			UserName:    "12",
			DisplayName: "Validate",
			UserEmail:   "validate@",
			UserRole:    acl.RoleAdmin.String(),
		}

		assert.Error(t, u.Validate())
	})
	t.Run("EmailEmpty", func(t *testing.T) {
		FirstOrCreateUser(&User{
			UserName: "nnomail",
		})

		u := &User{
			UserName:    "nomail",
			UserEmail:   "",
			DisplayName: "No Mail",
			UserRole:    acl.RoleAdmin.String(),
		}

		assert.NoError(t, u.Validate())
	})
	t.Run("RoleEmpty", func(t *testing.T) {
		u := &User{
			UserName:    "test.example",
			UserEmail:   "test@example.com",
			DisplayName: "Test Example",
			UserRole:    "",
		}

		assert.Error(t, u.Validate())
	})
	t.Run("RoleAdmin", func(t *testing.T) {
		u := &User{
			UserName:    "test.example",
			UserEmail:   "test@example.com",
			DisplayName: "Test Example",
			UserRole:    acl.RoleAdmin.String(),
		}

		assert.NoError(t, u.Validate())
	})
	t.Run("RoleInvalid", func(t *testing.T) {
		u := &User{
			UserName:    "test.example",
			UserEmail:   "test@example.com",
			DisplayName: "Test Example",
			UserRole:    "foobar",
		}

		assert.Error(t, u.Validate())
	})
}

func TestAddUser(t *testing.T) {
	t.Run("TooShort", func(t *testing.T) {
		u := form.User{
			UserName:  "thomas1",
			UserEmail: "thomas1@example.com",
			Password:  "hel",
		}

		err := AddUser(u)
		assert.Error(t, err)
	})
	t.Run("TooLong", func(t *testing.T) {
		u := form.User{
			UserName:  "thomas3",
			UserEmail: "thomas3@example.com",
			Password:  "1234567725244364789969hhkvnsgjlb;ghfnbn nd;dhewy8ortfgbkryeti7gfbie57yteoubgvlsiruwojflger",
			UserRole:  acl.RoleAdmin.String(),
		}

		err := AddUser(u)
		assert.Error(t, err)
	})
	t.Run("Invalid Role", func(t *testing.T) {
		u := form.User{
			UserName:  "thomas4",
			UserEmail: "thomas4@example.com",
			Password:  "helloworld",
			UserRole:  "invalid",
		}

		err := AddUser(u)
		assert.Error(t, err)
	})
	t.Run("Valid", func(t *testing.T) {
		u := form.User{
			UserName:  "thomas2",
			UserEmail: "thomas2@example.com",
			Password:  "helloworld",
			UserRole:  acl.RoleAdmin.String(),
		}

		err := AddUser(u)
		assert.Nil(t, err)
	})
}

func TestDeleteUser(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		u := &User{
			UserName:    "thomasdel",
			UserEmail:   "thomasdel@example.com",
			DisplayName: "Thomas Delete",
			UserRole:    acl.RoleAdmin.String(),
		}

		u = FirstOrCreateUser(u)
		err := u.Delete()
		assert.NoError(t, err)
	})
	t.Run("DoesNotExist", func(t *testing.T) {
		u := &User{
			UserName:    "thomasdel2",
			UserEmail:   "thomasdel2@example.com",
			DisplayName: "Thomas Delete 2",
			UserRole:    acl.RoleNone.String(),
		}

		err := u.Delete()
		assert.Error(t, err)
	})
	t.Run("Empty UID", func(t *testing.T) {
		u := NewUser()
		u.UserUID = ""
		u.ID = 500

		err := u.Delete()
		assert.Error(t, err)
	})
}

func TestUser_Deleted(t *testing.T) {
	assert.False(t, UserFixtures.Pointer("alice").IsDeleted())
	assert.True(t, UserFixtures.Pointer("deleted").IsDeleted())
}

func TestUser_Expired(t *testing.T) {
	t.Run("False", func(t *testing.T) {
		assert.False(t, UserFixtures.Pointer("alice").IsExpired())
		assert.False(t, UserFixtures.Pointer("deleted").IsExpired())
	})
	t.Run("True", func(t *testing.T) {
		u := NewUser()
		var expired = time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC)
		u.ExpiresAt = &expired

		assert.True(t, u.IsExpired())
	})
}

func TestUser_Disabled(t *testing.T) {
	assert.False(t, UserFixtures.Pointer("alice").IsDisabled())
	assert.True(t, UserFixtures.Pointer("deleted").IsDisabled())
}

func TestUser_UpdateLoginTime(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		alice := UserFixtures.Get("alice")
		time1 := alice.LoginAt
		assert.Nil(t, time1)
		alice.UpdateLoginTime()
		time2 := alice.LoginAt
		assert.NotNil(t, time2)
		alice.UpdateLoginTime()
		time3 := alice.LoginAt
		assert.NotNil(t, time3)
		assert.True(t, time3.After(*time2) || time3.Equal(*time2))
	})
	t.Run("User deleted", func(t *testing.T) {
		u := NewUser()
		var deleted = time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC)
		u.DeletedAt = &deleted
		assert.Nil(t, u.UpdateLoginTime())
	})
}

func TestUser_CanLogIn(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		alice := UserFixtures.Get("alice")
		assert.True(t, alice.CanLogIn())
		alice.SetProvider(authn.ProviderNone)
		assert.False(t, alice.CanLogIn())
		alice.SetProvider(authn.ProviderLocal)
		assert.True(t, alice.CanLogIn())

		assert.False(t, UserFixtures.Pointer("deleted").CanLogIn())
	})
	t.Run("False - !canlogin", func(t *testing.T) {
		u := NewUser()
		u.AuthProvider = "local"
		u.CanLogin = false
		assert.False(t, u.CanLogIn())
	})
	t.Run("False - unknown role", func(t *testing.T) {
		unknown := UserFixtures.Get("unauthorized")
		assert.False(t, unknown.CanLogIn())
		unknown.SetProvider(authn.ProviderLocal)
		assert.False(t, unknown.CanLogIn())
		unknown.SetProvider(authn.ProviderNone)
		assert.False(t, unknown.CanLogIn())
	})
}

func TestUser_CanUseWebDAV(t *testing.T) {
	alice := UserFixtures.Get("alice")
	assert.True(t, alice.CanUseWebDAV())
	alice.SetProvider(authn.ProviderNone)
	assert.False(t, alice.CanUseWebDAV())
	alice.SetProvider(authn.ProviderLocal)
	assert.True(t, alice.CanUseWebDAV())

	assert.False(t, UserFixtures.Pointer("deleted").CanUseWebDAV())
	assert.False(t, UserFixtures.Pointer("friend").CanUseWebDAV())
}

func TestUser_CanUpload(t *testing.T) {
	alice := UserFixtures.Get("alice")
	assert.True(t, alice.CanUpload())
	alice.SetProvider(authn.ProviderNone)
	assert.False(t, alice.CanUpload())
	alice.SetProvider(authn.ProviderLocal)
	assert.True(t, alice.CanUpload())

	assert.False(t, UserFixtures.Pointer("deleted").CanUpload())
	assert.True(t, UserFixtures.Pointer("friend").CanUpload())

	unknown := UserFixtures.Get("unauthorized")
	assert.False(t, unknown.CanUpload())
	unknown.SetProvider(authn.ProviderLocal)
	assert.False(t, unknown.CanUpload())
	unknown.SetProvider(authn.ProviderNone)
	assert.False(t, unknown.CanUpload())

}

func TestUser_SharedUIDs(t *testing.T) {
	t.Run("AliceAlbum", func(t *testing.T) {
		m := UserFixtures.Pointer("alice")
		assert.NotNil(t, m)

		result := m.SharedUIDs()
		assert.NotNil(t, result)
		assert.Len(t, result, 1)
		assert.Equal(t, UIDs{"as6sg6bxpogaaba9"}, result)
	})
}

func TestUser_Form(t *testing.T) {
	t.Run("Alice", func(t *testing.T) {
		m := FindUserByName("alice")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		frm, err := m.Form()

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("User: %#v", m)
		t.Logf("User.UserDetails: %#v", m.UserDetails)
		t.Logf("Form: %#v", frm)
		t.Logf("Form.UserDetails: %#v", frm.UserDetails)
	})
}

func TestUser_PrivilegeLevelChange(t *testing.T) {
	t.Run("True - Role changed", func(t *testing.T) {
		m := FindUserByName("alice")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		frm, err := m.Form()

		if err != nil {
			t.Fatal(err)
		}

		frm.UserRole = "guest"

		assert.True(t, m.PrivilegeLevelChange(frm))
	})
	t.Run("True - Path changed", func(t *testing.T) {
		m := FindUserByName("alice")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		frm, err := m.Form()

		if err != nil {
			t.Fatal(err)
		}

		frm.UploadPath = "uploads-alice"

		assert.True(t, m.PrivilegeLevelChange(frm))
	})
	t.Run("True - CanLogin changed", func(t *testing.T) {
		m := FindUserByName("alice")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		frm, err := m.Form()

		if err != nil {
			t.Fatal(err)
		}

		frm.CanLogin = false

		assert.True(t, m.PrivilegeLevelChange(frm))
	})
	t.Run("True - Webdav changed", func(t *testing.T) {
		m := FindUserByName("alice")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		frm, err := m.Form()

		if err != nil {
			t.Fatal(err)
		}

		frm.WebDAV = false

		assert.True(t, m.PrivilegeLevelChange(frm))
	})
	t.Run("False - Name changed", func(t *testing.T) {
		m := FindUserByName("alice")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		frm, err := m.Form()

		if err != nil {
			t.Fatal(err)
		}

		frm.UserName = "alice-new"

		assert.False(t, m.PrivilegeLevelChange(frm))
	})
	t.Run("False - No change", func(t *testing.T) {
		m := FindUserByName("alice")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		frm, err := m.Form()

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, m.PrivilegeLevelChange(frm))
	})
}

func TestUser_SaveForm(t *testing.T) {
	t.Run("UnknownUser", func(t *testing.T) {
		frm, err := UnknownUser.Form()
		assert.NoError(t, err)

		err = UnknownUser.SaveForm(frm, UserFixtures.Pointer("guest"))
		assert.Error(t, err)
	})
	t.Run("Admin", func(t *testing.T) {
		m := FindUser(Admin)

		if m == nil {
			t.Fatal("result should not be nil")
		}

		frm, err := m.Form()

		if err != nil {
			t.Fatal(err)
		}

		frm.UserEmail = "admin@example.com"
		frm.UserDetails.UserLocation = "GoLand"
		err = Admin.SaveForm(frm, UserFixtures.Pointer("guest"))

		assert.NoError(t, err)
		assert.Equal(t, "admin@example.com", Admin.UserEmail)
		assert.Equal(t, "GoLand", Admin.Details().UserLocation)

		m = FindUserByUID(Admin.UserUID)
		assert.Equal(t, "admin@example.com", m.UserEmail)
		assert.Equal(t, "GoLand", m.Details().UserLocation)
	})
	t.Run("UpdateRights", func(t *testing.T) {
		m := FindUser(Admin)

		if m == nil {
			t.Fatal("result should not be nil")
		}

		frm, err := m.Form()

		if err != nil {
			t.Fatal(err)
		}

		frm.UserEmail = "admin@example.com"
		frm.UserDetails.UserLocation = "GoLand"
		err = Admin.SaveForm(frm, UserFixtures.Pointer("alice"))

		assert.NoError(t, err)
		assert.Equal(t, "admin@example.com", Admin.UserEmail)
		assert.Equal(t, "GoLand", Admin.Details().UserLocation)

		m = FindUserByUID(Admin.UserUID)
		assert.Equal(t, "admin@example.com", m.UserEmail)
		assert.Equal(t, "GoLand", m.Details().UserLocation)
	})
	t.Run("Change display name", func(t *testing.T) {
		m := FindUser(Admin)

		if m == nil {
			t.Fatal("result should not be nil")
		}

		frm, err := m.Form()

		if err != nil {
			t.Fatal(err)
		}

		frm.DisplayName = "New Name"
		err = Admin.SaveForm(frm, UserFixtures.Pointer("alice"))

		assert.NoError(t, err)
		assert.Equal(t, "New Name", Admin.DisplayName)

		m = FindUserByUID(Admin.UserUID)
		assert.Equal(t, "New Name", m.DisplayName)
	})
	t.Run("Prevent log out initial admin", func(t *testing.T) {
		m := FindUser(Admin)

		if m == nil {
			t.Fatal("result should not be nil")
		}

		frm, err := m.Form()

		if err != nil {
			t.Fatal(err)
		}

		frm.CanLogin = false
		err = Admin.SaveForm(frm, UserFixtures.Pointer("alice"))

		assert.NoError(t, err)
		assert.Equal(t, true, Admin.CanLogin)

		m = FindUserByUID(Admin.UserUID)
		assert.Equal(t, true, m.CanLogin)
	})
	t.Run("AliceChangeAdminRights", func(t *testing.T) {
		user := UserFixtures.Get("alice")
		m := FindUser(user)

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, "admin", m.UserRole)
		assert.True(t, m.SuperAdmin)
		assert.True(t, m.CanLogin)

		frm, err := m.Form()

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "admin", frm.UserRole)
		assert.True(t, frm.SuperAdmin)
		assert.True(t, frm.CanLogin)

		frm.UserRole = "user"
		frm.SuperAdmin = false
		frm.CanLogin = false

		err = user.SaveForm(frm, &user)

		assert.NoError(t, err)
		assert.Equal(t, "admin", m.UserRole)
		assert.True(t, m.SuperAdmin)
		assert.True(t, m.CanLogin)
	})
	t.Run("Try to change role of superadmin", func(t *testing.T) {
		m := FindUser(Admin)

		if m == nil {
			t.Fatal("result should not be nil")
		}

		frm, err := m.Form()

		if err != nil {
			t.Fatal(err)
		}

		frm.UserRole = "user"
		err = Admin.SaveForm(frm, UserFixtures.Pointer("guest"))

		assert.Error(t, err)
		assert.Equal(t, "super admin must not have a non-admin role", err.Error())

		m = FindUserByUID(Admin.UserUID)
		assert.Equal(t, "admin", m.UserRole)
	})
	t.Run("Invalid base path", func(t *testing.T) {
		m := FindUser(Admin)

		if m == nil {
			t.Fatal("result should not be nil")
		}

		frm, err := m.Form()

		if err != nil {
			t.Fatal(err)
		}

		frm.BasePath = "//*?"
		err = Admin.SaveForm(frm, nil)

		assert.Error(t, err)
		assert.Equal(t, "invalid base folder", err.Error())

		m = FindUserByUID(Admin.UserUID)
		assert.Equal(t, "", m.BasePath)
	})
	t.Run("Invalid upload path", func(t *testing.T) {
		m := FindUser(Admin)

		if m == nil {
			t.Fatal("result should not be nil")
		}

		frm, err := m.Form()

		if err != nil {
			t.Fatal(err)
		}

		frm.UploadPath = "//*?"
		err = Admin.SaveForm(frm, nil)

		assert.Error(t, err)
		assert.Equal(t, "invalid upload folder", err.Error())

		m = FindUserByUID(Admin.UserUID)
		assert.Equal(t, "", m.UploadPath)
	})
}

func TestUser_SetDisplayName(t *testing.T) {
	t.Run("BillGates", func(t *testing.T) {
		user := NewUser()
		user.SetDisplayName("Sir William Henry Gates III", SrcAuto)
		d := user.Details()
		assert.Equal(t, "Sir", d.NameTitle)
		assert.Equal(t, "William", d.GivenName)
		assert.Equal(t, "Henry Gates", d.FamilyName)
		assert.Equal(t, "III", d.NameSuffix)
	})
	t.Run("Empty Name", func(t *testing.T) {
		user := NewUser()
		user.SetDisplayName("", SrcAuto)
		d := user.Details()
		assert.Equal(t, "", d.NameTitle)
		assert.Equal(t, "", d.GivenName)
		assert.Equal(t, "", d.FamilyName)
		assert.Equal(t, "", d.NameSuffix)
	})
	t.Run("priority too low", func(t *testing.T) {
		user := User{
			ID:      1234567,
			UserUID: "urqdrfb72479n047",
			UserDetails: &UserDetails{
				GivenName:  "John",
				MiddleName: "Percival",
				FamilyName: "Doe",
				NameSrc:    "manual",
			},
		}
		user.SetDisplayName("Sir William Henry Gates III", SrcAuto)
		d := user.Details()
		assert.Equal(t, "", d.NameTitle)
		assert.Equal(t, "John", d.GivenName)
		assert.Equal(t, "Doe", d.FamilyName)
		assert.Equal(t, "", d.NameSuffix)
	})
}

func TestUser_SetGivenName(t *testing.T) {
	user := User{
		ID:      1234567,
		UserUID: "urqdrfb72479n047",
		UserDetails: &UserDetails{
			GivenName:  "John",
			MiddleName: "Percival",
			FamilyName: "Doe",
		},
	}
	assert.Equal(t, "John", user.Details().GivenName)
	user.SetGivenName("Jane")
	assert.Equal(t, "Jane", user.Details().GivenName)
	user.SetGivenName("")
	assert.Equal(t, "Jane", user.Details().GivenName)
}

func TestUser_SetFamilyName(t *testing.T) {
	user := User{
		ID:      1234567,
		UserUID: "urqdrfb72479n047",
		UserDetails: &UserDetails{
			GivenName:  "John",
			MiddleName: "Percival",
			FamilyName: "Doe",
		},
	}
	assert.Equal(t, "Doe", user.Details().FamilyName)
	user.SetFamilyName("Maier")
	assert.Equal(t, "Maier", user.Details().FamilyName)
	user.SetFamilyName("")
	assert.Equal(t, "Maier", user.Details().FamilyName)
}

func TestUser_SetAvatar(t *testing.T) {
	t.Run("Visitor", func(t *testing.T) {
		assert.False(t, Visitor.HasAvatar())
		err := Visitor.SetAvatar("ebfc0aea7d3fd018b5fff57c76806b35181855ed", SrcManual)
		assert.Error(t, err)
	})
	t.Run("UnknownUser", func(t *testing.T) {
		assert.False(t, UnknownUser.HasAvatar())
		err := UnknownUser.SetAvatar("ebfc0aea7d3fd018b5fff57c76806b35181855ed", SrcManual)
		assert.Error(t, err)
	})
	t.Run("Admin", func(t *testing.T) {
		assert.False(t, Admin.HasAvatar())
		err := Admin.SetAvatar("ebfc0aea7d3fd018b5fff57c76806b35181855ed", SrcManual)
		assert.NoError(t, err)
		assert.Equal(t, "ebfc0aea7d3fd018b5fff57c76806b35181855ed", Admin.Thumb)
		assert.Equal(t, SrcManual, Admin.ThumbSrc)
		assert.True(t, Admin.HasAvatar())

		m := FindUserByUID(Admin.UserUID)
		assert.Equal(t, "ebfc0aea7d3fd018b5fff57c76806b35181855ed", m.Thumb)
		assert.Equal(t, SrcManual, m.ThumbSrc)
		assert.True(t, Admin.HasAvatar())
	})
	t.Run("No permissions", func(t *testing.T) {
		err := Admin.SetAvatar("ebfc0aea7d3fd018b5fff57c76806b35181855ed", SrcAuto)
		assert.Error(t, err)
	})
}

func TestUser_Username(t *testing.T) {
	t.Run("Visitor", func(t *testing.T) {
		assert.Equal(t, "", Visitor.Username())
	})
	t.Run("UnknownUser", func(t *testing.T) {
		assert.Equal(t, "", UnknownUser.Username())
	})
	t.Run("Admin", func(t *testing.T) {
		assert.Equal(t, "admin", Admin.Username())
	})
}

func TestUser_Provider(t *testing.T) {
	t.Run("Visitor", func(t *testing.T) {
		assert.Equal(t, authn.ProviderLink, Visitor.Provider())
		Visitor.AuthProvider = ""
		assert.Equal(t, authn.ProviderLink, Visitor.Provider())
		Visitor.SetProvider(authn.ProviderLink)
	})
	t.Run("UnknownUser", func(t *testing.T) {
		assert.Equal(t, authn.ProviderNone, UnknownUser.Provider())
	})
	t.Run("Admin", func(t *testing.T) {
		assert.Equal(t, authn.ProviderLocal, Admin.Provider())
		Admin.AuthProvider = ""
		assert.Equal(t, authn.ProviderLocal, Admin.Provider())
		Admin.SetProvider(authn.ProviderLocal)
	})
	t.Run("Username empty", func(t *testing.T) {
		user := NewUser()
		user.ID = 500
		user.AuthProvider = ""
		user.UserName = "test"
		assert.Equal(t, authn.ProviderDefault, user.Provider())
	})
}

func TestUser_SetProvider(t *testing.T) {
	t.Run("2FA", func(t *testing.T) {
		m := UserFixtures.Get("alice")
		assert.Equal(t, authn.ProviderLocal, m.Provider())
		assert.Equal(t, authn.MethodDefault, m.Method())
		m.SetMethod(authn.Method2FA)
		assert.Equal(t, authn.ProviderLocal, m.Provider())
		assert.Equal(t, authn.Method2FA, m.Method())
		m.SetProvider(authn.ProviderNone)
		assert.Equal(t, authn.ProviderNone, m.Provider())
		assert.Equal(t, authn.MethodUndefined, m.Method())
		m.SetProvider(authn.ProviderLocal)
	})
}

func TestUser_SetMethod(t *testing.T) {
	t.Run("2FA", func(t *testing.T) {
		m := UserFixtures.Get("unauthorized")
		assert.Equal(t, authn.ProviderNone, m.Provider())
		assert.Equal(t, authn.MethodDefault, m.Method())
		m.SetMethod(authn.Method2FA)
		assert.Equal(t, authn.ProviderNone, m.Provider())
		assert.Equal(t, authn.MethodDefault, m.Method())
	})
}

func TestUser_SetAuthID(t *testing.T) {
	id := rnd.UUID()
	issuer := "https://keycloak.localssl.dev/realms/master"

	t.Run("UUID", func(t *testing.T) {
		m := UserFixtures.Get("guest")

		m.SetAuthID(id, issuer)
		assert.Equal(t, id, m.AuthID)
		assert.Equal(t, issuer, m.AuthIssuer)
		m.SetAuthID(id, "")
		assert.Equal(t, id, m.AuthID)
		assert.Equal(t, "", m.AuthIssuer)
		m.SetAuthID("", issuer)
		assert.Equal(t, id, m.AuthID)
		assert.Equal(t, "", m.AuthIssuer)
	})
}

func TestUser_UpdateAuthID(t *testing.T) {
	id := rnd.UUID()
	issuer := "https://keycloak.localssl.dev/realms/master"

	t.Run("UUID", func(t *testing.T) {
		m := UserFixtures.Get("friend")

		m.SetAuthID("", issuer)
		assert.Equal(t, "", m.AuthID)
		assert.Equal(t, "", m.AuthIssuer)
		m.SetAuthID(id, issuer)
		assert.Equal(t, id, m.AuthID)
		assert.Equal(t, issuer, m.AuthIssuer)
		m.SetAuthID(id, "")
		assert.Equal(t, id, m.AuthID)
		assert.Equal(t, "", m.AuthIssuer)
	})
}

func TestUser_AuthInfo(t *testing.T) {
	t.Run("Alice", func(t *testing.T) {
		m := UserFixtures.Get("alice")
		assert.Equal(t, "Local", m.AuthInfo())
	})
	t.Run("Jane", func(t *testing.T) {
		m := UserFixtures.Get("jane")
		assert.Equal(t, "Local (2FA)", m.AuthInfo())
	})
	t.Run("Unauthorized", func(t *testing.T) {
		m := UserFixtures.Get("unauthorized")
		assert.Equal(t, "None", m.AuthInfo())
	})
}

func TestUser_Passcode(t *testing.T) {
	t.Run("Jane", func(t *testing.T) {
		m := UserFixtures.Get("jane")
		assert.IsType(t, &Passcode{}, m.Passcode("totp"))
	})
}

func TestUser_GetBasePath(t *testing.T) {
	t.Run("Visitor", func(t *testing.T) {
		assert.Equal(t, "", Visitor.GetBasePath())
	})
	t.Run("UnknownUser", func(t *testing.T) {
		assert.Equal(t, "", UnknownUser.GetBasePath())
	})
	t.Run("Admin", func(t *testing.T) {
		assert.Equal(t, "", Admin.GetBasePath())
	})
}

func TestUser_SetBasePath(t *testing.T) {
	t.Run("Test", func(t *testing.T) {
		u := User{
			ID:          1234567,
			UserUID:     "urqdrfb72479n047",
			UserName:    "test",
			UserRole:    acl.RoleAdmin.String(),
			DisplayName: "Test",
			SuperAdmin:  false,
			CanLogin:    true,
			WebDAV:      true,
			CanInvite:   false,
		}

		assert.Equal(t, "", u.SetBasePath("./").GetBasePath())
		assert.Equal(t, "base", u.SetBasePath("base").GetBasePath())
		assert.Equal(t, "users/test", u.SetBasePath("~").GetBasePath())
		assert.Equal(t, "users/test", u.DefaultBasePath())
		assert.Equal(t, "users/test", u.GetUploadPath())
	})
}

func TestUser_GetUploadPath(t *testing.T) {
	t.Run("Visitor", func(t *testing.T) {
		assert.Equal(t, "", Visitor.GetUploadPath())
	})
	t.Run("UnknownUser", func(t *testing.T) {
		assert.Equal(t, "", UnknownUser.GetUploadPath())
	})
	t.Run("Admin", func(t *testing.T) {
		assert.Equal(t, "", Admin.GetUploadPath())
	})
	t.Run("Use base path", func(t *testing.T) {
		user := NewUser()
		user.UserName = "test"
		user.UploadPath = "./"
		user.BasePath = "mybase"
		assert.Equal(t, "mybase", user.GetUploadPath())
	})
	t.Run("Upload path includes base path", func(t *testing.T) {
		user := NewUser()
		user.UserName = "test"
		user.UploadPath = "mybase/upload"
		user.BasePath = "mybase"
		assert.Equal(t, "mybase/upload", user.GetUploadPath())
	})
}

func TestUser_SetUploadPath(t *testing.T) {
	t.Run("Test", func(t *testing.T) {
		u := User{
			ID:          1234567,
			UserUID:     "urqdrfb72479n047",
			UserName:    "test",
			UserRole:    acl.RoleAdmin.String(),
			DisplayName: "Test",
			SuperAdmin:  false,
			CanLogin:    true,
			WebDAV:      true,
			CanInvite:   false,
		}

		assert.Equal(t, "upload", u.SetUploadPath("upload").GetUploadPath())
		assert.Equal(t, "", u.SetUploadPath("./").GetUploadPath())
		assert.Equal(t, "users/test", u.SetUploadPath("~").GetUploadPath())
		assert.Equal(t, "base/users/test", u.SetBasePath("base").GetUploadPath())
		assert.Equal(t, "base", u.SetUploadPath("~").GetUploadPath())
		assert.Equal(t, "users/test", u.DefaultBasePath())
		assert.Equal(t, "users/test", u.SetBasePath("~").GetUploadPath())
	})
}

func TestUser_Handle(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		u := User{
			ID:          1234567,
			UserUID:     "urqdrfb72479n047",
			UserName:    "mr-happy@cat.com",
			UserRole:    acl.RoleAdmin.String(),
			DisplayName: "",
			SuperAdmin:  false,
			CanLogin:    true,
			WebDAV:      true,
			CanInvite:   false,
		}

		assert.Equal(t, "mr-happy@cat.com", u.Username())
		assert.Equal(t, "mr-happy", u.Handle())

		u.UserName = "mr.happy@cat.com"

		assert.Equal(t, "mr.happy", u.Handle())

		u.UserName = "mr.happy"

		assert.Equal(t, "mr.happy", u.Handle())
	})
}

func TestUser_FullName(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		u := User{
			ID:          1234567,
			UserUID:     "urqdrfb72479n047",
			UserName:    "mr-happy@cat.com",
			UserRole:    acl.RoleAdmin.String(),
			DisplayName: "",
			SuperAdmin:  false,
			CanLogin:    true,
			WebDAV:      true,
			CanInvite:   false,
		}

		assert.Equal(t, "Mr-Happy", u.FullName())

		u.UserName = "mr.happy@cat.com"

		assert.Equal(t, "Mr Happy", u.FullName())

		u.UserName = "mr.happy"

		assert.Equal(t, "Mr Happy", u.FullName())

		u.UserName = "foo@bar.com"

		assert.Equal(t, "Foo", u.FullName())

		u.SetDisplayName("Jane Doe", SrcManual)

		assert.Equal(t, "Jane Doe", u.FullName())
	})
	t.Run("Name from Details", func(t *testing.T) {
		u := User{
			ID:          1234567,
			UserUID:     "urqdrfb72479n047",
			UserName:    "mr-happy@cat.com",
			UserRole:    acl.RoleAdmin.String(),
			DisplayName: "",
			SuperAdmin:  false,
			CanLogin:    true,
			WebDAV:      true,
			CanInvite:   false,
			UserDetails: &UserDetails{
				GivenName:  "John",
				FamilyName: "Doe",
			},
		}

		assert.Equal(t, "John Doe", u.FullName())

	})
	t.Run("Windows", func(t *testing.T) {
		u := User{
			ID:          1234567,
			UserUID:     "urqdrfb72479n047",
			UserName:    "DOMAIN\\Jens Mander",
			UserRole:    acl.RoleAdmin.String(),
			DisplayName: "",
			SuperAdmin:  false,
			CanLogin:    true,
			WebDAV:      true,
			CanInvite:   false,
		}

		assert.Equal(t, "jens.mander", u.Handle())
		assert.Equal(t, "domain\\jens mander", u.Username())
		assert.Equal(t, "Jens Mander", u.FullName())
	})
}

func TestUser_Settings(t *testing.T) {
	t.Run("return settings", func(t *testing.T) {
		u := User{
			ID:       1234567,
			UserUID:  "urqdrfb72479n047",
			UserName: "test",
			UserRole: "user",
			UserSettings: &UserSettings{
				UserUID:    "",
				UITheme:    "vanta",
				UILanguage: "de",
				MapsStyle:  "street",
				IndexPath:  "/photos",
			},
		}

		assert.Equal(t, "de", u.Settings().UILanguage)
		assert.Equal(t, "vanta", u.Settings().UITheme)
	})
	t.Run("empty uid", func(t *testing.T) {
		u := User{
			ID:       1234567,
			UserUID:  "",
			UserName: "test",
			UserRole: "user",
		}

		assert.Equal(t, "", u.Settings().UILanguage)
		assert.Equal(t, "", u.Settings().UITheme)
	})
}

func TestUser_Details(t *testing.T) {
	t.Run("return details", func(t *testing.T) {
		u := User{
			ID:       1234567,
			UserUID:  "urqdrfb72479n047",
			UserName: "test",
			UserRole: "user",
			UserDetails: &UserDetails{
				GivenName:  "John",
				FamilyName: "Doe",
			},
		}

		assert.Equal(t, "John", u.Details().GivenName)
		assert.Equal(t, "Doe", u.Details().FamilyName)
	})
	t.Run("empty uid", func(t *testing.T) {
		u := User{
			ID:       1234567,
			UserUID:  "",
			UserName: "test",
			UserRole: "user",
		}

		assert.Equal(t, "", u.Details().GivenName)
		assert.Equal(t, "", u.Details().FamilyName)
	})
}

func TestUser_Equal(t *testing.T) {
	assert.True(t, Admin.Equal(&Admin))
	assert.False(t, Admin.Equal(&Visitor))
}

func TestUser_DeleteSessions(t *testing.T) {
	t.Run("empty uid", func(t *testing.T) {
		u := User{
			ID:       1234567,
			UserUID:  "",
			UserName: "test",
			UserRole: "user",
		}

		assert.Equal(t, 0, u.DeleteSessions([]string{}))
	})
	t.Run("alice", func(t *testing.T) {
		m := FindLocalUser("alice")

		assert.Equal(t, 0, m.DeleteSessions([]string{rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0")}))
		assert.Equal(t, 1, m.DeleteSessions([]string{}))
	})
}

func TestUser_VerifyPassword(t *testing.T) {
	assert.True(t, Admin.VerifyPassword("photoprism"))
	assert.False(t, Admin.VerifyPassword("wrong"))
}

func TestUser_InvalidPasscode(t *testing.T) {
	m := UserFixtures.Get("jane")
	passcode := m.Passcode("totp")

	assert.True(t, m.InvalidPasscode("xxxxxx"))
	assert.False(t, m.InvalidPasscode(passcode.RecoveryCode))
}

func TestUser_Passcodes(t *testing.T) {
	t.Run("Alice", func(t *testing.T) {
		m := UserFixtures.Get("alice")
		assert.Equal(t, "default", m.AuthMethod)

		passcode, err := m.ActivatePasscode()

		assert.Error(t, err)

		assert.Equal(t, "default", m.AuthMethod)

		code, err := passcode.GenerateCode()

		if err != nil {
			t.Fatal(err)
		}

		_, _, errInvalid := m.VerifyPasscode("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
		assert.Error(t, errInvalid)
		assert.Equal(t, authn.ErrInvalidPasscode, errInvalid)

		v, _, err := m.VerifyPasscode(code)

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, v)

		_, errActivate := m.ActivatePasscode()

		if errActivate != nil {
			t.Fatal(errActivate)
		}

		assert.Equal(t, "2fa", m.AuthMethod)

		_, errDeactivate := m.DeactivatePasscode()

		if errDeactivate != nil {
			t.Fatal(errDeactivate)
		}

		assert.Equal(t, "default", m.AuthMethod)

	})
	t.Run("PassCodeNotSetup", func(t *testing.T) {
		m := UserFixtures.Get("bob")
		assert.Equal(t, "default", m.AuthMethod)

		passcode, err := m.ActivatePasscode()

		assert.Error(t, err)
		assert.Equal(t, authn.ErrPasscodeNotSetUp, err)

		code, err := passcode.GenerateCode()

		assert.Error(t, err)

		v, _, err := m.VerifyPasscode(code)
		assert.False(t, v)
		assert.Error(t, err)
		assert.Equal(t, authn.ErrPasscodeNotSetUp, err)

		assert.Equal(t, "default", m.AuthMethod)

		code2, err := m.DeactivatePasscode()

		if err != nil {
			t.Fatal(err)
		}

		assert.Nil(t, code2)

	})
	t.Run("PassCodeNotSupported", func(t *testing.T) {
		m := UserFixtures.Get("unauthorized")
		_, err := m.ActivatePasscode()

		assert.Error(t, err)
		assert.Equal(t, authn.ErrPasscodeNotSupported, err)

	})
}

func TestUser_RegenerateTokens(t *testing.T) {
	t.Run("Visitor", func(t *testing.T) {
		assert.Nil(t, Visitor.RegenerateTokens())
	})
	t.Run("Admin", func(t *testing.T) {
		preview := Admin.PreviewToken
		download := Admin.DownloadToken

		err := Admin.RegenerateTokens()

		if err != nil {
			t.Fatal(err)
		}

		assert.NotEqual(t, preview, Admin.PreviewToken)
		assert.NotEqual(t, download, Admin.DownloadToken)
	})
}

func TestUser_HasShares(t *testing.T) {
	t.Run("Visitor", func(t *testing.T) {
		assert.False(t, Visitor.HasShares())
	})
	t.Run("Alice", func(t *testing.T) {
		m := FindLocalUser("alice")
		assert.True(t, m.HasShares())
	})
}

func TestUser_HasShare(t *testing.T) {
	m := FindLocalUser("alice")
	m.RefreshShares()
	assert.True(t, m.HasShare("as6sg6bxpogaaba9"))
	assert.False(t, m.HasShare("as6sg6bxpogaaba8"))
	assert.False(t, Visitor.HasShare("as6sg6bxpogaaba8"))
}

func TestUser_RedeemToken(t *testing.T) {
	t.Run("Visitor", func(t *testing.T) {
		assert.Equal(t, 0, Visitor.RedeemToken("1234"))
	})
	t.Run("Alice", func(t *testing.T) {
		m := FindLocalUser("alice")
		m.RefreshShares()
		assert.Equal(t, "as6sg6bxpogaaba9", m.UserShares[0].ShareUID)
		assert.Equal(t, 0, m.RedeemToken("1234"))
		m.RefreshShares()
		assert.Equal(t, "as6sg6bxpogaaba9", m.UserShares[0].ShareUID)
		assert.Equal(t, 1, m.RedeemToken("4jxf3jfn2k"))
		m.RefreshShares()
		assert.Equal(t, "as6sg6bxpogaaba7", m.UserShares[0].ShareUID)
		assert.Equal(t, "as6sg6bxpogaaba9", m.UserShares[1].ShareUID)
	})
}
