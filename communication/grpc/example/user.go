package example

import (
	"context"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
	. "github.com/jinuopti/lilpop-server/log"
	"github.com/jinuopti/lilpop-server/database/gorm/userdb"
)

type UserServ struct {
	UserServer
}

func (s *UserServ) GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
	userID := req.UserId

	users := userdb.FindByIdUser(userID)
	if users == nil {
		Logd("User table is nil")
		return nil, nil
	}

	var userInfo *UserInfo = &UserInfo{}

	userInfo.UserId = users.UserId
	userInfo.Name = users.Name
	userInfo.CreateAt = timestamppb.New(users.CreatedAt)

	return &GetUserResponse{
		Info: userInfo,
	}, nil
}

func (s *UserServ) ListUser(ctx context.Context, req *ListUserRequest) (*ListUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListUser not implemented")
}
