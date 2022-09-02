package entity

import (
	"testing"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/form"

	"github.com/stretchr/testify/assert"
)

func TestFindUserByName(t *testing.T) {
	t.Run("admin", func(t *testing.T) {
		m := FindUserByLogin("admin")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, 1, m.ID)
		assert.NotEmpty(t, m.UserUID)
		assert.Equal(t, "admin", m.Username)
		assert.Equal(t, "admin", m.UserName())
		m.Username = "Admin "
		assert.Equal(t, "admin", m.UserName())
		assert.Equal(t, "Admin ", m.Username)
		assert.Equal(t, "Admin", m.DisplayName)
		assert.Equal(t, acl.RoleAdmin, m.AclRole())
		assert.False(t, m.IsEditor())
		assert.False(t, m.IsViewer())
		assert.False(t, m.IsGuest())
		assert.True(t, m.SuperAdmin)
		assert.True(t, m.CanLogin)
		assert.True(t, m.CanInvite)
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})

	t.Run("alice", func(t *testing.T) {
		m := FindUserByLogin("alice")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, 5, m.ID)
		assert.Equal(t, "uqxetse3cy5eo9z2", m.UserUID)
		assert.Equal(t, "alice", m.Username)
		assert.Equal(t, "Alice", m.DisplayName)
		assert.Equal(t, "alice@example.com", m.Email)
		assert.True(t, m.SuperAdmin)
		assert.Equal(t, acl.RoleAdmin, m.AclRole())
		assert.NotEqual(t, acl.RoleGuest, m.AclRole())
		assert.False(t, m.IsEditor())
		assert.False(t, m.IsViewer())
		assert.False(t, m.IsGuest())
		assert.True(t, m.CanLogin)
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})

	t.Run("bob", func(t *testing.T) {
		m := FindUserByLogin("bob")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, 7, m.ID)
		assert.Equal(t, "uqxc08w3d0ej2283", m.UserUID)
		assert.Equal(t, "bob", m.Username)
		assert.Equal(t, "Bob", m.DisplayName)
		assert.Equal(t, "bob@example.com", m.Email)
		assert.False(t, m.SuperAdmin)
		assert.True(t, m.IsEditor())
		assert.False(t, m.IsViewer())
		assert.False(t, m.IsGuest())
		assert.True(t, m.CanLogin)
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})

	t.Run("unknown", func(t *testing.T) {
		m := FindUserByLogin("")

		if m != nil {
			t.Fatal("result should be nil")
		}
	})

	t.Run("not found", func(t *testing.T) {
		m := FindUserByLogin("xxx")

		if m != nil {
			t.Fatal("result should be nil")
		}
	})
}

func TestUser_Create(t *testing.T) {
	t.Run("Slug", func(t *testing.T) {
		var m = User{
			Username:    "example",
			UserRole:    acl.RoleEditor.String(),
			DisplayName: "Example",
			SuperAdmin:  false,
			CanLogin:    true,
		}

		if err := m.Create(); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "example", m.UserName())
		assert.Equal(t, "example", m.Username)
		assert.Equal(t, m.Username, m.UserSlug)

		if err := m.UpdateName("example-editor"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "example-editor", m.UserName())
		assert.Equal(t, "example-editor", m.Username)
		assert.Equal(t, m.Username, m.UserSlug)
	})
}

func TestUser_SetName(t *testing.T) {
	t.Run("photoprism", func(t *testing.T) {
		m := FindUserByLogin("admin")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, "admin", m.UserName())
		assert.Equal(t, "admin", m.Username)

		m.SetUsername("photoprism")

		assert.Equal(t, "photoprism", m.UserName())
		assert.Equal(t, "photoprism", m.Username)
	})
}

func TestUser_InvalidPassword(t *testing.T) {
	t.Run("admin", func(t *testing.T) {
		m := FindUserByLogin("admin")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.False(t, m.InvalidPassword("photoprism"))
	})
	t.Run("admin invalid password", func(t *testing.T) {
		m := FindUserByLogin("admin")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.True(t, m.InvalidPassword("wrong-password"))
	})
	t.Run("no password existing", func(t *testing.T) {
		p := User{UserUID: "u000000000000010", Username: "Hans", DisplayName: ""}
		err := p.Save()
		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, p.InvalidPassword("abcdef"))

	})
	t.Run("not registered", func(t *testing.T) {
		p := User{UserUID: "u12", Username: "", DisplayName: ""}
		assert.True(t, p.InvalidPassword("abcdef"))
	})
	t.Run("password empty", func(t *testing.T) {
		p := User{UserUID: "u000000000000011", Username: "User", DisplayName: ""}
		assert.True(t, p.InvalidPassword(""))
	})
}

