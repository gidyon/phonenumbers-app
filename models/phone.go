package models

import (
	"errors"
	"regexp"
)

var (
	cameroonRegexp   = regexp.MustCompile(`\(237\)\ ?[2368]\d{7,8}$`)
	ethiopiaRegexp   = regexp.MustCompile(`\(251\)\ ?[1-59]\d{8}$`)
	moroccoRegexp    = regexp.MustCompile(`\(212\)\ ?[5-9]\d{8}$`)
	mozambiqueRegexp = regexp.MustCompile(`\(258\)\ ?[28]\d{7,8}$`)
	ugandaRegexp     = regexp.MustCompile(`\(256\)\ ?\d{9}$`)
)

const (
	ValidState    = "VALID"
	NotValidState = "NOT_VALID"
)

type Phone struct {
	ID      uint `gorm:"primaryKey;autoIncrement"`
	Country `gorm:"embedded;embeddedPrefix:country_"`
	State   string `gorm:"type:varchar(20)"`
	Number  string `gorm:"index;type:varchar(20);"`
}

func NewPhoneModel(country, number string) (*Phone, error) {
	phone := &Phone{
		ID: 0,
		Country: Country{
			Code: 0,
			Name: country,
		},
		State:  ValidState,
		Number: number,
	}
	err := phone.Validate()
	if err != nil {
		return nil, err
	}
	return phone, nil
}

func (p *Phone) IsValid() bool {
	return p.State == ValidState
}

func (p *Phone) Validate() error {
	var valid bool
	switch p.Country.Name {
	case "Cameroon":
		valid = cameroonRegexp.MatchString(p.Number)
		p.Code = 237
	case "Ethiopia":
		valid = ethiopiaRegexp.MatchString(p.Number)
		p.Code = 251
	case "Morocco":
		valid = moroccoRegexp.MatchString(p.Number)
		p.Code = 212
	case "Mozambique":
		valid = mozambiqueRegexp.MatchString(p.Number)
		p.Code = 258
	case "Uganda":
		valid = ugandaRegexp.MatchString(p.Number)
		p.Code = 256
	default:
		return errors.New("failed to validate phone")
	}
	if valid {
		p.State = ValidState
	} else {
		p.State = NotValidState
	}
	return nil
}
