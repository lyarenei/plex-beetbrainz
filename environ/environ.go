package environ

import (
	"os"
	"strings"
)

func GetApiToken(user string) string {
	ut := os.Getenv("USER_TOKENS")
	pairs := strings.Split(ut, ",")
	for _, pair := range pairs {
		values := strings.Split(pair, ":")
		if len(values) > 1 && strings.EqualFold(values[0], user) {
			return values[1]
		}
	}

	return ""
}
