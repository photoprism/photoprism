package config

var (
	SignUpURL  = "https://www.photoprism.app/membership"
	MsgSponsor = "Become a member today, support our mission and enjoy our member benefits! 💎"
	MsgSignUp  = "Visit " + SignUpURL + " to learn more."
	SignUp     = Map{"message": MsgSponsor, "url": SignUpURL}
)
