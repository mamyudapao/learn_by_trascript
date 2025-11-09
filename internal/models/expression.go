package models

import "time"

// Expression は単語または熟語・慣用表現を表す
type Expression struct {
	ID              int       `db:"id"`
	Expression      string    `db:"expression"`
	Type            string    `db:"type"` // "word" または "phrase"
	Meaning         string    `db:"meaning"`
	Priority        int       `db:"priority"` // 1(低) ~ 5(高)
	Category        string    `db:"category"` // "engineering" / "business" / "casual"
	OccurrenceCount int       `db:"occurrence_count"`
	FirstSeenAt     time.Time `db:"first_seen_at"`
	LastSeenAt      time.Time `db:"last_seen_at"`
	UpdatedAt       time.Time `db:"updated_at"`

	// 一時的なフィールド（DBには保存されない）
	Context string `db:"-"` // 使用された文脈（処理中のみ使用）
}

// ExpressionOccurrence は表現の出現履歴を表す
type ExpressionOccurrence struct {
	ID           int       `db:"id"`
	ExpressionID int       `db:"expression_id"`
	Context      string    `db:"context"`
	OccurredAt   time.Time `db:"occurred_at"`
}

// ExpressionType は表現の種類
type ExpressionType string

const (
	TypeWord   ExpressionType = "word"
	TypePhrase ExpressionType = "phrase"
)

// Category は表現のカテゴリ
type Category string

const (
	CategoryEngineering Category = "engineering"
	CategoryBusiness    Category = "business"
	CategoryCasual      Category = "casual"
)
