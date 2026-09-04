package hub

import (
	"context"
	"time"
)

func (s *Server) sloLoop() {
	if s.sloEval == nil {
		return
	}
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	s.runSLO()
	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		s.sloEval.Evaluate(ctx, s.store)
		cancel()
	}
}

func (s *Server) runSLO() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	s.sloEval.Evaluate(ctx, s.store)
}
