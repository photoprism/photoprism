package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestNewUser(t *testing.T) {
	m := NewUser()

	assert.True(t, rnd.IsRefID(m.RefID))
	assert.True(t, rnd.IsUID(m.UserUID, UserUID))
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
		assert.Equal(t, "admin", m.Name())
		m.UserName = "Admin "
		assert.Equal(t, "admin", m.Name())
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

		assert.Equal(t, "example", m.Name())
		assert.Equal(t, "example", m.UserName)

		if err := m.UpdateName("example-editor"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "example-editor", m.Name())
		assert.Equal(t, "example-editor", m.UserName)
	})
	t.Run("NewUser", func(t *testing.T) {
		if err := NewUser().Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestUser_SetName(t *testing.T) {
	t.Run("photoprism", func(t *testing.T) {
		m := FindUserByName("admin")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, "admin", m.Name())
		assert.Equal(t, "admin", m.UserName)

		if err := m.SetName("photoprism"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "photoprism", m.Name())
		assert.Equal(t, "photoprism", m.UserName)
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
	t.Run("Success", func(t *testing.T) {
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
		assert.Equal(t, "alice", m.Name())
		assert.Equal(t, "Alice", m.DisplayName)
		assert.Equal(t, "alice@example.com", m.UserEmail)
		assert.True(t, m.SuperAdmin)
		assert.True(t, m.IsAdmin())
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
	})

}

func TestUser_String(t *testing.T) {
	t.Run("UID", func(t *testing.T) {
		p := User{UserUID: "abc123", UserName: "", DisplayName: ""}
		assert.Equal(t, "abc123", p.String())
	})
	t.Run("FullName", func(t *testing.T) {
		p := User{UserUID: "abc123", UserName: "", DisplayName: "Test"}
		assert.Equal(t, "Test", p.String())
	})
	t.Run("UserName", func(t *testing.T) {
		p := User{UserUID: "abc123", UserName: "Super-User ", DisplayName: "Test"}
		assert.Equal(t, "super-user", p.String())
	})
}

func TestUser_Admin(t *testing.T) {
	t.Run("SuperAdmin", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", UserName: "Hanna", DisplayName: "", SuperAdmin: true}
		assert.True(t, p.IsAdmin())
	})
	t.Run("RoleAdmin", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", UserName: "Hanna", DisplayName: "", UserRole: acl.RoleAdmin.String()}
		assert.True(t, p.IsAdmin())
	})
	t.Run("False", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", UserName: "Hanna", DisplayName: "", SuperAdmin: false, UserRole: ""}
		assert.False(t, p.IsAdmin())
	})
}

func TestUser_Anonymous(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		p := User{UserUID: "", UserName: "Hanna", DisplayName: ""}
		assert.True(t, p.IsUnknown())
	})
	t.Run("False", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", UserName: "Hanna", DisplayName: "", SuperAdmin: true}
		assert.False(t, p.IsUnknown())
	})
}

func TestUser_Guest(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		p := User{UserUID: "", UserName: "Hanna", DisplayName: "", UserRole: acl.RoleVisitor.String()}
		assert.True(t, p.IsVisitor())
	})
	t.Run("False", func(t *testing.T) {
		p := User{UserUID: "", UserName: "Hanna", DisplayName: ""}
		assert.False(t, p.IsVisitor())
	})
}

func TestUser_SetPassword(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", UserName: "Hanna", DisplayName: ""}
		if err := p.SetPassword("insecure"); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("NotRegistered", func(t *testing.T) {
		p := User{UserUID: "", UserName: "Hanna", DisplayName: ""}
		assert.Error(t, p.SetPassword("insecure"))

	})
	t.Run("PasswordTooShort", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", UserName: "Hanna", DisplayName: ""}
		assert.Error(t, p.SetPassword("cat"))
	})
}

