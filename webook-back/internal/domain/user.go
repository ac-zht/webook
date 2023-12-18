package domain

import "time"

type User struct {
	Id       int64
	Email    string
	Phone    string
	Nickname string
	Birthday time.Time
	AboutMe  string

	Password string
	Ctime    time.Time
}
