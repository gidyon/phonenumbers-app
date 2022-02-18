package models

type Country struct {
	ID          uint   `gorm:"primaryKey;autoIncrement"`
	CountryCode uint   `gorm:"type:int(3)"`
	CountryName string `gorm:"type:varchar(40)"`
}
