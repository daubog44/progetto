package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestNginxContainer(t *testing.T) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "nginx:alpine",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForHTTP("/").WithPort("80/tcp"),
	}
	nginxC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}
	defer nginxC.Terminate(ctx)

	ip, err := nginxC.Host(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, ip)
}
