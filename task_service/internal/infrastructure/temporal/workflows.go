package temporal

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type CreateTaskInput struct {
	UserID      string `json:"user_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}
type ValidateUserInput struct {
	UserID string
}

type ValidateUserResult struct {
	TenantID uint64
}

type SaveTaskInput struct {
	UserID      string
	TenantID    uint64
	Title       string
	Description string
}

type SaveTaskResult struct {
	TaskID string
}

type PublishEventInput struct {
	TaskID      string
	UserID      string
	TenantID    uint64
	Title       string
	Description string
}

type PublishEventResult struct {
	Success bool
}

func CreateTaskWorkflow(ctx workflow.Context, input CreateTaskInput) error {

	retryPolicy := temporal.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    time.Minute,
		MaximumAttempts:    5,
	}
	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy:         &retryPolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, activityOpts)
	var validateResult ValidateUserResult
	err := workflow.ExecuteActivity(ctx, (&Activities{}).ValidateUserActivity, ValidateUserInput{
		UserID: input.UserID,
	}).Get(ctx, &validateResult)
	if err != nil {
		return err
	}

	// Step 2: Save task to database
	var saveResult SaveTaskResult
	err = workflow.ExecuteActivity(ctx, (&Activities{}).SaveTaskActivity, SaveTaskInput{
		UserID:      input.UserID,
		TenantID:    validateResult.TenantID,
		Title:       input.Title,
		Description: input.Description,
	}).Get(ctx, &saveResult)
	if err != nil {
		return err
	}

	// Step 3: Publish event to Kafka
	var publishResult PublishEventResult
	err = workflow.ExecuteActivity(ctx, (&Activities{}).PublishEventActivity, PublishEventInput{
		TaskID:      saveResult.TaskID,
		UserID:      input.UserID,
		TenantID:    validateResult.TenantID,
		Title:       input.Title,
		Description: input.Description,
	}).Get(ctx, &publishResult)
	if err != nil {
		return err
	}

	return nil

}
