package server

import (
	"context"

	userlogic "rpc-user/internal/logic/user"
	"rpc-user/internal/svc"
	pbuser "rpc-user/pb/user"
	"rpc-user/user"
)

type LegacyUserServer struct {
	svcCtx *svc.ServiceContext
	pbuser.UnimplementedUserServer
}

func NewLegacyUserServer(svcCtx *svc.ServiceContext) *LegacyUserServer {
	return &LegacyUserServer{
		svcCtx: svcCtx,
	}
}

func (s *LegacyUserServer) Register(ctx context.Context, in *pbuser.RegisterRequest) (*pbuser.RegisterResponse, error) {
	l := userlogic.NewRegisterLogic(ctx, s.svcCtx)
	resp, err := l.Register(&user.RegisterRequest{
		Username: in.Username,
		Mobile:   in.Mobile,
		Avatar:   in.Avatar,
		Password: in.Password,
	})
	if err != nil {
		return nil, err
	}
	var uid int64
	if resp != nil && resp.Data != nil {
		uid = resp.Data.UserId
	}
	return &pbuser.RegisterResponse{UserId: uid}, nil
}

func (s *LegacyUserServer) FindById(ctx context.Context, in *pbuser.FindByIdRequest) (*pbuser.FindByIdResponse, error) {
	l := userlogic.NewFindByIdLogic(ctx, s.svcCtx)
	resp, err := l.FindById(&user.FindByIdRequest{UserId: in.UserId})
	if err != nil {
		return nil, err
	}
	ret := new(pbuser.FindByIdResponse)
	if resp != nil && resp.Data != nil {
		ret.UserId = resp.Data.UserId
		ret.Username = resp.Data.Username
		ret.Mobile = resp.Data.Mobile
		ret.Avatar = resp.Data.Avatar
		ret.Role = resp.Data.Role
		ret.DisplayId = resp.Data.DisplayId
		ret.Bio = resp.Data.Bio
		ret.Gender = resp.Data.Gender
		ret.ProfileCover = resp.Data.ProfileCover
	}
	return ret, nil
}

func (s *LegacyUserServer) FindByMobile(ctx context.Context, in *pbuser.FindByMobileRequest) (*pbuser.FindByMobileResponse, error) {
	l := userlogic.NewFindByMobileLogic(ctx, s.svcCtx)
	resp, err := l.FindByMobile(&user.FindByMobileRequest{Mobile: in.Mobile})
	if err != nil {
		return nil, err
	}
	ret := new(pbuser.FindByMobileResponse)
	if resp != nil && resp.Data != nil {
		ret.UserId = resp.Data.UserId
		ret.Username = resp.Data.Username
		ret.Mobile = resp.Data.Mobile
		ret.Avatar = resp.Data.Avatar
		ret.Password = resp.Data.Password
		ret.Role = resp.Data.Role
	}
	return ret, nil
}

func (s *LegacyUserServer) SendSms(ctx context.Context, in *pbuser.SendSmsRequest) (*pbuser.SendSmsResponse, error) {
	l := userlogic.NewSendSmsLogic(ctx, s.svcCtx)
	_, err := l.SendSms(&user.SendSmsRequest{Mobile: in.Mobile})
	if err != nil {
		return nil, err
	}
	return &pbuser.SendSmsResponse{}, nil
}

func (s *LegacyUserServer) UpdateProfile(ctx context.Context, in *pbuser.UpdateProfileRequest) (*pbuser.UpdateProfileResponse, error) {
	l := userlogic.NewUpdateProfileLogic(ctx, s.svcCtx)
	_, err := l.UpdateProfile(&user.UpdateProfileRequest{
		UserId:       in.UserId,
		Username:     in.Username,
		Avatar:       in.Avatar,
		Gender:       in.Gender,
		Bio:          in.Bio,
		ProfileCover: in.ProfileCover,
	})
	if err != nil {
		return nil, err
	}
	return &pbuser.UpdateProfileResponse{}, nil
}

func (s *LegacyUserServer) UpgradePassword(ctx context.Context, in *pbuser.UpgradePasswordRequest) (*pbuser.UpgradePasswordResponse, error) {
	l := userlogic.NewUpgradePasswordLogic(ctx, s.svcCtx)
	_, err := l.UpgradePassword(&user.UpgradePasswordRequest{
		UserId:       in.UserId,
		PasswordHash: in.PasswordHash,
	})
	if err != nil {
		return nil, err
	}
	return &pbuser.UpgradePasswordResponse{}, nil
}

func (s *LegacyUserServer) AdminUserList(ctx context.Context, in *pbuser.AdminUserListRequest) (*pbuser.AdminUserListResponse, error) {
	l := userlogic.NewAdminUserListLogic(ctx, s.svcCtx)
	resp, err := l.AdminUserList(&user.AdminUserListRequest{
		Cursor:   in.Cursor,
		PageSize: in.PageSize,
		Keyword:  in.Keyword,
	})
	if err != nil {
		return nil, err
	}
	ret := new(pbuser.AdminUserListResponse)
	if resp != nil && resp.Data != nil {
		ret.Cursor = resp.Data.Cursor
		ret.IsEnd = resp.Data.IsEnd
		for _, u := range resp.Data.Users {
			if u != nil {
				ret.Items = append(ret.Items, &pbuser.AdminUserItem{
					UserId:       u.UserId,
					Username:     u.Username,
					Mobile:       u.Mobile,
					Avatar:       u.Avatar,
					Role:         u.Role,
					DisplayId:    u.DisplayId,
					Bio:          u.Bio,
					Gender:       u.Gender,
					ProfileCover: u.ProfileCover,
				})
			}
		}
	}
	return ret, nil
}
