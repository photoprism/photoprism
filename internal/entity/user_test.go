package entity

import (
	"testing"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/form"

	"github.com/stretchr/testify/assert"
)

func TestFindUserByName(t *testing.T) {
	t.Run("admin", func(t *testing.T) {
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
		assert.Equal(t, "Admin", m.FullName)
		assert.True(t, m.RoleAdmin)
		assert.False(t, m.RoleGuest)
		assert.False(t, m.UserDisabled)
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})

	t.Run("alice", func(t *testing.T) {
		m := FindUserByName("alice")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, 5, m.ID)
		assert.Equal(t, "uqxetse3cy5eo9z2", m.UserUID)
		assert.Equal(t, "alice", m.UserName)
		assert.Equal(t, "Alice", m.FullName)
		assert.Equal(t, "alice@example.com", m.PrimaryEmail)
		assert.True(t, m.RoleAdmin)
		assert.False(t, m.RoleGuest)
		assert.False(t, m.UserDisabled)
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})

	t.Run("bob", func(t *testing.T) {
		m := FindUserByName("bob")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, 7, m.ID)
		assert.Equal(t, "uqxc08w3d0ej2283", m.UserUID)
		assert.Equal(t, "bob", m.UserName)
		assert.Equal(t, "Bob", m.FullName)
		assert.Equal(t, "bob@example.com", m.PrimaryEmail)
		assert.False(t, m.RoleAdmin)
		assert.False(t, m.RoleGuest)
		assert.False(t, m.UserDisabled)
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})

	t.Run("unknown", func(t *testing.T) {
		m := FindUserByName("")

		if m != nil {
			t.Fatal("result should be nil")
		}
	})

	t.Run("not found", func(t *testing.T) {
		m := FindUserByName("xxx")

		if m != nil {
			t.Fatal("result should be nil")
		}
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
		p := User{UserUID: "u000000000000010", UserName: "Hans", FullName: ""}
		err := p.Save()
		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, p.InvalidPassword("abcdef"))

	})
	t.Run("not registered", func(t *testing.T) {
		p := User{UserUID: "u12", UserName: "", FullName: ""}
		assert.True(t, p.InvalidPassword("abcdef"))
	})
	t.Run("password empty", func(t *testing.T) {
		p := User{UserUID: "u000000000000011", UserName: "User", FullName: ""}
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
		assert.Equal(t, "", m.UserName)
		assert.Equal(t, "Guest", m.FullName)
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
		assert.Equal(t, "alice", m.Username())
		assert.Equal(t, "Alice", m.FullName)
		assert.Equal(t, "alice@example.com", m.PrimaryEmail)
		assert.True(t, m.RoleAdmin)
		assert.False(t, m.RoleGuest)
		assert.False(t, m.UserDisabled)
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
		assert.Equal(t, "bob", m.UserName)
		assert.Equal(t, "Bob", m.FullName)
		assert.Equal(t, "bob@example.com", m.PrimaryEmail)
		assert.False(t, m.RoleAdmin)
		assert.False(t, m.RoleGuest)
		assert.False(t, m.UserDisabled)
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
		assert.False(t, m.RoleAdmin)
		assert.False(t, m.RoleGuest)
		assert.True(t, m.RoleFriend)
		assert.True(t, m.UserDisabled)
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})

}

func TestUser_String(t *testing.T) {
	t.Run("UID", func(t *testing.T) {
		p := User{UserUID: "abc123", UserName: "", FullName: ""}
		assert.Equal(t, "abc123", p.String())
	})
	t.Run("DisplayName", func(t *testing.T) {
		p := User{UserUID: "abc123", UserName: "", FullName: "Test"}
		assert.Equal(t, "Test", p.String())
	})
	t.Run("UserName", func(t *testing.T) {
		p := User{UserUID: "abc123", UserName: "Super-User ", FullName: "Test"}
		assert.Equal(t, "super-user", p.String())
	})
}

func TestUser_Admin(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", UserName: "Hanna", FullName: "", RoleAdmin: true}
		assert.True(t, p.Admin())
	})
	t.Run("false", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", UserName: "Hanna", FullName: "", RoleAdmin: false}
		assert.False(t, p.Admin())
	})
}

func TestUser_Anonymous(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		p := User{UserUID: "", UserName: "Hanna", FullName: ""}
		assert.True(t, p.Anonymous())
	})
	t.Run("false", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", UserName: "Hanna", FullName: "", RoleAdmin: true}
		assert.False(t, p.Anonymous())
	})
}

func TestUser_Guest(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		p := User{UserUID: "", UserName: "Hanna", FullName: "", RoleGuest: true}
		assert.True(t, p.Guest())
	})
	t.Run("false", func(t *testing.T) {
		p := User{UserUID: "", UserName: "Hanna", FullName: ""}
		assert.False(t, p.Guest())
	})
}

func TestUser_SetPassword(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", UserName: "Hanna", FullName: ""}
		if err := p.SetPassword("insecure"); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("not registered", func(t *testing.T) {
		p := User{UserUID: "", UserName: "Hanna", FullName: ""}
		assert.Error(t, p.SetPassword("insecure"))

	})
	t.Run("password too short", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", UserName: "Hanna", FullName: ""}
		assert.Error(t, p.SetPassword("cat"))
	})
}

