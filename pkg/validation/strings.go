package validation

import "regexp"

var OnlyDigits = OnlyDigitsValidator()

type StringValidator func(string) bool

func OnlyDigitsValidator() StringValidator {
	pattern := regexp.MustCompile("\\d+")
	return func(s string) bool {
		return pattern.MatchString(s)
	}
}
