package queries

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/abrshDev/task-service/internal/domain/entities"
	"github.com/abrshDev/task-service/internal/transport/grpc/proto/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- MOCKS ---

type MockTaskRepo struct {
	mock.Mock
}

func (m *MockTaskRepo) GetByID(ctx context.Context, id uuid.UUID) (*entities.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Task), args.Error(1)
}

// Stub remaining methods to satisfy interface
func (m *MockTaskRepo) Create(ctx context.Context, t *entities.Task) error { return nil }
func (m *MockTaskRepo) UpdateStatus(ctx context.Context, id uuid.UUID, s entities.TaskStatus) error {
	return nil
}
func (m *MockTaskRepo) ListByUserID(ctx context.Context, uid uuid.UUID) ([]entities.Task, error) {
	return nil, nil
}
func (m *MockTaskRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }

type MockUserClient struct {
	mock.Mock
}

func (m *MockUserClient) GetUser(ctx context.Context, userID string) (*user.UserResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.UserResponse), args.Error(1)
}

// --- TESTS ---

func TestGetTaskHandler_Execute(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	t.Run("Success Path", func(t *testing.T) {
		repo := new(MockTaskRepo)
		uClient := new(MockUserClient)
		handler := NewGetTaskHandler(repo, uClient, logger)

		tID, uID := uuid.New(), uuid.New()

		// Setup Expectations
		repo.On("GetByID", mock.Anything, tID).Return(&entities.Task{ID: tID, UserID: uID, Title: "Test"}, nil)
		uClient.On("GetUser", mock.Anything, uID.String()).Return(&user.UserResponse{Username: "Abraham"}, nil)

		res, err := handler.Execute(context.Background(), GetTaskQuery{ID: tID.String()})

		assert.NoError(t, err)
		assert.Equal(t, "Abraham", res.User.Name)
	})

	t.Run("Graceful Degradation Path", func(t *testing.T) {
		repo := new(MockTaskRepo)
		uClient := new(MockUserClient)
		handler := NewGetTaskHandler(repo, uClient, logger)

		tID := uuid.New()
		repo.On("GetByID", mock.Anything, tID).Return(&entities.Task{ID: tID, Title: "Test"}, nil)
		uClient.On("GetUser", mock.Anything, mock.Anything).Return(nil, errors.New("network fail"))

		res, err := handler.Execute(context.Background(), GetTaskQuery{ID: tID.String()})

		assert.NoError(t, err)                    // Should NOT error
		assert.Equal(t, "Unknown", res.User.Name) // Should fallback
	})
}
