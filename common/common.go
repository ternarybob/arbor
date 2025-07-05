package common

import "strings"

func IsEmpty(input string) bool {
	return (len(strings.TrimSpace(input)) <= 0)
}
