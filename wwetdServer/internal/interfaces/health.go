package interfaces

import (
	"context"

	"wwetd-server/internal/dto"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type HealthChecker interface {
	Check(ctx context.Context) dto.HealthReport
}
