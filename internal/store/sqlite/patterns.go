package sqlite

import (
	"context"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/insights"
	"github.com/HarryRaddatz/argus-observability/internal/model"
	"github.com/HarryRaddatz/argus-observability/internal/topology"
)

func (s *SQLite) RecordLogPatterns(ctx context.Context, entries []model.LogEntry) error {
	if len(entries) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `
INSERT INTO log_patterns (pattern_key, pattern, container, service, count, last_seen, sample)
VALUES (?, ?, ?, ?, 1, ?, ?)
ON CONFLICT(pattern_key, container) DO UPDATE SET
  count = count + 1,
  last_seen = excluded.last_seen,
  sample = CASE WHEN length(excluded.sample) > 0 THEN excluded.sample ELSE sample END
`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range entries {
		norm := insights.NormalizeLogMessage(e.Message)
		if norm == "" {
			continue
		}
		key := insights.PatternKey(norm)
		container := e.Labels["container"]
		service := e.Labels["service"]
		if service == "" {
			service = insights.InferServiceFromContainer(container)
		}
		sample := e.Message
		if len(sample) > 200 {
			sample = sample[:200]
		}
		if _, err := stmt.ExecContext(ctx, key, norm, container, service,
			e.TS.UTC().Format(time.RFC3339Nano), sample); err != nil {
			return err
		}
	}
	return tx.Commit()
}

type LogPatternRow struct {
	PatternKey string    `json:"pattern_key"`
	Pattern    string    `json:"pattern"`
	Container  string    `json:"container"`
	Service    string    `json:"service"`
	Count      int       `json:"count"`
	LastSeen   time.Time `json:"last_seen"`
	Sample     string    `json:"sample"`
}

func (s *SQLite) ListLogPatterns(ctx context.Context, since time.Time, limit int) ([]model.LogPattern, error) {
	rows, err := s.listLogPatterns(ctx, since, limit)
	if err != nil {
		return nil, err
	}
	out := make([]model.LogPattern, len(rows))
	for i, r := range rows {
		out[i] = model.LogPattern(r)
	}
	return out, nil
}

func (s *SQLite) listLogPatterns(ctx context.Context, since time.Time, limit int) ([]LogPatternRow, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT pattern_key, pattern, container, service, count, last_seen, sample
FROM log_patterns WHERE last_seen >= ?
ORDER BY count DESC LIMIT ?
`, since.UTC().Format(time.RFC3339Nano), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []LogPatternRow
	for rows.Next() {
		var r LogPatternRow
		var ts string
		if err := rows.Scan(&r.PatternKey, &r.Pattern, &r.Container, &r.Service, &r.Count, &ts, &r.Sample); err != nil {
			return nil, err
		}
		r.LastSeen, _ = time.Parse(time.RFC3339Nano, ts)
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *SQLite) RecordTopologyEdges(ctx context.Context, entries []model.LogEntry) error {
	if len(entries) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `
INSERT INTO topology_edges (source, target, kind, count, last_seen)
VALUES (?, ?, ?, 1, ?)
ON CONFLICT(source, target, kind) DO UPDATE SET
  count = count + 1,
  last_seen = excluded.last_seen
`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range entries {
		edges := topology.InferTopologyFromJSON(e)
		if len(edges) == 0 {
			edges = topology.InferTopologyEdges(e)
		}
		for _, edge := range edges {
			if _, err := stmt.ExecContext(ctx, edge.Source, edge.Target, edge.Kind,
				e.TS.UTC().Format(time.RFC3339Nano)); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

type TopologyNode struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type TopologyEdgeRow struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Kind   string `json:"kind"`
	Count  int    `json:"count"`
}

type TopologyResponse struct {
	Nodes []TopologyNode    `json:"nodes"`
	Edges []TopologyEdgeRow `json:"edges"`
}

func (s *SQLite) GetTopology(ctx context.Context, since time.Time) (model.TopologyGraph, error) {
	resp, err := s.queryTopology(ctx, since)
	if err != nil {
		return model.TopologyGraph{}, err
	}
	nodes := make([]model.TopologyNode, len(resp.Nodes))
	for i, n := range resp.Nodes {
		nodes[i] = model.TopologyNode(n)
	}
	edges := make([]model.TopologyEdge, len(resp.Edges))
	for i, e := range resp.Edges {
		edges[i] = model.TopologyEdge(e)
	}
	return model.TopologyGraph{Nodes: nodes, Edges: edges}, nil
}

func (s *SQLite) queryTopology(ctx context.Context, since time.Time) (TopologyResponse, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT source, target, kind, count FROM topology_edges
WHERE last_seen >= ? ORDER BY count DESC LIMIT 200
`, since.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return TopologyResponse{}, err
	}
	defer rows.Close()

	nodeSet := map[string]struct{}{}
	var edges []TopologyEdgeRow
	for rows.Next() {
		var e TopologyEdgeRow
		if err := rows.Scan(&e.Source, &e.Target, &e.Kind, &e.Count); err != nil {
			return TopologyResponse{}, err
		}
		nodeSet[e.Source] = struct{}{}
		nodeSet[e.Target] = struct{}{}
		edges = append(edges, e)
	}
	if err := rows.Err(); err != nil {
		return TopologyResponse{}, err
	}
	nodes := make([]TopologyNode, 0, len(nodeSet))
	for id := range nodeSet {
		nodes = append(nodes, TopologyNode{ID: id, Label: id})
	}
	return TopologyResponse{Nodes: nodes, Edges: edges}, nil
}
