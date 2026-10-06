package postgres

import (
	"context"
	"sort"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/insights"
	"github.com/HarryRaddatz/argus-observability/internal/model"
	"github.com/HarryRaddatz/argus-observability/internal/topology"
)

// topologyEdgeLimit caps the graph returned to the panel.
const topologyEdgeLimit = 200

func (s *Postgres) RecordLogPatterns(ctx context.Context, entries []model.LogEntry) error {
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
  count = log_patterns.count + 1,
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

func (s *Postgres) ListLogPatterns(ctx context.Context, since time.Time, limit int) ([]model.LogPattern, error) {
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

func (s *Postgres) listLogPatterns(ctx context.Context, since time.Time, limit int) ([]LogPatternRow, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.rdb.QueryContext(ctx, `
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

func (s *Postgres) RecordTopologyEdges(ctx context.Context, entries []model.LogEntry) error {
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
  count = topology_edges.count + 1,
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

// RecordTopologyLinks stores dependencies observed by an agent. Counts arrive
// as per-window increments, so they accumulate on conflict.
func (s *Postgres) RecordTopologyLinks(ctx context.Context, links []model.TopologyLink) error {
	if len(links) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `
INSERT INTO topology_links (source, target, kind, port, count, last_seen)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(source, target, port) DO UPDATE SET
  kind = excluded.kind,
  count = topology_links.count + excluded.count,
  last_seen = excluded.last_seen
`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, link := range links {
		if link.Source == "" || link.Target == "" || link.Count == 0 {
			continue
		}
		if _, err := stmt.ExecContext(ctx, link.Source, link.Target, link.Kind, link.Port,
			link.Count, link.TS.UTC().Format(time.RFC3339Nano)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// GetTopology merges dependencies observed by an agent with the ones inferred
// from log text. When both describe the same pair the observed edge wins: it
// carries the port and cannot be a false positive from a URL inside a message.
func (s *Postgres) GetTopology(ctx context.Context, since time.Time) (model.TopologyGraph, error) {
	observed, err := s.queryTopologyLinks(ctx, since)
	if err != nil {
		return model.TopologyGraph{}, err
	}
	inferred, err := s.queryTopology(ctx, since)
	if err != nil {
		return model.TopologyGraph{}, err
	}

	edges := make([]model.TopologyEdge, 0, len(observed)+len(inferred.Edges))
	seen := make(map[string]struct{}, len(observed))
	for _, edge := range observed {
		seen[edge.Source+"\x00"+edge.Target] = struct{}{}
		edges = append(edges, edge)
	}
	for _, row := range inferred.Edges {
		if _, ok := seen[row.Source+"\x00"+row.Target]; ok {
			continue
		}
		edges = append(edges, model.TopologyEdge{
			Source: row.Source,
			Target: row.Target,
			Kind:   row.Kind,
			Count:  row.Count,
			Origin: model.TopologyOriginLog,
		})
	}
	sort.SliceStable(edges, func(i, j int) bool { return edges[i].Count > edges[j].Count })
	if len(edges) > topologyEdgeLimit {
		edges = edges[:topologyEdgeLimit]
	}

	nodeSet := map[string]struct{}{}
	for _, edge := range edges {
		nodeSet[edge.Source] = struct{}{}
		nodeSet[edge.Target] = struct{}{}
	}
	nodes := make([]model.TopologyNode, 0, len(nodeSet))
	for id := range nodeSet {
		nodes = append(nodes, model.TopologyNode{ID: id, Label: id})
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })

	return model.TopologyGraph{Nodes: nodes, Edges: edges}, nil
}

func (s *Postgres) queryTopologyLinks(ctx context.Context, since time.Time) ([]model.TopologyEdge, error) {
	rows, err := s.rdb.QueryContext(ctx, `
SELECT source, target, kind, port, count FROM topology_links
WHERE last_seen >= ? ORDER BY count DESC LIMIT 200
`, since.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.TopologyEdge
	for rows.Next() {
		edge := model.TopologyEdge{Origin: model.TopologyOriginKernel}
		if err := rows.Scan(&edge.Source, &edge.Target, &edge.Kind, &edge.Port, &edge.Count); err != nil {
			return nil, err
		}
		out = append(out, edge)
	}
	return out, rows.Err()
}

func (s *Postgres) queryTopology(ctx context.Context, since time.Time) (TopologyResponse, error) {
	rows, err := s.rdb.QueryContext(ctx, `
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
