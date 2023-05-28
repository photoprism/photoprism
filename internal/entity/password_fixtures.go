package entity

type PasswordMap map[string]Password

func (m PasswordMap) Get(name string) Password {
	if result, ok := m[name]; ok {
		return result
	}

	return Password{}
}

func (m PasswordMap) Pointer(name string) *Password {
	if result, ok := m[name]; ok {
		return &result
	}

	return &Password{}
}

var PasswordFixtures = PasswordMap{
	"alice":  NewPassword("uqxetse3cy5eo9z2", "Alice123!", false),
	"bob":    NewPassword("uqxc08w3d0ej2283", "Bobbob123!", false),
	"friend": NewPassword("uqxqg7i1kperxvu7", "!Friend321", false),
	"fowler": NewPassword("urinotv3d6jedvlm", "PleaseChange$42", false),
}

// CreatePasswordFixtures inserts known entities into the database for testing.
func CreatePasswordFixtures() {
	for _, entity := range PasswordFixtures {
		Db().Create(&entity)
	}
}
