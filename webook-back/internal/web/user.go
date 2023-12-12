package web

import (
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"github.com/zht-account/webook/internal/domain"
	"github.com/zht-account/webook/internal/service"
	"net/http"
	"time"
)

var _ handler = &UserHandler{}

const (
	emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	// 和上面比起来，用 ` 看起来就比较清爽
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
)

type UserHandler struct {
	srv              *service.UserService
	emailRegexExp    *regexp.Regexp
	passwordRegexExp *regexp.Regexp
}

func NewUserHandler(srv *service.UserService) *UserHandler {
	return &UserHandler{
		srv:              srv,
		emailRegexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
	}
}

func (c *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `form:"email" json:"email"`
		Password string `form:"password" json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	isEmail, err := c.emailRegexExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isEmail {
		ctx.String(http.StatusOK, "邮箱不正确")
		return
	}
	isPassword, err := c.passwordRegexExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isPassword {
		ctx.String(http.StatusOK, "密码格式不正确")
		return
	}
	_, err = c.srv.Login(ctx, req.Email, req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "账号信息错误")
		return
	}
	ctx.String(http.StatusOK, "登录成功")
}

func (c *UserHandler) SignUp(ctx *gin.Context) {
	type SingUpReq struct {
		Email           string `form:"email" json:"email"`
		Password        string `form:"password" json:"password"`
		ConfirmPassword string `form:"confirmPassword" json:"confirmPassword"`
	}
	var req SingUpReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	isEmail, err := c.emailRegexExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isEmail {
		ctx.String(http.StatusOK, "邮箱不正确")
		return
	}
	isPassword, err := c.passwordRegexExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isPassword {
		ctx.String(http.StatusOK, "密码格式不正确")
		return
	}
	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次输入的密码不相同")
		return
	}

	err = c.srv.Signup(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
		Ctime:    time.Now(),
	})
	if err != nil {
		ctx.String(http.StatusOK, "注册失败")
		return
	}
	ctx.String(http.StatusOK, "注册成功")
}

func (c *UserHandler) Edit(ctx *gin.Context) {
}

func (c *UserHandler) Profile(ctx *gin.Context) {
}

func (c *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("/signup", c.SignUp)
	ug.POST("/login", c.Login)
	ug.POST("/edit", c.Edit)
	ug.POST("/profile", c.Profile)
}
