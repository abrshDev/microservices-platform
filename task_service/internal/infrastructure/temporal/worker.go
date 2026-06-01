package temporal

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func NewWorker(c client.Client, taskQueue string, activities *Activities) worker.Worker {
	w := worker.New(c, taskQueue, worker.Options{})
	w.RegisterWorkflow(CreateTaskWorkflow)
	w.RegisterActivity(activities.ValidateUserActivity)
	w.RegisterActivity(activities.SaveTaskActivity)
	w.RegisterActivity(activities.PublishEventActivity)
	return w
}
