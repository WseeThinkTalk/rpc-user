package userlogic

import (
	"context"
	"time"

	model "rpc-user/internal/model/user"
	"rpc-user/internal/svc"
	"rpc-user/pkg/code"
	"rpc-user/pkg/encrypt"
	"rpc-user/pkg/util"
	"rpc-user/user"

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

func (l *RegisterLogic) Register(in *user.RegisterRequest) (*user.RegisterResponse, error) {
	if len(in.Username) == 0 {
		return nil, code.RegisterNameEmpty
	}

	ret, err := l.svcCtx.UserModel.Insert(l.ctx, &model.User{
		Username:   in.Username,
		Mobile:     in.Mobile,
		Avatar:     in.Avatar,
		Password:   hashPassword(in.Password),
		DisplayId:  util.GenerateCode(8),
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

	return &user.RegisterResponse{
		Code: 200,
		Msg:  "success",
		Data: &user.RegisterData{
			UserId: userId,
		},
	}, nil
}
