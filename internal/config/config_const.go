package config

import "time"

// ApiUri is the relative path for handling REST requests.
const ApiUri = "/api/v1"

// StaticUri is the relative path for serving static content.
const StaticUri = "/static"

// MsgSponsor and MsgSignUp provide sponsorship info messages;
// SignUpURL a signup link.
const MsgSponsor = "PhotoPrism® needs your support!"
const SignUpURL = "https://docs.photoprism.app/funding/"
const MsgSignUp = "Visit " + SignUpURL + " to learn more."
const MsgSponsorCommand = "Since running this command puts additional load on our infrastructure," +
	" we unfortunately can only offer it to sponsors."

// DefaultAutoIndexDelay and DefaultAutoImportDelay set the default safety delay duration
// before starting to index/import in the background.
const DefaultAutoIndexDelay = int(5 * 60)  // 5 Minutes
const DefaultAutoImportDelay = int(3 * 60) // 3 Minutes

// MinWakeupInterval and MaxWakeupInterval limit the interval duration
// in which the background worker can be invoked.
const MinWakeupInterval = time.Minute             // 1 Minute
const MaxWakeupInterval = time.Hour * 24          // 1 Day
const DefaultWakeupIntervalSeconds = int(15 * 60) // 15 Minutes
const DefaultWakeupInterval = time.Second * time.Duration(DefaultWakeupIntervalSeconds)

// Megabyte in bytes.
const Megabyte = 1000 * 1000 // 1,000,000 Bytes

// Gigabyte in bytes.
const Gigabyte = Megabyte * 1000 // 1,000,000,000 Bytes

// MinMem is the minimum amount of system memory required.
const MinMem = Gigabyte

// RecommendedMem is the recommended amount of system memory.
const RecommendedMem = 3 * Gigabyte // 3,000,000,000 Bytes

// serialName is the name of the unique storage serial.
const serialName = "serial"
