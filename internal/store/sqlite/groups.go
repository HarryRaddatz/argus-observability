package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/HarryRaddatz/argus-observability/internal/model"
	"github.com/google/uuid"
)

func (s *SQLite) ListWorkloadGroups(ctx context.Context) ([]model.WorkloadGroup, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, name, kind, description, label_key, label_value, containers_json, created_at, updated_at
FROM workload_groups ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWorkloadGroups(rows)
}

func (s *SQLite) GetWorkloadGroup(ctx context.Context, id string) (model.WorkloadGroup, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, name, kind, description, label_key, label_value, containers_json, created_at, updated_at
FROM workload_groups WHERE id=?`, id)
	g, err := scanWorkloadGroupRow(row)
	if errors.Is(err, sql.ErrNoRows) {
		return model.WorkloadGroup{}, ErrNotFound
	}
	return g, err
}

func (s *SQLite) CreateWorkloadGroup(ctx context.Context, in model.WorkloadGroupInput) (model.WorkloadGroup, error) {
	now := time.Now().UTC()
	id := uuid.NewString()
	containersJSON, err := json.Marshal(in.Containers)
	if err != nil {
		return model.WorkloadGroup{}, err
	}
	_, err = s.db.ExecContext(ctx, `
INSERT INTO workload_groups (id, name, kind, description, label_key, label_value, containers_json, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, in.Name, in.Kind, in.Description, in.LabelKey, in.LabelValue, string(containersJSON),
		now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano),
	)
	if err != nil {
		return model.WorkloadGroup{}, err
	}
	return s.GetWorkloadGroup(ctx, id)
}

func (s *SQLite) UpdateWorkloadGroup(ctx context.Context, id string, in model.WorkloadGroupInput) (model.WorkloadGroup, error) {
	containersJSON, err := json.Marshal(in.Containers)
	if err != nil {
		return model.WorkloadGroup{}, err
	}
	now := time.Now().UTC()
	res, err := s.db.ExecContext(ctx, `
UPDATE workload_groups SET name=?, kind=?, description=?, label_key=?, label_value=?, containers_json=?, updated_at=?
WHERE id=?`,
		in.Name, in.Kind, in.Description, in.LabelKey, in.LabelValue, string(containersJSON),
		now.Format(time.RFC3339Nano), id,
	)
	if err != nil {
		return model.WorkloadGroup{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.WorkloadGroup{}, ErrNotFound
	}
	return s.GetWorkloadGroup(ctx, id)
}

func (s *SQLite) DeleteWorkloadGroup(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM workload_groups WHERE id=?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func scanWorkloadGroups(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]model.WorkloadGroup, error) {
	var out []model.WorkloadGroup
	for rows.Next() {
		g, err := scanWorkloadGroupFromScan(rows.Scan)
		if err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func scanWorkloadGroupRow(row interface{ Scan(dest ...any) error }) (model.WorkloadGroup, error) {
	return scanWorkloadGroupFromScan(row.Scan)
}

func scanWorkloadGroupFromScan(scan func(dest ...any) error) (model.WorkloadGroup, error) {
	var g model.WorkloadGroup
	var desc, labelKey, labelValue, containersJSON, createdAt, updatedAt string
	if err := scan(&g.ID, &g.Name, &g.Kind, &desc, &labelKey, &labelValue, &containersJSON, &createdAt, &updatedAt); err != nil {
		return model.WorkloadGroup{}, err
	}
	g.Description = desc
	g.LabelKey = labelKey
	g.LabelValue = labelValue
	_ = json.Unmarshal([]byte(containersJSON), &g.Containers)
	if g.Containers == nil {
		g.Containers = []string{}
	}
	g.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	g.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return g, nil
}
