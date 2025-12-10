package config

var (
	// SignUpURL points to the membership sign-up page.
	SignUpURL = "https://www.photoprism.app/membership"
	// MsgSponsor is the default sponsorship message shown to users.
	MsgSponsor = "Become a member today, support our mission and enjoy our member benefits! 💎"
	// MsgSignUp is the default sign-up helper text.
	MsgSignUp = "Visit " + SignUpURL + " to learn more."
	// SignUp bundles the sign-up message and URL used in client responses.
	SignUp = Values{"message": MsgSponsor, "url": SignUpURL}
)