func TestUser_InitLogin(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		p := User{UserUID: "u000000000000009", UserName: "Hanna", DisplayName: ""}
		assert.Nil(t, FindPassword("u000000000000009"))
		p.InitAccount("admin", "insecure")
		m := FindPassword("u000000000000009")

		if m == nil {
			t.Fatal("result should not be nil")
		}
	})
	t.Run("AlreadyExists", func(t *testing.T) {
		p := User{UserUID: "u000000000000010", UserName: "Hans", DisplayName: ""}

		if err := p.Save(); err != nil {
			t.Logf("cannot user %s: ", err)
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
		p := User{UserUID: "u000000000000008", UserName: "Hanna", DisplayName: "", SuperAdmin: true, UserRole: ""}
		assert.Equal(t, acl.RoleAdmin, p.AclRole())
		assert.True(t, p.IsAdmin())
		assert.False(t, p.IsVisitor())
	})
	t.Run("RoleAdmin", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", UserName: "Hanna", DisplayName: "", SuperAdmin: false, UserRole: acl.RoleAdmin.String()}
		assert.Equal(t, acl.RoleAdmin, p.AclRole())
		assert.True(t, p.IsAdmin())
		assert.False(t, p.IsVisitor())
	})
	t.Run("NoName", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", UserName: "", DisplayName: "", UserRole: acl.RoleAdmin.String()}
		assert.Equal(t, acl.RoleVisitor, p.AclRole())
		assert.False(t, p.IsAdmin())
		assert.True(t, p.IsVisitor())
	})
	t.Run("Unauthorized", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", UserName: "Hanna", DisplayName: ""}
		assert.Equal(t, acl.RoleUnauthorized, p.AclRole())
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
	t.Run("NameTooShort", func(t *testing.T) {
		u := &User{
			UserName:    "va",
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
			UserName:    "jens.mander",
			UserEmail:   "jens@mander.de",
			DisplayName: "Jens Mander",
			UserRole:    "",
		}

		assert.Error(t, u.Validate())
	})
	t.Run("RoleAdmin", func(t *testing.T) {
		u := &User{
			UserName:    "jens.mander",
			UserEmail:   "jens@mander.de",
			DisplayName: "Jens Mander",
			UserRole:    acl.RoleAdmin.String(),
		}

		assert.NoError(t, u.Validate())
	})
	t.Run("RoleInvalid", func(t *testing.T) {
		u := &User{
			UserName:    "jens.mander",
			UserEmail:   "jens@mander.de",
			DisplayName: "Jens Mander",
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
	t.Run("Success", func(t *testing.T) {
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
			UserRole:    acl.RoleUnauthorized.String(),
		}

		err := u.Delete()
		assert.Error(t, err)
	})
}

func TestUser_Deleted(t *testing.T) {
	assert.False(t, UserFixtures.Pointer("alice").Deleted())
	assert.True(t, UserFixtures.Pointer("deleted").Deleted())
}

func TestUser_Expired(t *testing.T) {
	assert.False(t, UserFixtures.Pointer("alice").Expired())
	assert.False(t, UserFixtures.Pointer("deleted").Expired())
}

func TestUser_Disabled(t *testing.T) {
	assert.False(t, UserFixtures.Pointer("alice").Disabled())
	assert.True(t, UserFixtures.Pointer("deleted").Disabled())
}

func TestUser_LoginAllowed(t *testing.T) {
	assert.True(t, UserFixtures.Pointer("alice").LoginAllowed())
	assert.False(t, UserFixtures.Pointer("deleted").LoginAllowed())
}

func TestUser_SyncAllowed(t *testing.T) {
	assert.True(t, UserFixtures.Pointer("alice").SyncAllowed())
	assert.False(t, UserFixtures.Pointer("deleted").SyncAllowed())
	assert.False(t, UserFixtures.Pointer("friend").SyncAllowed())
}

func TestUser_UploadAllowed(t *testing.T) {
	assert.True(t, UserFixtures.Pointer("alice").UploadAllowed())
	assert.False(t, UserFixtures.Pointer("deleted").UploadAllowed())
	assert.True(t, UserFixtures.Pointer("friend").UploadAllowed())
}

func TestUser_SharedUIDs(t *testing.T) {
	t.Run("AliceAlbum", func(t *testing.T) {
		m := UserFixtures.Pointer("alice")
		assert.NotNil(t, m)

		result := m.SharedUIDs()
		assert.NotNil(t, result)
		assert.Len(t, result, 1)
		assert.Equal(t, UIDs{"at9lxuqxpogaaba9"}, result)
	})
}
