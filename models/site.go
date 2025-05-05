package models

type Site struct {
	Id          int
	Sitename    string
	Siteurl     string
	Sitelogo    string
	Description string
	Keywords    string
	Aboutweb    string
	Aboutme     string
	Copyright   string
}

func (Site) TableName() string {
	return "site"
}
