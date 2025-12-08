package api

import (
	"net/http"
)

// MethodsPutPost defines a string slice that contains the PUT and POST request methods.
var MethodsPutPost = []string{http.MethodPut, http.MethodPost}
