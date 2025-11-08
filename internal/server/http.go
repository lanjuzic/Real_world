package server

import (
	"context"
	v1 "kratos-realworld/api/realworld/v1"
	"kratos-realworld/internal/conf"
	"kratos-realworld/internal/service"

	myjwt "kratos-realworld/internal/pkg/jwt"

	"github.com/go-kratos/kratos/v2/log"
	kjwt "github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/golang-jwt/jwt/v5"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, a *conf.Auth, realworld *service.RealWorldService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			selector.Server(
				kjwt.Server(
					func(token *jwt.Token) (interface{}, error) {
						return []byte(a.JwtSecret), nil
					},
					// ✅ 指定使用你的自定义 claims
					kjwt.WithClaims(func() jwt.Claims {
						return &myjwt.CustomClaims{}
					}),
				),
			).Match(func(ctx context.Context, operation string) bool {
				// 这里返回 true 表示需要 JWT 鉴权
				// 登录和注册接口跳过鉴权
				return operation != "/realworld.v1.RealWorld/Login" && operation != "/realworld.v1.RealWorld/Register"
			}).Build(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	v1.RegisterRealWorldHTTPServer(srv, realworld)
	return srv
}
