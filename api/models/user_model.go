package models

type User struct {
	Id             int
	Username       string
	Saltedpassword string
	Salt           string
}
