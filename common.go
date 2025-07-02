package arbor

import "strings"

func isEmpty(input string) bool {
	return (len(strings.TrimSpace(input)) <= 0)
}
