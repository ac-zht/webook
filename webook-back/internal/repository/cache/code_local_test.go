package cache

import (
	"context"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/magiconair/properties/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestLocalCodeCache_Set(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func() *lru.Cache[string, CodeItem]
		biz     string
		phone   string
		code    string
		wantErr error
	}{
		{
			name: "设置成功",
			mock: func() *lru.Cache[string, CodeItem] {
				l, err := lru.New[string, CodeItem](10)
				require.NoError(t, err)
				return l
			},
			biz:   "login",
			phone: "10086",
			code:  "123456",
		},
		{
			name: "1分钟内已设置过",
			mock: func() *lru.Cache[string, CodeItem] {
				l, err := lru.New[string, CodeItem](10)
				require.NoError(t, err)
				l.Add("phone_code:login:10086", CodeItem{
					code: "123456",
					cnt:  3,
					//还有4分钟1秒过期
					deadline: time.Now().Add(time.Minute*4 + time.Second),
				})
				return l
			},
			biz:     "login",
			phone:   "10086",
			code:    "123456",
			wantErr: ErrCodeSendTooMany,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache := NewLocalCodeCache(tc.mock(), time.Minute*5)
			err := cache.Set(context.Background(), tc.biz, tc.phone, tc.code)
			assert.Equal(t, err, tc.wantErr)
		})
	}
}

func TestLocalCodeCache_Verify(t *testing.T) {
	testCases := []struct {
		name         string
		mock         func() *lru.Cache[string, CodeItem]
		biz          string
		phone        string
		inputCode    string
		wantVerified bool
		wantErr      error
	}{
		{
			name: "验证成功",
			mock: func() *lru.Cache[string, CodeItem] {
				l, err := lru.New[string, CodeItem](10)
				require.NoError(t, err)
				l.Add("phone_code:login:10086", CodeItem{
					code:     "123456",
					cnt:      3,
					deadline: time.Now().Add(time.Minute * 5),
				})
				return l
			},
			biz:          "login",
			phone:        "10086",
			inputCode:    "123456",
			wantVerified: true,
		},
		{
			name: "验证失败，code不一致",
			mock: func() *lru.Cache[string, CodeItem] {
				l, err := lru.New[string, CodeItem](10)
				require.NoError(t, err)
				l.Add("phone_code:login:10086", CodeItem{
					code:     "123457",
					cnt:      3,
					deadline: time.Now().Add(time.Minute * 5),
				})
				return l
			},
			biz:          "login",
			phone:        "10086",
			inputCode:    "123456",
			wantVerified: false,
		},
		{
			name: "验证失败，未发送过验证码",
			mock: func() *lru.Cache[string, CodeItem] {
				l, err := lru.New[string, CodeItem](10)
				require.NoError(t, err)
				return l
			},
			biz:          "login",
			phone:        "10086",
			inputCode:    "123456",
			wantVerified: false,
		},
		{
			name: "验证失败，验证次数过多",
			mock: func() *lru.Cache[string, CodeItem] {
				l, err := lru.New[string, CodeItem](10)
				require.NoError(t, err)
				l.Add("phone_code:login:10086", CodeItem{
					code:     "123456",
					cnt:      0,
					deadline: time.Now().Add(time.Minute * 5),
				})
				return l
			},
			biz:          "login",
			phone:        "10086",
			inputCode:    "123456",
			wantVerified: false,
			wantErr:      ErrCodeVerifyTooManyTimes,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache := NewLocalCodeCache(tc.mock(), time.Minute*5)
			verified, err := cache.Verify(context.Background(), tc.biz, tc.phone, tc.inputCode)
			assert.Equal(t, verified, tc.wantVerified)
			assert.Equal(t, err, tc.wantErr)
		})
	}
}
