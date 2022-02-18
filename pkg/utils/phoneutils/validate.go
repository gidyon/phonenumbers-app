package phoneutils

import (
	"regexp"

	phonebook_v1 "github.com/gidyon/jumia-exercise/pkg/api/phonebook/v1"
)

const (
	ValidState    = "VALID"
	NotValidState = "NOT_VALID"
)

var (
	cameroonRegexp   = regexp.MustCompile(`\(237\)\ ?[2368]\d{7,8}$`)
	ethiopiaRegexp   = regexp.MustCompile(`\(251\)\ ?[1-59]\d{8}$`)
	moroccoRegexp    = regexp.MustCompile(`\(212\)\ ?[5-9]\d{8}$`)
	mozambiqueRegexp = regexp.MustCompile(`\(258\)\ ?[28]\d{7,8}$`)
	ugandaRegexp     = regexp.MustCompile(`\(256\)\ ?\d{9}$`)
)

func ValidatePhone(pr *phonebook_v1.PhoneRecord) bool {
	var valid bool
	switch pr.CountryName {
	case "Cameroon":
		valid = cameroonRegexp.MatchString(pr.Number)
		pr.CountryCode = 237
	case "Ethiopia":
		valid = ethiopiaRegexp.MatchString(pr.Number)
		pr.CountryCode = 251
	case "Morocco":
		valid = moroccoRegexp.MatchString(pr.Number)
		pr.CountryCode = 212
	case "Mozambique":
		valid = mozambiqueRegexp.MatchString(pr.Number)
		pr.CountryCode = 258
	case "Uganda":
		valid = ugandaRegexp.MatchString(pr.Number)
		pr.CountryCode = 256
	default:
		valid = false
	}
	if valid {
		pr.PhoneValid = true
	} else {
		pr.PhoneValid = false
	}
	return valid
}
