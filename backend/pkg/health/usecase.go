package health

import "context"

// Checker represents a dependency health check.
type Checker interface {
	Name() string
	Check(ctx context.Context) error
}

// ReadinessUseCase describes readiness verification.
type ReadinessUseCase interface {
	Ready(ctx context.Context) error
}

type service struct {
	checkers []Checker
}

// NewService aggregates dependency checkers.
func NewService(checkers ...Checker) ReadinessUseCase {
	return &service{checkers: checkers}
}

func (s *service) Ready(ctx context.Context) error {
	for _, ch := range s.checkers {
		if err := ch.Check(ctx); err != nil {
			return err
		}
	}
	return nil
}
