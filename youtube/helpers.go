// Copyright The Golang Crawler Author
// SPDX-License-Identifier: Apache-2.0

package youtube

import (
	"strconv"
	"strings"
)

func parseViewCount(val string) uint64 {
	if val == "" {
		return 0
	}

	// Remove non-numeric formatting (commas, text, etc.)
	clean := strings.ReplaceAll(val, ",", "")
	clean = strings.Fields(clean)[0] // Take only the number part

	// Convert to uint64
	num, err := strconv.ParseUint(clean, 10, 64)

	if err != nil {
		return 0
	}

	return num
}
func parseDurationToSeconds(input string) uint64 {
	segments := strings.Split(input, ":")
	var h, m, s uint64

	if len(segments) == 3 {
		h, _ = strconv.ParseUint(segments[0], 10, 64)
		m, _ = strconv.ParseUint(segments[1], 10, 64)
		s, _ = strconv.ParseUint(segments[2], 10, 64)
	} else if len(segments) == 2 {
		m, _ = strconv.ParseUint(segments[0], 10, 64)
		s, _ = strconv.ParseUint(segments[1], 10, 64)
	} else if len(segments) == 1 {
		s, _ = strconv.ParseUint(segments[0], 10, 64)
	} else {
		return 0
	}

	return h*3600 + m*60 + s
}
