package config

import "time"

const MsgSponsor = "PhotoPrism® needs your support!"
const SignUpURL = "https://docs.photoprism.app/funding/"
const MsgSignUp = "Visit " + SignUpURL + " to learn more."
const MsgSponsorCommand = "Since running this command puts additional load on our infrastructure," +
	" we unfortunately can only offer it to sponsors."

const ApiUri = "/api/v1"    // REST API
const StaticUri = "/static" // Static Content

const DefaultAutoIndexDelay = int(5 * 60)  // 5 Minutes
const DefaultAutoImportDelay = int(3 * 60) // 3 Minutes

const DefaultWakeupIntervalSeconds = int(15 * 60) // 15 Minutes
const DefaultWakeupInterval = time.Second * time.Duration(DefaultWakeupIntervalSeconds)
const MaxWakeupInterval = time.Hour * 24 // 1 Day

// Megabyte in bytes.
const Megabyte = 1000 * 1000

// Gigabyte in bytes.
const Gigabyte = Megabyte * 1000

// MinMem is the minimum amount of system memory required.
const MinMem = Gigabyte

// RecommendedMem is the recommended amount of system memory.
const RecommendedMem = 3 * Gigabyte

// serialName is the name of the unique storage serial.
const serialName = "serial"
