package web

import (
	"fmt"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	intrDomain "github.com/zht-account/webook/interactive/domain"
	intr "github.com/zht-account/webook/interactive/service"
	"github.com/zht-account/webook/internal/domain"
	"github.com/zht-account/webook/internal/errs"
	"github.com/zht-account/webook/internal/service"
	"github.com/zht-account/webook/internal/web/jwt"
	"github.com/zht-account/webook/pkg/ginx"
	"github.com/zht-account/webook/pkg/logger"
	"golang.org/x/sync/errgroup"
	"net/http"
	"strconv"
	"time"
)

type ArticleHandler struct {
	svc     service.ArticleService
	intrSvc intr.InteractiveService
	l       logger.Logger
	biz     string
}

func NewArticleHandler(svc service.ArticleService, intrSvc intr.InteractiveService, l logger.Logger) *ArticleHandler {
	return &ArticleHandler{
		svc:     svc,
		intrSvc: intrSvc,
		l:       l,
		biz:     "article",
	}
}

func (a *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")
	g.POST("/edit", ginx.WrapReqAndClaims[ArticleReq](a.Edit))
	g.POST("/publish", ginx.WrapReqAndClaims[ArticleReq](a.Publish))
	g.POST("/withdraw", ginx.WrapReqAndClaims[ArticleReq](a.Withdraw))

	g.GET("/detail/:id", a.Detail)
	g.POST("/list", a.List)

	pub := g.Group("/pub")
	pub.GET("/:id", ginx.WrapClaims(a.PubDetail))
	pub.POST("/like", ginx.WrapReqAndClaims[LikeReq](a.Like))
	pub.POST("/collect", ginx.WrapReqAndClaims[CollectReq](a.Collect))
}

func (a *ArticleHandler) Edit(ctx *gin.Context, req ArticleReq, user jwt.UserClaims) (Result, error) {
	id, err := a.svc.Save(ctx, req.toDomain(user.Id))
	if err != nil {
		return Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return Result{
		Data: id,
	}, nil
}

func (a *ArticleHandler) Publish(ctx *gin.Context, req ArticleReq, user jwt.UserClaims) (Result, error) {
	id, err := a.svc.Publish(ctx, req.toDomain(user.Id))
	if err != nil {
		return Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return Result{
		Data: id,
	}, err
}

func (a *ArticleHandler) Withdraw(ctx *gin.Context, req ArticleReq, user jwt.UserClaims) (Result, error) {
	if err := a.svc.Withdraw(ctx, user.Id, req.Id); err != nil {
		a.l.Error("设置为仅可自己可见失败", logger.Error(err), logger.Field{Key: "id", Value: req.Id})
		return Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		}, nil
	}
	return Result{
		Msg: "OK",
	}, nil
}

func (a *ArticleHandler) Detail(ctx *gin.Context) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: errs.UserInvalidInput,
			Msg:  "参数错误",
		})
		a.l.Error("前端输入的 ID 不对", logger.Error(err))
		return
	}
	user, ok := ctx.MustGet("user").(jwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		})
		a.l.Error("获取用户会话信息失败")
		return
	}
	art, err := a.svc.GetById(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		})
		a.l.Error("获得文章信息失败", logger.Error(err))
		return
	}
	if art.Author.Id != user.Id {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "输入有误",
		})
		a.l.Error("非法访问文章，创作者ID不匹配", logger.Int64("uid", user.Id))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: ArticleVo{
			Id:      art.Id,
			Title:   art.Title,
			Status:  art.Status.ToUint8(),
			Content: art.Content,
			Ctime:   art.Ctime.Format(time.DateTime),
			Utime:   art.Utime.Format(time.DateTime),
		},
	})
	return
}

func (a *ArticleHandler) List(ctx *gin.Context) {
	type Req struct {
		Offset int `json:"offset"`
		Limit  int `json:"limit"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		a.l.Error("反序列化请求失败", logger.Error(err))
		return
	}
	if req.Limit > 100 {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "请求有误",
		})
		a.l.Error("获得用户会话信息失败")
		return
	}
	user, ok := ctx.MustGet("user").(jwt.UserClaims)
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		})
		a.l.Error("获得用户会话信息失败")
		return
	}
	arts, err := a.svc.List(ctx, user.Id, req.Offset, req.Limit)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		})
		a.l.Error("获得用户会话信息失败")
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: slice.Map[domain.Article, ArticleVo](arts,
			func(idx int, src domain.Article) ArticleVo {
				return ArticleVo{
					Id:       src.Id,
					Title:    src.Title,
					Abstract: src.Abstract(),
					Status:   src.Status.ToUint8(),
					Ctime:    src.Ctime.Format(time.DateTime),
					Utime:    src.Utime.Format(time.DateTime),
				}
			}),
	})
}

func (a *ArticleHandler) PubDetail(ctx *gin.Context, uc ginx.UserClaims) (Result, error) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		a.l.Error("前端输入的 ID 不对", logger.Error(err))
		return Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "参数错误",
		}, fmt.Errorf("查询文章详情的 ID %s 不正确, %w", idStr, err)
	}
	var (
		eg       errgroup.Group
		art      domain.Article
		intrResp intrDomain.Interactive
	)
	eg.Go(func() error {
		var er error
		art, er = a.svc.GetPublishedById(ctx, id, uc.Id)
		return er
	})

	eg.Go(func() error {
		var er error
		intrResp, er = a.intrSvc.Get(ctx, a.biz, id, uc.Id)
		return er
	})
	err = eg.Wait()
	if err != nil {
		return Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		}, fmt.Errorf("获取文章信息失败 %w", err)
	}
	return Result{
		Data: ArticleVo{
			Id:         art.Id,
			Title:      art.Title,
			Status:     art.Status.ToUint8(),
			Content:    art.Content,
			Author:     art.Author.Name,
			Ctime:      art.Ctime.Format(time.DateTime),
			Utime:      art.Utime.Format(time.DateTime),
			ReadCnt:    intrResp.ReadCnt,
			CollectCnt: intrResp.CollectCnt,
			LikeCnt:    intrResp.LikeCnt,
			Liked:      intrResp.Liked,
			Collected:  intrResp.Collected,
		},
	}, nil
}

func (a *ArticleHandler) Like(ctx *gin.Context, req LikeReq, uc jwt.UserClaims) (ginx.Result, error) {
	var err error
	if req.Like {
		err = a.intrSvc.Like(ctx, a.biz, req.Id, uc.Id)
	} else {
		err = a.intrSvc.CancelLike(ctx, a.biz, req.Id, uc.Id)
	}
	if err != nil {
		return Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return Result{Msg: "OK"}, nil
}

func (a *ArticleHandler) Collect(ctx *gin.Context, req CollectReq, uc jwt.UserClaims) (Result, error) {
	if err := a.intrSvc.Collect(ctx, a.biz, req.Id, req.Cid, uc.Id); err != nil {
		return Result{
			Code: errs.ArticleInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	return Result{Msg: "OK"}, nil
}
