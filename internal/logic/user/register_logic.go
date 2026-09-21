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
		return encrypt.MD5Password(password)
	}
	return hash
}

func (l *RegisterLogic) Register(in *user.RegisterRequest) (resp *user.RegisterResponse, err error) {
	resp = new(user.RegisterResponse)
	resp.Code = 200
	resp.Msg = "success"
	resp.Data = new(user.RegisterData)

	if len(in.Username) == 0 {
		resp.Code = int64(code.RegisterNameEmpty.Code())
		resp.Msg = code.RegisterNameEmpty.Message()
		resp.Data = nil
		return resp, nil
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
		resp.Code = 500
		resp.Msg = err.Error()
		resp.Data = nil
		return resp, nil
	}
	userId, err := ret.LastInsertId()
	if err != nil {
		resp.Code = 500
		resp.Msg = err.Error()
		resp.Data = nil
		return resp, nil
	}

	resp.Data.UserId = userId
	return resp, nil
}
