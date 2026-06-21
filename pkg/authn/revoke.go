package authn

// SessionScope represents a session revocation target.
type SessionScope uint8

const (
	// RevokeLoginSessions removes interactive login sessions while keeping app
	// passwords, client access tokens, and sessions derived from app passwords.
	RevokeLoginSessions SessionScope = iota
	// RevokeDerivedSessions additionally removes sessions derived from an app
	// password while keeping the app passwords themselves, so configured devices keep working.
	RevokeDerivedSessions
	// RevokeAllSessions removes every session, including app password records,
	// client access tokens, and sessions derived from them.
	RevokeAllSessions
)
