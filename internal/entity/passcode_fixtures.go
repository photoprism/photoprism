package entity

type PasscodeMap map[string]Passcode

func (m PasscodeMap) Get(name string) Passcode {
	if result, ok := m[name]; ok {
		return result
	}

	return Passcode{}
}

func (m PasscodeMap) Pointer(name string) *Passcode {
	if result, ok := m[name]; ok {
		return &result
	}

	return &Passcode{}
}

var (
	PasscodeFixtureAlice, _ = NewPasscode("uqxetse3cy5eo9z2", "otpauth://totp/PhotoPrism:alice?algorithm=SHA1&digits=6&issuer=PhotoPrism%20Pro&period=30&secret=LKBTPGHABW2BVQVIROIGFTLQV4IRBXMV", "0t37foocgp2w")
	PasscodeFixtureJane, _  = NewPasscode("usamyuogp49vd4lh", "otpauth://totp/PhotoPrism:jane?algorithm=SHA1&digits=6&issuer=PhotoPrism%20Pro&period=30&secret=RUYYIDJZBJLKD6OL6WFBJO6PXEZOYIZW", "0wg68oc6jg92")
	PasscodeFixtures        = PasscodeMap{
		"alice": *PasscodeFixtureAlice,
		"jane":  *PasscodeFixtureJane,
	}
)

// CreatePasscodeFixtures inserts known entities into the database for testing.
func CreatePasscodeFixtures() {
	for _, entity := range PasscodeFixtures {
		Db().Create(&entity)
	}
}
