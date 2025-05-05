package models

type Group struct {
	Id       int       `form:"groupid" json:"groupid"`
	Name     string    `form:"groupname" json:"groupname"`
	Logo     string    `form:"grouplogo" json:"grouplogo"`
	Sort     int       `form:"sort" json:"sort"`
	Websites []Website `gorm:"ForeignKey:GroupId"`
}
