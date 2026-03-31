package handlers

import (
	"context"

	"github.com/abrshDev/user-service/internal/app/user/queries"
	pb "github.com/abrshDev/user-service/internal/transport/grpc/proto"
)

type UserGRPCHandler struct {
	pb.UnimplementedUserServiceServer
	getUserQuery *queries.GetUserHandler
}

// Added the missing closing brace here
func NewUserGRPCHandler(getUser *queries.GetUserHandler) *UserGRPCHandler {
	return &UserGRPCHandler{
		getUserQuery: getUser,
	}
}

func (h *UserGRPCHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	// 1. Wrap the string ID into the Query struct your logic expects
	query := queries.GetUserQuery{
		ID: req.Id,
	}

	// 2. Pass the struct (query) instead of the raw string (req.Id)
	user, err := h.getUserQuery.Execute(ctx, query)
	if err != nil {
		return nil, err
	}

	// 3. Map the database entity to the gRPC response
	return &pb.UserResponse{
		Id:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}, nil
}
