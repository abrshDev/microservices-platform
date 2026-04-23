package handlers

import (
	"context"

	"github.com/abrshDev/user-service/internal/app/user/commands"
	"github.com/abrshDev/user-service/internal/app/user/queries"
	pb "github.com/abrshDev/user-service/internal/transport/grpc/proto"
)

type UserGRPCHandler struct {
	pb.UnimplementedUserServiceServer
	getUserQuery     *queries.GetUserHandler
	deleteUserCmd    *commands.DeleteUserHandler
	checkStatusQuery *queries.CheckUserStatusHandler
}

func NewUserGRPCHandler(getUser *queries.GetUserHandler, deleteUser *commands.DeleteUserHandler, checkuserstatus *queries.CheckUserStatusHandler) *UserGRPCHandler {
	return &UserGRPCHandler{
		getUserQuery:     getUser,
		deleteUserCmd:    deleteUser,
		checkStatusQuery: checkuserstatus,
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
		TenantId: uint32(user.TenantID),
	}, nil
}
func (h *UserGRPCHandler) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	err := h.deleteUserCmd.Execute(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.DeleteUserResponse{
		Message: "User deleted successfully",
	}, nil
}

func (h *UserGRPCHandler) CheckUserStatus(ctx context.Context, req *pb.CheckUserStatusRequest) (*pb.CheckUserStatusResponse, error) {
	query := queries.GetUserStatusQuery{
		ID: req.Id,
	}
	result, err := h.checkStatusQuery.Execute(ctx, query)
	if err != nil {
		return nil, err
	}

	return &pb.CheckUserStatusResponse{
		IsActive: result.IsActive,
		Role:     result.Role,
	}, nil

}
