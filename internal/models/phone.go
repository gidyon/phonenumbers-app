package models

import "time"

type Phone struct {
	ID         uint `gorm:"primaryKey;autoIncrement"`
	Country    `gorm:"embedded"`
	Number     string    `gorm:"index;type:varchar(20);"`
	CustId     string    `gorm:"index;type:varchar(32);"`
	PhoneValid bool      `gorm:"index;type:tinyint(1)"`
	CreateDate time.Time `gorm:"index;autoCreateTime"`
}

func (*Phone) TableName() string {
	return "phones"
}
