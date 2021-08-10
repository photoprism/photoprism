package entity

type UserMap map[string]User

func (m UserMap) Get(name string) User {
	if result, ok := m[name]; ok {
		return result
	}

	return User{}
}

func (m UserMap) Pointer(name string) *User {
	if result, ok := m[name]; ok {
		return &result
	}

	return &User{}
}

var UserFixtures = UserMap{
	"alice": {
		ID:           5,
		AddressID:    1,
		UserUID:      "uqxetse3cy5eo9z2",
		UserName:     "alice",
		FullName:     "Alice",
		RoleAdmin:    true,
		RoleGuest:    false,
		UserDisabled: false,
		PrimaryEmail: "alice@example.com",
	},
	"bob": {
		ID:           7,
		AddressID:    1,
		UserUID:      "uqxc08w3d0ej2283",
		UserName:     "bob",
		FullName:     "Bob",
		RoleAdmin:    false,
		RoleGuest:    false,
		UserDisabled: false,
		PrimaryEmail: "bob@example.com",
	},
}

// CreateUserFixtures inserts known entities into the database for testing.
func CreateUserFixtures() {
	for _, entity := range UserFixtures {
		Db().Create(&entity)
	}
}
