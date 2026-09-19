package memberrpc

import (
	"rpc-user/member/internal/config"
	"rpc-user/member/internal/server"
	"rpc-user/member/internal/svc"
	"rpc-user/member/pb"

	"google.golang.org/grpc"
)

type Config = config.Config

func Register(grpcServer *grpc.Server, c Config) {
	ctx := svc.NewServiceContext(c)
	pb.RegisterMemberServer(grpcServer, server.NewMemberServer(ctx))
}
