package web

import (
	"bytes"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zht-account/webook/internal/domain"
	"github.com/zht-account/webook/internal/service"
	svcmocks "github.com/zht-account/webook/internal/service/mocks"
	ijwt "github.com/zht-account/webook/internal/web/jwt"
	jwtmocks "github.com/zht-account/webook/internal/web/jwt/mocks"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEmailPattern(t *testing.T) {
	testCases := []struct {
		name  string
		email string
		match bool
	}{
		{
			name:  "合法邮箱",
			email: "123@qq.com",
			match: true,
		},
		{
			name:  "没有@",
			email: "123",
			match: false,
		},
		{
			name:  "有@但没后缀",
			email: "123@",
			match: false,
		},
	}
	hdl := NewUserHandler(nil, nil, nil)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			match, err := hdl.emailRegexExp.MatchString(tc.email)
			require.NoError(t, err)
			assert.Equal(t, tc.match, match)
		})
	}
}

func TestPasswordPattern(t *testing.T) {
	testCases := []struct {
		name     string
		password string
		match    bool
	}{
		{
			name:     "合法密码格式",
			password: "hello@world123",
			match:    true,
		},
		{
			name:     "缺少特殊字符",
			password: "world1234",
			match:    false,
		},
		{
			name:     "缺少数字",
			password: "hello@world",
			match:    false,
		},
		{
			name:     "缺少字母",
			password: "123@45678",
			match:    false,
		},
		{
			name:     "长度小于8",
			password: "1@a",
			match:    false,
		},
	}
	hdl := NewUserHandler(nil, nil, nil)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			match, err := hdl.passwordRegexExp.MatchString(tc.password)
			require.NoError(t, err)
			assert.Equal(t, tc.match, match)
		})
	}
}

func TestUserHandler_SignUp(t *testing.T) {
	const signupUrl = "/users/signup"
	testCases := []struct {
		name       string
		mock       func(ctrl *gomock.Controller) (service.UserService, service.CodeService, ijwt.Handler)
		reqBuilder func(t *testing.T) *http.Request

		wantCode int
		wantBody string
	}{
		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, ijwt.Handler) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().Signup(gomock.Any(), gomock.Any()).Return(nil)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc, nil
			},
			reqBuilder: func(t *testing.T) *http.Request {
				body := bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello@world123","confirmPassword":"hello@world123"}`))
				req, err := http.NewRequest(http.MethodPost, signupUrl, body)
				req.Header.Set("Content-Type", "application/json")
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantCode: http.StatusOK,
			wantBody: "注册成功",
		},
		{
			name: "参数错误",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, ijwt.Handler) {
				return nil, nil, nil
			},
			reqBuilder: func(t *testing.T) *http.Request {
				body := bytes.NewBuffer([]byte(`{"email":"123@qq.com",}`))
				req, err := http.NewRequest(http.MethodPost, signupUrl, body)
				req.Header.Set("Content-Type", "application/json")
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "邮箱格式错误",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, ijwt.Handler) {
				return nil, nil, nil
			},
			reqBuilder: func(t *testing.T) *http.Request {
				body := bytes.NewBuffer([]byte(`{"email":"123"}`))
				req, err := http.NewRequest(http.MethodPost, signupUrl, body)
				req.Header.Set("Content-Type", "application/json")
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantCode: http.StatusOK,
			wantBody: "邮箱不正确",
		},
		{
			name: "密码格式错误",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, ijwt.Handler) {
				return nil, nil, nil
			},
			reqBuilder: func(t *testing.T) *http.Request {
				body := bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"123"}`))
				req, err := http.NewRequest(http.MethodPost, signupUrl, body)
				req.Header.Set("Content-Type", "application/json")
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantCode: http.StatusOK,
			wantBody: "密码必须包含数字、特殊字符，并且长度不能小于 8 位",
		},
		{
			name: "两次输入的密码不相同",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, ijwt.Handler) {
				return nil, nil, nil
			},
			reqBuilder: func(t *testing.T) *http.Request {
				body := bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello@world123","confirmPassword":"hello@world1234"}`))
				req, err := http.NewRequest(http.MethodPost, signupUrl, body)
				req.Header.Set("Content-Type", "application/json")
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantCode: http.StatusOK,
			wantBody: "两次输入的密码不相同",
		},
		{
			name: "邮箱冲突",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, ijwt.Handler) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().Signup(gomock.Any(), gomock.Any()).Return(service.ErrUserDuplicateEmail)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc, nil
			},
			reqBuilder: func(t *testing.T) *http.Request {
				body := bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello@world123","confirmPassword":"hello@world123"}`))
				req, err := http.NewRequest(http.MethodPost, signupUrl, body)
				req.Header.Set("Content-Type", "application/json")
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantCode: http.StatusOK,
			wantBody: "邮箱冲突",
		},
		{
			name: "系统异常",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, ijwt.Handler) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().Signup(gomock.Any(), gomock.Any()).Return(errors.New("this is a error"))
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc, nil
			},
			reqBuilder: func(t *testing.T) *http.Request {
				body := bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello@world123","confirmPassword":"hello@world123"}`))
				req, err := http.NewRequest(http.MethodPost, signupUrl, body)
				req.Header.Set("Content-Type", "application/json")
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantCode: http.StatusOK,
			wantBody: "系统异常",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			userSvc, codeSvc, jwtHdl := tc.mock(ctrl)
			hdl := NewUserHandler(userSvc, codeSvc, jwtHdl)
			gin.SetMode(gin.ReleaseMode)
			server := gin.Default()
			hdl.RegisterRoutes(server)
			req := tc.reqBuilder(t)
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, req)

			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantBody, recorder.Body.String())
		})
	}
}

