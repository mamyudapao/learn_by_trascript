-- 英語表現テーブル
CREATE TABLE IF NOT EXISTS expressions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    expression TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL,
    meaning TEXT,
    priority INTEGER,
    category TEXT,
    occurrence_count INTEGER DEFAULT 1,
    first_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 出現履歴テーブル
CREATE TABLE IF NOT EXISTS expression_occurrences (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    expression_id INTEGER NOT NULL,
    context TEXT,
    occurred_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (expression_id) REFERENCES expressions(id) ON DELETE CASCADE
);

-- インデックス
CREATE INDEX IF NOT EXISTS idx_expressions_priority ON expressions(priority DESC);
CREATE INDEX IF NOT EXISTS idx_expressions_occurrence ON expressions(occurrence_count DESC);
CREATE INDEX IF NOT EXISTS idx_expressions_category ON expressions(category);
CREATE INDEX IF NOT EXISTS idx_occurrences_expression ON expression_occurrences(expression_id);
CREATE INDEX IF NOT EXISTS idx_occurrences_date ON expression_occurrences(occurred_at DESC);

-- トリガー: 出現回数の自動更新
CREATE TRIGGER IF NOT EXISTS update_occurrence_count
AFTER INSERT ON expression_occurrences
BEGIN
    UPDATE expressions
    SET
        occurrence_count = occurrence_count + 1,
        last_seen_at = NEW.occurred_at,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = NEW.expression_id;
END;
