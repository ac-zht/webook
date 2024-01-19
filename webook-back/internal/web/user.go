package web

import (
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"github.com/zht-account/webook/internal/domain"
	"github.com/zht-account/webook/internal/service"
	ijwt "github.com/zht-account/webook/internal/web/jwt"
	"github.com/zht-account/webook/pkg/ginx"
	"go.uber.org/zap"
	"net/http"
	"time"
)

var _ handler = &UserHandler{}

const (
	emailRegexPattern    = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	bizLogin             = "login"
)

type UserHandler struct {
	svc              service.UserService
	codeSvc          service.CodeService
	emailRegexExp    *regexp.Regexp
	passwordRegexExp *regexp.Regexp
	ijwt.Handler
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService, jwtHdl ijwt.Handler) *UserHandler {
	return &UserHandler{
		svc:              svc,
		codeSvc:          codeSvc,
		emailRegexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		Handler:          jwtHdl,
	}
}

func (c *UserHandler) LoginJWT(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `form:"email" json:"email"`
		Password string `form:"password" json:"password"`
	}
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{Msg: "参数错误"})
		return
	}
	isEmail, err := c.emailRegexExp.MatchString(req.Email)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{Msg: "系统错误"})
		return
	}
	if !isEmail {
		ctx.JSON(http.StatusOK, ginx.Result{Msg: "邮箱格式不正确"})
		return
	}
	isPassword, err := c.passwordRegexExp.MatchString(req.Password)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{Msg: "系统错误"})
		return
	}
	if !isPassword {
		ctx.JSON(http.StatusOK, ginx.Result{Msg: "密码必须包含数字、特殊字符，并且长度不能小于 8 位"})
		return
	}
	u, err := c.svc.Login(ctx, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.JSON(http.StatusOK, ginx.Result{Msg: "用户名或密码错误"})
		return
	}
	if c.SetLoginToken(ctx, u.Id) != nil {
		ctx.JSON(http.StatusOK, ginx.Result{Msg: "系统异常"})
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{Msg: "登录成功"})
}

func (c *UserHandler) SendSMSLoginCode(ctx *gin.Context) {
	type Req struct {
		Phone string `form:"phone" json:"phone"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	if req.Phone == "" {
		ctx.JSON(http.StatusOK, ginx.Result{Code: 4, Msg: "请输入手机号码"})
		return
	}
	err := c.codeSvc.Send(ctx, bizLogin, req.Phone)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, ginx.Result{Msg: "发送成功"})
	case service.ErrCodeSendTooMany:
		ctx.JSON(http.StatusOK, ginx.Result{Code: 4, Msg: "短信发送太频繁，请稍后再试"})
	default:
		ctx.JSON(http.StatusOK, ginx.Result{Code: 5, Msg: "系统错误"})
		return
	}
}

func (c *UserHandler) LoginSMS(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	ok, err := c.codeSvc.Verify(ctx, bizLogin, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{Code: 5, Msg: "系统异常"})
		zap.L().Error("用户手机号码登录失败", zap.Error(err))
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, ginx.Result{Code: 4, Msg: "验证码错误"})
		return
	}
	u, err := c.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{Code: 4, Msg: "系统错误"})
		return
	}
	if c.SetLoginToken(ctx, u.Id) != nil {
		ctx.JSON(http.StatusOK, ginx.Result{Msg: "系统异常"})
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{Msg: "登录成功"})
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
		ctx.String(http.StatusOK, "密码必须包含数字、特殊字符，并且长度不能小于 8 位")
		return
	}
	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次输入的密码不相同")
		return
	}
	err = c.svc.Signup(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
		Ctime:    time.Now(),
	})
	if err == service.ErrUserDuplicateEmail {
		ctx.String(http.StatusOK, "邮箱冲突")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统异常")
		return
	}
	ctx.String(http.StatusOK, "注册成功")
}

func (c *UserHandler) Edit(ctx *gin.Context) {
	type Request struct {
		Nickname string `form:"nickname" json:"nickname"`
		Birthday string `form:"birthday" json:"birthday"`
		AboutMe  string `form:"aboutMe" json:"aboutMe"`
	}
	var req Request
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{Code: 4, Msg: "参数错误"})
		return
	}
	if req.Nickname == "" {
		ctx.JSON(http.StatusOK, ginx.Result{Code: 4, Msg: "昵称不能为空"})
		return
	}
	if len(req.AboutMe) > 1024 {
		ctx.JSON(http.StatusOK, ginx.Result{Code: 4, Msg: "简介内容过长"})
	}
	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{Code: 4, Msg: "日期格式错误"})
		return
	}
	uc := ctx.MustGet("user").(ginx.UserClaims)
	err = c.svc.UpdateNonSensitiveInfo(ctx, domain.User{
		Id:       uc.Id,
		Nickname: req.Nickname,
		AboutMe:  req.AboutMe,
		Birthday: birthday,
	})
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{Code: 5, Msg: "系统错误"})
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{Msg: "成功"})
}

func (c *UserHandler) ProfileJWT(ctx *gin.Context) {
	type Profile struct {
		Email    string
		Phone    string
		Nickname string
		Birthday string
		AboutMe  string
	}
	uc := ctx.MustGet("user").(ginx.UserClaims)
	u, err := c.svc.Profile(ctx, uc.Id)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.JSON(http.StatusOK, Profile{
		Email:    u.Email,
		Phone:    u.Phone,
		Nickname: u.Nickname,
		Birthday: u.Birthday.Format(time.DateOnly),
		AboutMe:  u.AboutMe,
	})
}

func (c *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("/signup", c.SignUp)
	ug.POST("/login", c.LoginJWT)
	ug.POST("/edit", c.Edit)
	ug.GET("/profile", c.ProfileJWT)
	ug.POST("/login_sms/code/send", c.SendSMSLoginCode)
	ug.POST("/login_sms", c.LoginSMS)
}
