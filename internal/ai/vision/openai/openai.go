/*
Package openai implements the PhotoPrism vision adapter that calls the OpenAI
Responses API for captions, labels, and optional markers.

Copyright (c) 2018 - 2026 PhotoPrism UG. All rights reserved.

	This program is free software: you can redistribute it and/or modify
	it under Version 3 of the GNU Affero General Public License (the "AGPL"):
	<https://docs.photoprism.app/license/agpl>

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	The AGPL is supplemented by our Trademark and Brand Guidelines,
	which describe how our Brand Assets may be used:
	<https://www.photoprism.app/trademark/>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>
*/
package openai

import (
	"strings"
)

// CloudModelFamilies lists the identifier families OpenAI uses for the models it hosts itself.
var CloudModelFamilies = []string{"gpt", "chatgpt", "o1", "o3", "o4"}

// IsCloudModel reports whether the model identifier belongs to OpenAI's own catalog.
// OpenAI-compatible servers such as vLLM, llama.cpp, and LM Studio serve open-weight models
// under their own names, so the family is what separates the hosted service from a local one.
func IsCloudModel(name string) bool {
	name = strings.ToLower(strings.TrimSpace(name))

	for _, family := range CloudModelFamilies {
		// Match the family itself or a variant of it, never a longer name that merely starts
		// with the same letters, so a local "o11vision" is not read as OpenAI's "o1".
		if name == family || strings.HasPrefix(name, family+"-") {
			return true
		}
	}

	return false
}
