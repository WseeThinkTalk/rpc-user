package main

import (
	"flag"
	"fmt"

	followrpc "rpc-user/follow"
	memberrpc "rpc-user/member"
	"rpc-user/pkg/env"
	"rpc-user/pkg/interceptors"
	userrpc "rpc-user/user"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/user.yaml", "the config file")

func main() {
	flag.Parse()

	env.LoadEnv()

	var c userrpc.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	if c.DB.DataSource == "" {
		c.DB.DataSource = c.DataSource
	}

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		userrpc.Register(grpcServer, c)
		memberrpc.Register(grpcServer, memberrpc.Config{
			RpcServerConf: c.RpcServerConf,
			DB:            c.DB,
			BizRedis:      c.BizRedis,
		})
		followrpc.Register(grpcServer, followrpc.Config{
			RpcServerConf: c.RpcServerConf,
			DB:            c.DB,
			BizRedis:      c.BizRedis,
			KqPusherConf:  c.KqPusherConf,
		})

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	s.AddUnaryInterceptors(interceptors.ServerErrorInterceptor())
	defer s.Stop()

	fmt.Printf("Starting unified user rpc server (user, member, follow) at %s...\n", c.ListenOn)
	s.Start()
}
