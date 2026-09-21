package userlogic

import (
	"rpc-user/pkg/code"
	"context"

	"rpc-user/internal/svc"
	"rpc-user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminUserListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminUserListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUserListLogic {
	return &AdminUserListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AdminUserListLogic) AdminUserList(in *user.AdminUserListRequest) (resp *user.AdminUserListResponse, err error) {
	resp = new(user.AdminUserListResponse)
	resp.Data = new(user.AdminUserListData)
	resp.Data.Users = make([]*user.UserInfoData, 0)

	users, err := l.svcCtx.UserModel.FindAll(l.ctx, in.Keyword, in.Cursor, in.PageSize)
	if err != nil {
		resp.Code = int64(code.ServerErr.Code())
		resp.Msg = err.Error()
		return resp, nil
	}

	var isEnd bool
	if len(users) < int(in.PageSize) {
		isEnd = true
	}

	nextCursor := in.Cursor
	// 转换用户数据模型为响应 DTO
	for _, v := range users {
		resp.Data.Users = append(resp.Data.Users, &user.UserInfoData{
			UserId:       int64(v.Id),
			Username:     v.Username,
			Mobile:       v.Mobile,
			Avatar:       v.Avatar,
			Role:         v.Role,
			DisplayId:    v.DisplayId,
			Bio:          v.Bio,
			Gender:       int32(v.Gender),
			ProfileCover: v.ProfileCover,
		})
		nextCursor = int64(v.Id)
	}

	resp.Data.Cursor = nextCursor
	resp.Data.IsEnd = isEnd

	return resp, nil
}
