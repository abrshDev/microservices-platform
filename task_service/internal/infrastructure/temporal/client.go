package temporal

import (
	"fmt"

	"go.temporal.io/sdk/client"
)

func NewClient(cfg Config) (client.Client, error) {
	c, err := client.Dial(client.Options{
		HostPort:  cfg.HostPort,
		Namespace: cfg.Namespace,
	})
	if err != nil {
		if err != nil {
			return nil, fmt.Errorf("failed to create temporal client: %w", err)
		}
	}

	return c, nil
}
