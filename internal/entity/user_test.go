package entity

import (
	"testing"

	"github.com/photoprism/photoprism/internal/acl"

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
}

func TestUser_String(t *testing.T) {
	t.Run("return UID", func(t *testing.T) {
		p := User{UserUID: "abc123", UserName: "", FullName: ""}
		assert.Equal(t, "abc123", p.String())
	})
	t.Run("return display name", func(t *testing.T) {
		p := User{UserUID: "abc123", UserName: "", FullName: "Test"}
		assert.Equal(t, "Test", p.String())
	})
	t.Run("return user name", func(t *testing.T) {
		p := User{UserUID: "abc123", UserName: "Super-User", FullName: "Test"}
		assert.Equal(t, "Super-User", p.String())
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
		if err := p.SetPassword("abcdefgt"); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("not registered", func(t *testing.T) {
		p := User{UserUID: "", UserName: "Hanna", FullName: ""}
		assert.Error(t, p.SetPassword("abchjy"))

	})
	t.Run("password too short", func(t *testing.T) {
		p := User{UserUID: "u000000000000008", UserName: "Hanna", FullName: ""}
		assert.Error(t, p.SetPassword("abc"))
	})
}

func TestUser_InitPassword(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		p := User{UserUID: "u000000000000009", UserName: "Hanna", FullName: ""}
		assert.Nil(t, FindPassword("u000000000000009"))
		p.InitPassword("abcdek")
		m := FindPassword("u000000000000009")

		if m == nil {
			t.Fatal("result should not be nil")
		}
	})
	t.Run("already existing", func(t *testing.T) {
		p := User{UserUID: "u000000000000010", UserName: "Hans", FullName: ""}
		if err := p.Save(); err != nil {
			t.Fatal(err)
		}
		if err := p.SetPassword("hutfdt"); err != nil {
			t.Fatal(err)
		}
		assert.NotNil(t, FindPassword("u000000000000010"))
		p.InitPassword("hutfdt")
		m := FindPassword("u000000000000010")

		if m == nil {
			t.Fatal("result should not be nil")
		}
	})
	t.Run("not registered", func(t *testing.T) {
		p := User{UserUID: "u12", UserName: "", FullName: ""}
		assert.Nil(t, FindPassword("u12"))
		p.InitPassword("dcjygkh")
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
