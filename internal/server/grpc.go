package server

import (
	v1 "kratos-realworld/api/realworld/v1"
	"kratos-realworld/internal/conf"
	"kratos-realworld/internal/service"

	"context"

	"github.com/go-kratos/kratos/v2/log"
	kjwt "github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/golang-jwt/jwt/v5"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server, a *conf.Auth, realworld *service.RealWorldService, logger log.Logger) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			selector.Server(kjwt.Server(func(token *jwt.Token) (interface{}, error) {
				return []byte(a.JwtSecret), nil
			})).Match(func(ctx context.Context, operation string) bool {
				// 这里返回 true 表示需要 JWT 鉴权
				// 登录和注册接口跳过鉴权
				return operation != "/realworld.v1.RealWorld/Registrat" && operation != "/realworld.v1.RealWorld/Login"
			}).Build(),
		),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	v1.RegisterRealWorldServer(srv, realworld)
	return srv
}