func TestUser_Save(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		p := User{}

		err := p.Save()

		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestFirstOrCreateUser(t *testing.T) {
	t.Run("not existing", func(t *testing.T) {
		p := &User{ID: 555}

		result := FirstOrCreateUser(p)
		if result == nil {
			t.Fatal("result should not be nil")
		}

		assert.NotEmpty(t, result.ID)

	})
	t.Run("existing", func(t *testing.T) {
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
	t.Run("guest", func(t *testing.T) {
		m := FindUserByUID("u000000000000002")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, -2, m.ID)
		assert.NotEmpty(t, m.UserUID)
		assert.Equal(t, "", m.Username)
		assert.Equal(t, "Guest", m.DisplayName)
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})

	t.Run("unknown", func(t *testing.T) {
		m := FindUserByUID("")

		if m != nil {
			t.Fatal("result should be nil")
		}
	})

	t.Run("not found", func(t *testing.T) {
		m := FindUserByUID("xxx")

		if m != nil {
			t.Fatal("result should be nil")
		}
	})

	t.Run("alice", func(t *testing.T) {
		m := FindUserByUID("uqxetse3cy5eo9z2")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, 5, m.ID)
		assert.Equal(t, "uqxetse3cy5eo9z2", m.UserUID)
		assert.Equal(t, "alice", m.UserName())
		assert.Equal(t, "Alice", m.DisplayName)
		assert.Equal(t, "alice@example.com", m.Email)
		assert.True(t, m.SuperAdmin)
		assert.True(t, m.IsAdmin())
		assert.False(t, m.IsEditor())
		assert.False(t, m.IsViewer())
		assert.False(t, m.IsGuest())
		assert.True(t, m.CanLogin)
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})

	t.Run("bob", func(t *testing.T) {
		m := FindUserByUID("uqxc08w3d0ej2283")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, 7, m.ID)
		assert.Equal(t, "uqxc08w3d0ej2283", m.UserUID)
		assert.Equal(t, "bob", m.Username)
		assert.Equal(t, "Bob", m.DisplayName)
		assert.Equal(t, "bob@example.com", m.Email)
		assert.False(t, m.SuperAdmin)
		assert.False(t, m.IsAdmin())
		assert.True(t, m.IsEditor())
		assert.False(t, m.IsViewer())
		assert.False(t, m.IsGuest())
		assert.True(t, m.CanLogin)
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})
	t.Run("friend", func(t *testing.T) {
		m := FindUserByUID("uqxqg7i1kperxvu7")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, 8, m.ID)
		assert.Equal(t, "uqxqg7i1kperxvu7", m.UserUID)
		assert.False(t, m.SuperAdmin)
		assert.False(t, m.IsAdmin())
		assert.False(t, m.IsEditor())
		assert.True(t, m.IsViewer())
		assert.False(t, m.IsGuest())
		assert.True(t, m.CanLogin)
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})

}

func TestUser_String(t *testing.T) {
	t.Run("UID", func(t *testing.T) {
		p := User{UserUID: "abc123", Username: "", DisplayName: ""}
		assert.Equal(t, "abc123", p.String())
	})
	t.Run("FullName", func(t *testing.T) {
		p := User{UserUID: "abc123", Username: "", DisplayName: "Test"}
		assert.Equal(t, "Test", p.String())
	})
	t.Run("Username", func(t *testing.T) {
		p := User{UserUID: "abc123", Username: "Super-User ", DisplayName: "Test"}
		assert.Equal(t, "super-user", p.String())
	})
}

func TestUser_Admin(t *testing.T) {
	t.Run("SuperAdmin", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", Username: "Hanna", DisplayName: "", SuperAdmin: true}
		assert.True(t, p.IsAdmin())
	})
	t.Run("RoleAdmin", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", Username: "Hanna", DisplayName: "", UserRole: acl.RoleAdmin.String()}
		assert.True(t, p.IsAdmin())
	})
	t.Run("false", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", Username: "Hanna", DisplayName: "", SuperAdmin: false, UserRole: ""}
		assert.False(t, p.IsAdmin())
	})
}

func TestUser_Anonymous(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		p := User{UserUID: "", Username: "Hanna", DisplayName: ""}
		assert.True(t, p.IsAnonymous())
	})
	t.Run("false", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", Username: "Hanna", DisplayName: "", SuperAdmin: true}
		assert.False(t, p.IsAnonymous())
	})
}

func TestUser_Guest(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		p := User{UserUID: "", Username: "Hanna", DisplayName: "", UserRole: "guest"}
		assert.True(t, p.IsGuest())
	})
	t.Run("false", func(t *testing.T) {
		p := User{UserUID: "", Username: "Hanna", DisplayName: ""}
		assert.False(t, p.IsGuest())
	})
}

func TestUser_SetPassword(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", Username: "Hanna", DisplayName: ""}
		if err := p.SetPassword("insecure"); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("not registered", func(t *testing.T) {
		p := User{UserUID: "", Username: "Hanna", DisplayName: ""}
		assert.Error(t, p.SetPassword("insecure"))

	})
	t.Run("password too short", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", Username: "Hanna", DisplayName: ""}
		assert.Error(t, p.SetPassword("cat"))
	})
}

