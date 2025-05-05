package models

import "time"

type Website struct {
	Id          int    `form:"id" json:"id"`
	Name        string `form:"name" json:"name"`
	GroupId     int    `form:"group_id" json:"group_id"`
	Url         string `form:"url" json:"url"`
	Logo        string `form:"logo" json:"logo"`
	Description string `form:"description" json:"description"`
	CreatedAt   time.Time
	Group       Group `gorm:"ForeignKey:GroupId"`
}
