package service

import (
	"context"
	"time"

	"wwetd-server/internal/dto"
	"wwetd-server/internal/interfaces"
)

type healthService struct {
	dependencies map[string]interfaces.Pinger
}

func NewHealthService(dependencies map[string]interfaces.Pinger) interfaces.HealthChecker {
	return &healthService{dependencies: dependencies}
}

func (s *healthService) Check(ctx context.Context) dto.HealthReport {
	report := dto.HealthReport{
		Status:       "ok",
		CheckedAt:    time.Now().UTC(),
		Dependencies: make(map[string]string, len(s.dependencies)),
	}

	for name, dependency := range s.dependencies {
		checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		err := dependency.Ping(checkCtx)
		cancel()

		if err != nil {
			report.Status = "degraded"
			report.Dependencies[name] = "down"
			continue
		}

		report.Dependencies[name] = "ok"
	}

	return report
}