func TestUserHandler_LoginJWT(t *testing.T) {
	const loginUrl = "/users/login"
	testCases := []struct {
		name       string
		mock       func(ctrl *gomock.Controller) (service.UserService, service.CodeService, ijwt.Handler)
		reqBuilder func(t *testing.T) *http.Request

		wantCode int
		wantBody string
	}{
		{
			name: "登录成功",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, ijwt.Handler) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().Login(gomock.Any(), gomock.Any(), gomock.Any()).Return(domain.User{}, nil)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				jwtHdl := jwtmocks.NewMockHandler(ctrl)
				jwtHdl.EXPECT().SetLoginToken(gomock.Any(), gomock.Any()).Return(nil)
				return userSvc, codeSvc, jwtHdl
			},
			reqBuilder: func(t *testing.T) *http.Request {
				body := bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello@world123"}`))
				req, err := http.NewRequest(http.MethodPost, loginUrl, body)
				req.Header.Set("Content-Type", "application/json")
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantCode: http.StatusOK,
			wantBody: `{"code":0,"msg":"登录成功","data":null}`,
		},
		{
			name: "用户名或密码错误",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, ijwt.Handler) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().Login(gomock.Any(), gomock.Any(), gomock.Any()).Return(domain.User{}, service.ErrInvalidUserOrPassword)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc, nil
			},
			reqBuilder: func(t *testing.T) *http.Request {
				body := bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello@world123"}`))
				req, err := http.NewRequest(http.MethodPost, loginUrl, body)
				req.Header.Set("Content-Type", "application/json")
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantCode: http.StatusOK,
			wantBody: `{"code":0,"msg":"用户名或密码错误","data":null}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			userSvc, codeSvc, jwtHdl := tc.mock(ctrl)
			hdl := NewUserHandler(userSvc, codeSvc, jwtHdl)
			gin.SetMode(gin.ReleaseMode)
			server := gin.Default()
			hdl.RegisterRoutes(server)
			req := tc.reqBuilder(t)
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, req)

			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantBody, recorder.Body.String())
		})
	}
}

func TestUserHandler_Edit(t *testing.T) {
	const editUrl = "/users/edit"
	testCases := []struct {
		name       string
		mock       func(ctrl *gomock.Controller) (service.UserService, service.CodeService, ijwt.Handler)
		reqBuilder func(t *testing.T) *http.Request

		wantCode int
		wantBody string
	}{
		{
			name: "编辑成功",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, ijwt.Handler) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().UpdateNonSensitiveInfo(gomock.Any(), gomock.Any()).Return(nil)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc, nil
			},
			reqBuilder: func(t *testing.T) *http.Request {
				body := bytes.NewBuffer([]byte(`{"nickname":"user","birthday":"2000-04-12","aboutMe":"info"}`))
				req, err := http.NewRequest(http.MethodPost, editUrl, body)
				req.Header.Set("Content-Type", "application/json")
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantCode: http.StatusOK,
			wantBody: `{"code":0,"msg":"成功","data":null}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			userSvc, codeSvc, jwtHdl := tc.mock(ctrl)
			hdl := NewUserHandler(userSvc, codeSvc, jwtHdl)
			gin.SetMode(gin.TestMode)
			server := gin.Default()
			hdl.RegisterRoutes(server)
			req := tc.reqBuilder(t)
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Set("user", ijwt.UserClaims{})
			ctx.Request = req
			server.HandleContext(ctx)
			//server.ServeHTTP(recorder, req)
			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantBody, recorder.Body.String())
		})
	}
}
