package entity

import (
	"github.com/photoprism/photoprism/internal/acl"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindPersonByUserName(t *testing.T) {
	t.Run("admin", func(t *testing.T) {
		m := FindPersonByUserName("admin")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, 1, m.ID)
		assert.NotEmpty(t, m.PersonUID)
		assert.Equal(t, "admin", m.UserName)
		assert.Equal(t, "Admin", m.DisplayName)
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})

	t.Run("unknown", func(t *testing.T) {
		m := FindPersonByUserName("")

		if m != nil {
			t.Fatal("result should be nil")
		}
	})

	t.Run("not found", func(t *testing.T) {
		m := FindPersonByUserName("xxx")

		if m != nil {
			t.Fatal("result should be nil")
		}
	})
}

func TestPerson_InvalidPassword(t *testing.T) {
	t.Run("admin", func(t *testing.T) {
		m := FindPersonByUserName("admin")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.False(t, m.InvalidPassword("photoprism"))
	})
	t.Run("no password existing", func(t *testing.T) {
		p := Person{PersonUID: "u000000000000010", UserName: "Hans", DisplayName: ""}
		p.Save()
		assert.True(t, p.InvalidPassword("abcdef"))

	})
	t.Run("not registered", func(t *testing.T) {
		p := Person{PersonUID: "u12", UserName: "", DisplayName: ""}
		assert.True(t, p.InvalidPassword("abcdef"))
	})
	t.Run("password empty", func(t *testing.T) {
		p := Person{PersonUID: "u000000000000011", UserName: "User", DisplayName: ""}
		assert.True(t, p.InvalidPassword(""))
	})
}

func TestPerson_Save(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		p := Person{}

		err := p.Save()

		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestFirstOrCreatePerson(t *testing.T) {
	t.Run("not existing", func(t *testing.T) {
		p := &Person{ID: 555}

		result := FirstOrCreatePerson(p)
		if result == nil {
			t.Fatal("result should not be nil")
		}

		assert.NotEmpty(t, result.ID)

	})
	t.Run("existing", func(t *testing.T) {
		p := &Person{ID: 1234}
		err := p.Save()

		if err != nil {
			t.Fatal(err)
		}

		result := FirstOrCreatePerson(p)

		if result == nil {
			t.Fatal("result should not be nil")
		}
		assert.NotEmpty(t, result.ID)
	})
}

func TestFindPersonByUID(t *testing.T) {
	t.Run("guest", func(t *testing.T) {
		m := FindPersonByUID("u000000000000002")

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, -2, m.ID)
		assert.NotEmpty(t, m.PersonUID)
		assert.Equal(t, "", m.UserName)
		assert.Equal(t, "Guest", m.DisplayName)
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})

	t.Run("unknown", func(t *testing.T) {
		m := FindPersonByUID("")

		if m != nil {
			t.Fatal("result should be nil")
		}
	})

	t.Run("not found", func(t *testing.T) {
		m := FindPersonByUID("xxx")

		if m != nil {
			t.Fatal("result should be nil")
		}
	})
}

func TestPerson_String(t *testing.T) {
	t.Run("return UID", func(t *testing.T) {
		p := Person{PersonUID: "abc123", UserName: "", DisplayName: ""}
		assert.Equal(t, "abc123", p.String())
	})
	t.Run("return display name", func(t *testing.T) {
		p := Person{PersonUID: "abc123", UserName: "", DisplayName: "Test"}
		assert.Equal(t, "Test", p.String())
	})
	t.Run("return user name", func(t *testing.T) {
		p := Person{PersonUID: "abc123", UserName: "Super-User", DisplayName: "Test"}
		assert.Equal(t, "Super-User", p.String())
	})
}

func TestPerson_Admin(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		p := Person{PersonUID: "u000000000000008", UserName: "Hanna", DisplayName: "", RoleAdmin: true}
		assert.True(t, p.Admin())
	})
	t.Run("false", func(t *testing.T) {
		p := Person{PersonUID: "u000000000000008", UserName: "Hanna", DisplayName: "", RoleAdmin: false}
		assert.False(t, p.Admin())
	})
}

