package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mamyudapao/learn-by-transcript/internal/models"
)

// SQLiteRepository はSQLiteベースのRepository実装
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository は新しいSQLiteRepositoryを作成
func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// マイグレーション実行
	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &SQLiteRepository{db: db}, nil
}

// runMigrations はDBマイグレーションを実行
func runMigrations(db *sql.DB) error {
	schema, err := os.ReadFile("migrations/001_initial_schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
}

// SaveExpression は新しい表現を保存
func (r *SQLiteRepository) SaveExpression(ctx context.Context, expr *models.Expression) error {
	query := `
		INSERT INTO expressions (expression, type, meaning, priority, category)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query, expr.Expression, expr.Type, expr.Meaning, expr.Priority, expr.Category)
	if err != nil {
		return fmt.Errorf("failed to save expression: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	expr.ID = int(id)
	return nil
}

// GetExpression は表現を取得
func (r *SQLiteRepository) GetExpression(ctx context.Context, expression string) (*models.Expression, error) {
	query := `
		SELECT id, expression, type, meaning, priority, category, occurrence_count,
		       first_seen_at, last_seen_at, updated_at
		FROM expressions
		WHERE expression = ?
	`

	var expr models.Expression
	err := r.db.QueryRowContext(ctx, query, expression).Scan(
		&expr.ID, &expr.Expression, &expr.Type, &expr.Meaning, &expr.Priority, &expr.Category,
		&expr.OccurrenceCount, &expr.FirstSeenAt, &expr.LastSeenAt, &expr.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get expression: %w", err)
	}

	return &expr, nil
}

// ExpressionExists は表現が既に存在するか確認
func (r *SQLiteRepository) ExpressionExists(ctx context.Context, expression string) (bool, error) {
	expr, err := r.GetExpression(ctx, expression)
	if err != nil {
		return false, err
	}
	return expr != nil, nil
}

// AddOccurrence は出現履歴を追加
func (r *SQLiteRepository) AddOccurrence(ctx context.Context, expressionID int, context string) error {
	query := `
		INSERT INTO expression_occurrences (expression_id, context)
		VALUES (?, ?)
	`

	_, err := r.db.ExecContext(ctx, query, expressionID, context)
	if err != nil {
		return fmt.Errorf("failed to add occurrence: %w", err)
	}

	return nil
}

// UpdatePriority は優先度を更新
func (r *SQLiteRepository) UpdatePriority(ctx context.Context, expressionID int, priority int) error {
	query := `
		UPDATE expressions
		SET priority = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query, priority, expressionID)
	if err != nil {
		return fmt.Errorf("failed to update priority: %w", err)
	}

	return nil
}

// GetAllExpressions はすべての表現を取得
func (r *SQLiteRepository) GetAllExpressions(ctx context.Context) ([]*models.Expression, error) {
	query := `
		SELECT id, expression, type, meaning, priority, category, occurrence_count,
		       first_seen_at, last_seen_at, updated_at
		FROM expressions
		ORDER BY priority DESC, occurrence_count DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query expressions: %w", err)
	}
	defer rows.Close()

	var expressions []*models.Expression
	for rows.Next() {
		var expr models.Expression
		err := rows.Scan(
			&expr.ID, &expr.Expression, &expr.Type, &expr.Meaning, &expr.Priority, &expr.Category,
			&expr.OccurrenceCount, &expr.FirstSeenAt, &expr.LastSeenAt, &expr.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expression: %w", err)
		}
		expressions = append(expressions, &expr)
	}

	return expressions, nil
}

// GetTopExpressions は優先度・出現頻度の高い表現を取得
func (r *SQLiteRepository) GetTopExpressions(ctx context.Context, limit int) ([]*models.Expression, error) {
	query := `
		SELECT id, expression, type, meaning, priority, category, occurrence_count,
		       first_seen_at, last_seen_at, updated_at
		FROM expressions
		ORDER BY priority DESC, occurrence_count DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top expressions: %w", err)
	}
	defer rows.Close()

	var expressions []*models.Expression
	for rows.Next() {
		var expr models.Expression
		err := rows.Scan(
			&expr.ID, &expr.Expression, &expr.Type, &expr.Meaning, &expr.Priority, &expr.Category,
			&expr.OccurrenceCount, &expr.FirstSeenAt, &expr.LastSeenAt, &expr.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expression: %w", err)
		}
		expressions = append(expressions, &expr)
	}

	return expressions, nil
}

// ListExpressions はすべての表現を取得（優先度・出現回数順）
func (r *SQLiteRepository) ListExpressions(ctx context.Context) ([]*models.Expression, error) {
	query := `
		SELECT id, expression, type, meaning, priority, category, occurrence_count, first_seen_at, last_seen_at, updated_at
		FROM expressions
		ORDER BY priority DESC, occurrence_count DESC, expression ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query expressions: %w", err)
	}
	defer rows.Close()

	var expressions []*models.Expression
	for rows.Next() {
		var expr models.Expression
		err := rows.Scan(
			&expr.ID, &expr.Expression, &expr.Type, &expr.Meaning, &expr.Priority, &expr.Category,
			&expr.OccurrenceCount, &expr.FirstSeenAt, &expr.LastSeenAt, &expr.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expression: %w", err)
		}
		expressions = append(expressions, &expr)
	}

	return expressions, nil
}

// GetOccurrences は表現の出現履歴を取得
func (r *SQLiteRepository) GetOccurrences(ctx context.Context, expressionID int) ([]*models.ExpressionOccurrence, error) {
	query := `
		SELECT id, expression_id, context, occurred_at
		FROM expression_occurrences
		WHERE expression_id = ?
		ORDER BY occurred_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, expressionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query occurrences: %w", err)
	}
	defer rows.Close()

	var occurrences []*models.ExpressionOccurrence
	for rows.Next() {
		var occ models.ExpressionOccurrence
		err := rows.Scan(&occ.ID, &occ.ExpressionID, &occ.Context, &occ.OccurredAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan occurrence: %w", err)
		}
		occurrences = append(occurrences, &occ)
	}

	return occurrences, nil
}

// Close はリソースをクリーンアップ
func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}
