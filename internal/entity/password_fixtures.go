package entity

// PasswordFixture is a test password with its UID, hashed only when a Password is built from it.
type PasswordFixture struct {
	UID      string
	Password string
}

// PasswordMap maps fixture names to test passwords.
type PasswordMap map[string]PasswordFixture

// Get returns the named password fixture with a fresh hash, or an empty Password if there is none.
func (m PasswordMap) Get(name string) Password {
	if f, ok := m[name]; ok {
		return NewPassword(f.UID, f.Password, false)
	}

	return Password{}
}

// Pointer returns a pointer to the named password fixture, see Get.
func (m PasswordMap) Pointer(name string) *Password {
	result := m.Get(name)
	return &result
}

var PasswordFixtures = PasswordMap{
	"alice":         {"uqxetse3cy5eo9z2", "Alice123!"},
	"bob":           {"uqxc08w3d0ej2283", "Bobbob123!"},
	"friend":        {"uqxqg7i1kperxvu7", "!Friend321"},
	"fowler":        {"urinotv3d6jedvlm", "PleaseChange$42"},
	"jane":          {"usamyuogp49vd4lh", "Jane123!"},
	"deleted":       {"uqxqg7i1kperxvu8", "Deleted123!"},
	"metrics":       {"cs5cpu17n6gj2qo5", "xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"},
	"isolated":      {"cs5cpu17n6gj2rvk", "aQw7tPz2LkNv9XrBeCdH4FsMgJyU1Two"},
	"no_local_auth": {"usg73p55zwgr1ytr", "None123!"},
	"2fa":           {"usg73p55zwgr1ojy", "2fa-123!"},
	"gandalf":       {"uqxetse3cy5eu3r0", "Gandalf123!"},
}

// CreatePasswordFixtures inserts known entities into the database for testing.
func CreatePasswordFixtures() {
	for name := range PasswordFixtures {
		createFixture(PasswordFixtures.Pointer(name))
	}
}
