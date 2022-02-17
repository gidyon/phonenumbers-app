package models

type Country struct {
	ID   uint   `gorm:"primaryKey;autoIncrement"`
	Code uint   `gorm:"type:int(3)"`
	Name string `gorm:"type:varchar(40)"`
}
