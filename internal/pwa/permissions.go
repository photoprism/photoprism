package pwa

import (
	"github.com/photoprism/photoprism/pkg/list"
)

// Permissions specifies the default web app manifest permissions.
var Permissions = list.List{"geolocation", "downloads", "storage"}
