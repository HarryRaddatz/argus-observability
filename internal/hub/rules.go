package hub

import (
	"context"
	"time"
)

func (s *Server) rulesLoop() {
	if s.rules == nil {
		return
	}
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	s.runRules()
	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		s.rules.Evaluate(ctx, s.store)
		cancel()
	}
}

func (s *Server) runRules() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	s.rules.Evaluate(ctx, s.store)
}
