package web

import "github.com/zht-account/webook/internal/domain"

type ArticleVo struct {
	Id       int64  `json:"id"`
	Title    string `json:"title"`
	Abstract string `json:"abstract"`

	Content string `json:"content"`
	Status  uint8  `json:"status"`
	Author  string `json:"author"`
	Ctime   string `json:"ctime"`
	Utime   string `json:"utime"`

	LikeCnt    int64 `json:"likeCnt"`
	CollectCnt int64 `json:"collectCnt"`
	ReadCnt    int64 `json:"readCnt"`

	Liked     bool `json:"liked"`
	Collected bool `json:"collected"`
}

type LikeReq struct {
	Id   int64 `json:"id"`
	Like bool  `json:"like"`
}

type CollectReq struct {
	Id  int64 `json:"id"`
	Cid int64 `json:"cid"`
}

type ArticleReq struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (req ArticleReq) toDomain(uid int64) domain.Article {
	return domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uid,
		},
	}
}
