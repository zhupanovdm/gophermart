package validation

import "regexp"

func OnlyDigits(number string) bool {
	ok, _ := regexp.MatchString("\\d+", number)
	return ok
}
