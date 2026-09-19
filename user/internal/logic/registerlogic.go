package logic

import (
	"rpc-user/user/internal/code"
	"rpc-user/user/internal/model"
	"context"
	"time"

	"rpc-user/user/internal/svc"
	"rpc-user/user/service"
	"rpc-user/pkg/encrypt"
	"rpc-user/pkg/util"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func hashPassword(password string) string {
	hash, err := encrypt.HashPassword(password)
	if err != nil {
		logx.Errorf("HashPassword error: %v", err)
		return encrypt.MD5Password(password)
	}
	return hash
}

func (l *RegisterLogic) Register(in *service.RegisterRequest) (*service.RegisterResponse, error) {
	// 当注册名字为空的时候，返回业务自定义错误码
	if len(in.Username) == 0 {
		return nil, code.RegisterNameEmpty
	}

	ret, err := l.svcCtx.UserModel.Insert(l.ctx, &model.User{
		Username:   in.Username,
		Mobile:     in.Mobile,
		Avatar:     in.Avatar,
		Password:   hashPassword(in.Password),
		DisplayId:  util.GenerateCode(8), // Generate unique display_id
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	})
	if err != nil {
		logx.Errorf("Register req: %v error: %v", in, err)
		return nil, err
	}
	userId, err := ret.LastInsertId()
	if err != nil {
		logx.Errorf("LastInsertId error: %v", err)
		return nil, err
	}

	return &service.RegisterResponse{UserId: userId}, nil
}