func TestUser_InitPassword(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		p := User{UserUID: "u000000000000009", UserName: "Hanna", FullName: ""}
		assert.Nil(t, FindPassword("u000000000000009"))
		p.InitPassword("insecure")
		m := FindPassword("u000000000000009")

		if m == nil {
			t.Fatal("result should not be nil")
		}
	})
	t.Run("already existing", func(t *testing.T) {
		p := User{UserUID: "u000000000000010", UserName: "Hans", FullName: ""}

		if err := p.Save(); err != nil {
			t.Logf("cannot user %s: ", err)
		}

		if err := p.SetPassword("insecure"); err != nil {
			t.Fatal(err)
		}

		assert.NotNil(t, FindPassword("u000000000000010"))
		p.InitPassword("insecure")
		m := FindPassword("u000000000000010")

		if m == nil {
			t.Fatal("result should not be nil")
		}
	})
	t.Run("not registered", func(t *testing.T) {
		p := User{UserUID: "u12", UserName: "", FullName: ""}
		assert.Nil(t, FindPassword("u12"))
		p.InitPassword("insecure")
		assert.Nil(t, FindPassword("u12"))
	})
	t.Run("password empty", func(t *testing.T) {
		p := User{UserUID: "u000000000000011", UserName: "User", FullName: ""}
		assert.Nil(t, FindPassword("u000000000000011"))
		p.InitPassword("")
		assert.Nil(t, FindPassword("u000000000000011"))
	})
}

func TestUser_Role(t *testing.T) {
	t.Run("admin", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", UserName: "Hanna", FullName: "", RoleAdmin: true}
		assert.Equal(t, acl.Role("admin"), p.Role())
	})
	t.Run("child", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", UserName: "Hanna", FullName: "", RoleChild: true}
		assert.Equal(t, acl.Role("child"), p.Role())
	})
	t.Run("family", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", UserName: "Hanna", FullName: "", RoleFamily: true}
		assert.Equal(t, acl.Role("family"), p.Role())
	})
	t.Run("friend", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", UserName: "Hanna", FullName: "", RoleFriend: true}
		assert.Equal(t, acl.Role("friend"), p.Role())
	})
	t.Run("guest", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", UserName: "Hanna", FullName: "", RoleGuest: true}
		assert.Equal(t, acl.Role("guest"), p.Role())
	})
	t.Run("default", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", UserName: "Hanna", FullName: ""}
		assert.Equal(t, acl.Role("*"), p.Role())
	})
}

func TestUser_Validate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		u := &User{
			AddressID:    1,
			UserName:     "validate",
			FullName:     "Validate",
			PrimaryEmail: "validate@example.com",
		}
		err := u.Validate()
		assert.Nil(t, err)
	})
	t.Run("username empty", func(t *testing.T) {
		u := &User{
			AddressID:    1,
			UserName:     "",
			FullName:     "Validate",
			PrimaryEmail: "validate@example.com",
		}
		err := u.Validate()
		assert.Error(t, err)
	})
	t.Run("username too short", func(t *testing.T) {
		u := &User{
			AddressID:    1,
			UserName:     "va",
			FullName:     "Validate",
			PrimaryEmail: "validate@example.com",
		}
		err := u.Validate()
		assert.Error(t, err)
	})
	t.Run("username not unique", func(t *testing.T) {
		FirstOrCreateUser(&User{
			AddressID: 1,
			UserName:  "notunique1",
		})
		u := &User{
			AddressID:    1,
			UserName:     "notunique1",
			FullName:     "Not Unique",
			PrimaryEmail: "notunique1@example.com",
		}
		assert.Error(t, u.Validate())
	})
	t.Run("email not unique", func(t *testing.T) {
		FirstOrCreateUser(&User{
			AddressID:    1,
			PrimaryEmail: "notunique2@example.com",
		})
		u := &User{
			AddressID:    1,
			UserName:     "notunique2",
			FullName:     "Not Unique",
			PrimaryEmail: "notunique2@example.com",
		}
		assert.Error(t, u.Validate())
	})
	t.Run("update user - email not unique", func(t *testing.T) {
		FirstOrCreateUser(&User{
			AddressID:    1,
			UserName:     "notunique3",
			FullName:     "Not Unique",
			PrimaryEmail: "notunique3@example.com",
		})
		u := FirstOrCreateUser(&User{
			AddressID:    1,
			UserName:     "notunique30",
			FullName:     "Not Unique",
			PrimaryEmail: "notunique3@example.com",
		})
		u.UserName = "notunique3"
		assert.Error(t, u.Validate())
	})
	t.Run("primary email empty", func(t *testing.T) {
		FirstOrCreateUser(&User{
			AddressID: 1,
			UserName:  "nnomail",
		})
		u := &User{
			AddressID:    1,
			UserName:     "nomail",
			FullName:     "No Mail",
			PrimaryEmail: "",
		}
		assert.Nil(t, u.Validate())
	})
}

func TestCreateWithPassword(t *testing.T) {
	t.Run("password too short", func(t *testing.T) {
		u := form.UserCreate{
			UserName: "thomas1",
			FullName: "Thomas One",
			Email:    "thomas1@example.com",
			Password: "hel",
		}
		err := CreateWithPassword(u)
		assert.Error(t, err)
	})
	t.Run("valid", func(t *testing.T) {
		u := form.UserCreate{
			UserName: "thomas2",
			FullName: "Thomas Two",
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
			AddressID:    1,
			UserName:     "thomasdel",
			FullName:     "Thomas Delete",
			PrimaryEmail: "thomasdel@example.com",
		}
		u = FirstOrCreateUser(u)
		err := u.Delete()
		assert.NoError(t, err)
	})
	t.Run("delete user - not in db", func(t *testing.T) {
		u := &User{
			AddressID:    1,
			UserName:     "thomasdel2",
			FullName:     "Thomas Delete 2",
			PrimaryEmail: "thomasdel2@example.com",
		}
		err := u.Delete()
		assert.Error(t, err)
	})
}

func TestUser_Deleted(t *testing.T) {
	assert.False(t, UserFixtures.Pointer("alice").Deleted())
	assert.True(t, UserFixtures.Pointer("deleted").Deleted())
}