func TestPerson_Anonymous(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		p := Person{PersonUID: "", UserName: "Hanna", DisplayName: ""}
		assert.True(t, p.Anonymous())
	})
	t.Run("false", func(t *testing.T) {
		p := Person{PersonUID: "u000000000000008", UserName: "Hanna", DisplayName: "", RoleAdmin: true}
		assert.False(t, p.Anonymous())
	})
}

func TestPerson_Guest(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		p := Person{PersonUID: "", UserName: "Hanna", DisplayName: "", RoleGuest: true}
		assert.True(t, p.Guest())
	})
	t.Run("false", func(t *testing.T) {
		p := Person{PersonUID: "", UserName: "Hanna", DisplayName: ""}
		assert.False(t, p.Guest())
	})
}

func TestPerson_SetPassword(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		p := Person{PersonUID: "u000000000000008", UserName: "Hanna", DisplayName: ""}
		if err := p.SetPassword("abcdefgt"); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("not registered", func(t *testing.T) {
		p := Person{PersonUID: "", UserName: "Hanna", DisplayName: ""}
		assert.Error(t, p.SetPassword("abchjy"))

	})
	t.Run("password too short", func(t *testing.T) {
		p := Person{PersonUID: "u000000000000008", UserName: "Hanna", DisplayName: ""}
		assert.Error(t, p.SetPassword("abc"))
	})
}

func TestPerson_InitPassword(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		p := Person{PersonUID: "u000000000000009", UserName: "Hanna", DisplayName: ""}
		assert.Nil(t, FindPassword("u000000000000009"))
		p.InitPassword("abcdek")
		m := FindPassword("u000000000000009")

		if m == nil {
			t.Fatal("result should not be nil")
		}
	})
	t.Run("already existing", func(t *testing.T) {
		p := Person{PersonUID: "u000000000000010", UserName: "Hans", DisplayName: ""}
		p.Save()
		p.SetPassword("hutfdt")
		assert.NotNil(t, FindPassword("u000000000000010"))
		p.InitPassword("hutfdt")
		m := FindPassword("u000000000000010")

		if m == nil {
			t.Fatal("result should not be nil")
		}
	})
	t.Run("not registered", func(t *testing.T) {
		p := Person{PersonUID: "u12", UserName: "", DisplayName: ""}
		assert.Nil(t, FindPassword("u12"))
		p.InitPassword("dcjygkh")
		assert.Nil(t, FindPassword("u12"))
	})
	t.Run("password empty", func(t *testing.T) {
		p := Person{PersonUID: "u000000000000011", UserName: "User", DisplayName: ""}
		assert.Nil(t, FindPassword("u000000000000011"))
		p.InitPassword("")
		assert.Nil(t, FindPassword("u000000000000011"))
	})
}

func TestPerson_Role(t *testing.T) {
	t.Run("admin", func(t *testing.T) {
		p := Person{PersonUID: "u000000000000008", UserName: "Hanna", DisplayName: "", RoleAdmin: true}
		assert.Equal(t, acl.Role("admin"), p.Role())
	})
	t.Run("child", func(t *testing.T) {
		p := Person{PersonUID: "u000000000000008", UserName: "Hanna", DisplayName: "", RoleChild: true}
		assert.Equal(t, acl.Role("child"), p.Role())
	})
	t.Run("family", func(t *testing.T) {
		p := Person{PersonUID: "u000000000000008", UserName: "Hanna", DisplayName: "", RoleFamily: true}
		assert.Equal(t, acl.Role("family"), p.Role())
	})
	t.Run("friend", func(t *testing.T) {
		p := Person{PersonUID: "u000000000000008", UserName: "Hanna", DisplayName: "", RoleFriend: true}
		assert.Equal(t, acl.Role("friend"), p.Role())
	})
	t.Run("guest", func(t *testing.T) {
		p := Person{PersonUID: "u000000000000008", UserName: "Hanna", DisplayName: "", RoleGuest: true}
		assert.Equal(t, acl.Role("guest"), p.Role())
	})
	t.Run("default", func(t *testing.T) {
		p := Person{PersonUID: "u000000000000008", UserName: "Hanna", DisplayName: ""}
		assert.Equal(t, acl.Role("*"), p.Role())
	})
}
