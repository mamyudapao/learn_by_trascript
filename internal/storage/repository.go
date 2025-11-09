package storage

import (
	"context"

	"github.com/mamyudapao/learn-by-transcript/internal/models"
)

// Repository はストレージの抽象インターフェース
type Repository interface {
	// SaveExpression は新しい表現を保存
	SaveExpression(ctx context.Context, expr *models.Expression) error

	// GetExpression は表現を取得
	GetExpression(ctx context.Context, expression string) (*models.Expression, error)

	// ExpressionExists は表現が既に存在するか確認
	ExpressionExists(ctx context.Context, expression string) (bool, error)

	// AddOccurrence は出現履歴を追加
	AddOccurrence(ctx context.Context, expressionID int, context string) error

	// UpdatePriority は優先度を更新
	UpdatePriority(ctx context.Context, expressionID int, priority int) error

	// GetAllExpressions はすべての表現を取得
	GetAllExpressions(ctx context.Context) ([]*models.Expression, error)

	// GetTopExpressions は優先度・出現頻度の高い表現を取得
	GetTopExpressions(ctx context.Context, limit int) ([]*models.Expression, error)

	// ListExpressions はすべての表現を取得（GetAllExpressionsのエイリアス）
	ListExpressions(ctx context.Context) ([]*models.Expression, error)

	// GetOccurrences は表現の出現履歴を取得
	GetOccurrences(ctx context.Context, expressionID int) ([]*models.ExpressionOccurrence, error)

	// Close はリソースをクリーンアップ
	Close() error
}