func TestUser_InitLogin(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		p := User{UserUID: "u000000000000009", Username: "Hanna", DisplayName: ""}
		assert.Nil(t, FindPassword("u000000000000009"))
		p.InitAccount("admin", "insecure")
		m := FindPassword("u000000000000009")

		if m == nil {
			t.Fatal("result should not be nil")
		}
	})
	t.Run("already existing", func(t *testing.T) {
		p := User{UserUID: "u000000000000010", Username: "Hans", DisplayName: ""}

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
	t.Run("not registered", func(t *testing.T) {
		p := User{UserUID: "u12", Username: "", DisplayName: ""}
		assert.Nil(t, FindPassword("u12"))
		p.InitAccount("admin", "insecure")
		assert.Nil(t, FindPassword("u12"))
	})
	t.Run("password empty", func(t *testing.T) {
		p := User{UserUID: "u000000000000011", Username: "User", DisplayName: ""}
		assert.Nil(t, FindPassword("u000000000000011"))
		p.InitAccount("admin", "")
		assert.Nil(t, FindPassword("u000000000000011"))
	})
}

func TestUser_Role(t *testing.T) {
	t.Run("admin", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", Username: "Hanna", DisplayName: "", SuperAdmin: true, UserRole: acl.RoleAdmin.String()}
		assert.Equal(t, acl.Role("admin"), p.AclRole())
		assert.True(t, p.IsAdmin())
		assert.False(t, p.IsGuest())
	})
	t.Run("guest", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", Username: "Hanna", DisplayName: "", UserRole: acl.RoleGuest.String()}
		assert.Equal(t, acl.Role("guest"), p.AclRole())
		assert.False(t, p.IsAdmin())
		assert.True(t, p.IsGuest())
	})
	t.Run("default", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", Username: "Hanna", DisplayName: ""}
		assert.Equal(t, acl.Role("*"), p.AclRole())
		assert.False(t, p.IsAdmin())
		assert.False(t, p.IsGuest())
	})
}

func TestUser_Validate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		u := &User{
			Username:    "validate",
			DisplayName: "Validate",
			Email:       "validate@example.com",
		}
		err := u.Validate()
		assert.Nil(t, err)
	})
	t.Run("username empty", func(t *testing.T) {
		u := &User{
			Username:    "",
			DisplayName: "Validate",
			Email:       "validate@example.com",
		}
		err := u.Validate()
		assert.Error(t, err)
	})
	t.Run("username too short", func(t *testing.T) {
		u := &User{
			Username:    "va",
			DisplayName: "Validate",
			Email:       "validate@example.com",
		}
		err := u.Validate()
		assert.Error(t, err)
	})
	t.Run("username not unique", func(t *testing.T) {
		FirstOrCreateUser(&User{
			Username: "notunique1",
		})
		u := &User{
			Username:    "notunique1",
			DisplayName: "Not Unique",
			Email:       "notunique1@example.com",
		}
		assert.Error(t, u.Validate())
	})
	t.Run("email not unique", func(t *testing.T) {
		FirstOrCreateUser(&User{
			Email: "notunique2@example.com",
		})
		u := &User{
			Username:    "notunique2",
			Email:       "notunique2@example.com",
			DisplayName: "Not Unique",
		}
		assert.Error(t, u.Validate())
	})
	t.Run("update user - email not unique", func(t *testing.T) {
		FirstOrCreateUser(&User{
			Username:    "notunique3",
			Email:       "notunique3@example.com",
			DisplayName: "Not Unique",
		})
		u := FirstOrCreateUser(&User{
			Username:    "notunique30",
			Email:       "notunique3@example.com",
			DisplayName: "Not Unique",
		})
		u.Username = "notunique3"
		assert.Error(t, u.Validate())
	})
	t.Run("primary email empty", func(t *testing.T) {
		FirstOrCreateUser(&User{
			Username: "nnomail",
		})
		u := &User{
			Username:    "nomail",
			Email:       "",
			DisplayName: "No Mail",
		}
		assert.Nil(t, u.Validate())
	})
}

func TestCreateWithPassword(t *testing.T) {
	t.Run("password too short", func(t *testing.T) {
		u := form.UserCreate{
			Username: "thomas1",
			Email:    "thomas1@example.com",
			Password: "hel",
		}

		err := CreateWithPassword(u)
		assert.Error(t, err)
	})
	t.Run("valid", func(t *testing.T) {
		u := form.UserCreate{
			Username: "thomas2",
			Email:    "thomas2@example.com",
			Password: "helloworld",
		}

		err := CreateWithPassword(u)
		assert.Nil(t, err)
	})
}

func TestDeleteUser(t *testing.T) {
	t.Run("delete user - success", func(t *testing.T) {
		u := &User{
			Username:    "thomasdel",
			Email:       "thomasdel@example.com",
			DisplayName: "Thomas Delete",
		}

		u = FirstOrCreateUser(u)
		err := u.Delete()
		assert.NoError(t, err)
	})
	t.Run("delete user - not in db", func(t *testing.T) {
		u := &User{
			Username:    "thomasdel2",
			Email:       "thomasdel2@example.com",
			DisplayName: "Thomas Delete 2",
		}

		err := u.Delete()
		assert.Error(t, err)
	})
}

func TestUser_Deleted(t *testing.T) {
	assert.False(t, UserFixtures.Pointer("alice").Deleted())
	assert.True(t, UserFixtures.Pointer("deleted").Deleted())
}
