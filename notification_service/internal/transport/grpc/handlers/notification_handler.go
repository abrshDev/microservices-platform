package handlers

import (
	"context"

	"github.com/abrshDev/notification_service/internal/app/notification/commands"
	"github.com/abrshDev/notification_service/internal/transport/grpc/proto/notification"
)

type NotificationGRPCHandler struct {
	notification.UnimplementedNotificationServiceServer
	sendHandler *commands.SendNotificationHandler
}

func NewNotificationGRPCHandler(sh *commands.SendNotificationHandler) *NotificationGRPCHandler {
	return &NotificationGRPCHandler{sendHandler: sh}
}

func (h *NotificationGRPCHandler) SendNotification(ctx context.Context, req *notification.NotificationRequest) (*notification.NotificationResponse, error) {
	cmd := commands.SendNotificationCommand{
		UserID:  req.UserId,
		Message: req.Message,
		Type:    req.Type,
	}

	if err := h.sendHandler.Handle(ctx, cmd); err != nil {
		return &notification.NotificationResponse{Success: false}, err
	}

	return &notification.NotificationResponse{Success: true}, nil
}
