package domain

import "time"

type Article struct {
	Id      int64
	Title   string
	Status  ArticleStatus
	Content string
	Author  Author
	Ctime   time.Time
	Utime   time.Time
}

func (a Article) Abstract() string {
	cs := []rune(a.Content)
	if len(cs) < 100 {
		return a.Content
	}
	return string(cs[:100])
}

type Author struct {
	Id   int64
	Name string
}

type ArticleStatus uint8

const (
	ArticleStatusUnknown ArticleStatus = iota
	ArticleStatusUnpublished
	ArticleStatusPublished
	ArticleStatusPrivate
)
