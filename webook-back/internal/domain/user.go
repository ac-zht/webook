package domain

import "time"

type User struct {
	Id       int64
	Email    string
	Password string
	Phone    string
	Nickname string
	Birthday time.Time
	AboutMe  string

	Ctime time.Time
}
