// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"strings"
)

// tokenPrefix is the string preceding the token in the Authorization header.
const tokenPrefix = "Bearer "

// ExtractToken extracts the token from a correctly formatted Authorization header.
func ExtractToken(headerValue string) string {
	if strings.HasPrefix(headerValue, tokenPrefix) {
		return headerValue[len(tokenPrefix):]
	} else {
		return ""
	}
}
