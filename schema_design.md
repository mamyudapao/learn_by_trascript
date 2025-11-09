# データベーススキーマ設計

## 基本方針

- SQLite3を使用
- シンプルな正規化（過度な正規化は避ける）
- 重複チェックを効率化するインデックス
- 将来の拡張性を考慮

## スキーマ案（v1）

### expressionsテーブル

```sql
CREATE TABLE expressions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    expression TEXT NOT NULL,           -- 単語または熟語
    type TEXT NOT NULL,                 -- 'word' または 'phrase'
    meaning TEXT,                       -- 日本語の意味
    context TEXT,                       -- 使用された文脈（元の文）
    priority INTEGER,                   -- 1(低) ~ 5(高)
    category TEXT,                      -- 'engineering' / 'business' / 'casual'
    source_meeting TEXT,                -- 会議名/日付
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(expression)                  -- 重複防止
);

-- インデックス
CREATE INDEX idx_expressions_priority ON expressions(priority DESC);
CREATE INDEX idx_expressions_category ON expressions(category);
CREATE INDEX idx_expressions_type ON expressions(type);
CREATE INDEX idx_expressions_created_at ON expressions(created_at DESC);
```

## 議論ポイント

### 1. 会議履歴の管理

現在は `source_meeting` を単なる文字列（例: "2025-01-15 Team Sync"）で保存していますが、以下のどちらが良いですか？

**Option A: 現状のまま（シンプル）**
```sql
source_meeting TEXT  -- "2025-01-15 Team Sync"
```

**Option B: 会議テーブルを分離（正規化）**
```sql
CREATE TABLE meetings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    date DATE NOT NULL,
    duration_minutes INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- expressionsテーブルに外部キー追加
ALTER TABLE expressions ADD COLUMN meeting_id INTEGER REFERENCES meetings(id);
```

メリット：
- 会議ごとの統計が取りやすい
- 会議メタデータを追加しやすい

デメリット：
- 複雑になる
- CLIツールとしてはオーバースペック？

### 2. 同じ表現が複数の会議で出現した場合

**現在の設計**：`UNIQUE(expression)` なので、同じ表現は1つしか保存されない

**問題**：
- "deprecate" という単語が3つの会議で出てきた場合、最初の1つだけ保存される
- 後から「どの会議で出てきたか」が分からない

**解決策の選択肢：**

**Option A: 現状維持（シンプル）**
- 同じ表現は1つだけ保存
- `source_meeting` は最初に出現した会議のみ

**Option B: 多対多の関係を作る**
```sql
CREATE TABLE expressions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    expression TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL,
    meaning TEXT,
    priority INTEGER,
    category TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE expression_occurrences (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    expression_id INTEGER NOT NULL REFERENCES expressions(id),
    meeting_name TEXT NOT NULL,
    context TEXT,                       -- その会議での使用文脈
    occurred_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(expression_id, meeting_name)
);
```

メリット：
- 同じ表現が複数の会議で出てきた履歴を保持
- 頻出度の分析が可能

デメリット：
- 複雑になる

### 3. priorityとcategoryの管理

**現在の設計**：
```sql
priority INTEGER,    -- 1 ~ 5
category TEXT,       -- 'engineering' / 'business' / 'casual'
```

**問題**：
- 文字列で保存すると、タイポのリスク（"enginering" など）
- 後からカテゴリ追加・変更が難しい

**選択肢：**

**Option A: 現状維持（シンプル）**
- アプリ側で文字列をバリデーション

**Option B: マスターテーブルを作る**
```sql
CREATE TABLE categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL
);

INSERT INTO categories (name, display_name) VALUES
    ('engineering', 'ソフトウェアエンジニアリング'),
    ('business', '仕事・議論'),
    ('casual', '雑談');

-- expressionsテーブルを変更
ALTER TABLE expressions ADD COLUMN category_id INTEGER REFERENCES categories(id);
```

### 4. meaningの管理

現在はLLMが日本語訳を生成しますが：

**問題**：
- 同じ単語でも文脈によって意味が変わる
- 複数の意味を持つ単語（例: "run" → 実行する、走る、運営する）

**選択肢：**

**Option A: 現状維持（1つの意味だけ）**
```sql
meaning TEXT  -- "非推奨にする"
```

**Option B: 複数の意味を保存**
```sql
meaning TEXT  -- JSON形式 ["非推奨にする", "廃止する"]
```

**Option C: 別テーブルで管理**
```sql
CREATE TABLE meanings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    expression_id INTEGER REFERENCES expressions(id),
    meaning TEXT NOT NULL,
    context TEXT,
    is_primary BOOLEAN DEFAULT 0
);
```

## 決定：出現頻度ベースの優先度更新

**要件：**
- 同じ表現が複数の会議で出現した場合、履歴を保存
- 出現回数（頻度）を追跡
- 出現頻度が高い表現は優先度を上げる

**採用する設計：多対多関係**

## 確定スキーマ（v1）

### 1. expressions テーブル（基本情報）

```sql
CREATE TABLE expressions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    expression TEXT NOT NULL UNIQUE,    -- 単語または熟語
    type TEXT NOT NULL,                 -- 'word' または 'phrase'
    meaning TEXT,                       -- 日本語の意味
    priority INTEGER,                   -- 1(低) ~ 5(高)、出現頻度で更新
    category TEXT,                      -- 'engineering' / 'business' / 'casual'
    occurrence_count INTEGER DEFAULT 1, -- 出現回数（自動計算）
    first_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- インデックス
CREATE INDEX idx_expressions_priority ON expressions(priority DESC);
CREATE INDEX idx_expressions_occurrence ON expressions(occurrence_count DESC);
CREATE INDEX idx_expressions_category ON expressions(category);
```

### 2. expression_occurrences テーブル（出現履歴）

```sql
CREATE TABLE expression_occurrences (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    expression_id INTEGER NOT NULL,
    context TEXT,                       -- 使用された文脈
    occurred_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (expression_id) REFERENCES expressions(id) ON DELETE CASCADE
);

-- インデックス
CREATE INDEX idx_occurrences_expression ON expression_occurrences(expression_id);
CREATE INDEX idx_occurrences_date ON expression_occurrences(occurred_at DESC);
```

### 3. トリガー：出現回数の自動更新

```sql
-- 新しい出現が追加されたら、occurrence_count と last_seen_at を更新
CREATE TRIGGER update_occurrence_count
AFTER INSERT ON expression_occurrences
BEGIN
    UPDATE expressions
    SET
        occurrence_count = occurrence_count + 1,
        last_seen_at = NEW.occurred_at,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = NEW.expression_id;
END;
```

### 4. 優先度更新ロジック（アプリケーション側）

**基本方針：**
- 初回出現時：LLMが判定した優先度をそのまま使用
- 2回目以降：出現頻度に応じて優先度をブースト

**ロジック案：**
```go
func UpdatePriorityBasedOnOccurrence(basePriority int, occurrenceCount int) int {
    // 出現回数に応じて優先度をブースト
    boost := 0
    if occurrenceCount >= 5 {
        boost = 2  // 5回以上出現 → +2
    } else if occurrenceCount >= 3 {
        boost = 1  // 3回以上出現 → +1
    }

    newPriority := basePriority + boost
    if newPriority > 5 {
        newPriority = 5  // 最大5
    }

    return newPriority
}
```

**例：**
- "deprecate" が初回出現、LLMが priority=3 と判定 → 3
- 2回目の出現 → 3（変わらず）
- 3回目の出現 → 4（+1）
- 5回目の出現 → 5（+2、最大値）

## クエリ例

### 1. 頻出表現トップ10
```sql
SELECT expression, occurrence_count, priority, category
FROM expressions
ORDER BY occurrence_count DESC
LIMIT 10;
```

### 2. 最近出現した表現
```sql
SELECT e.expression, e.meaning, e.priority, e.last_seen_at, e.occurrence_count
FROM expressions e
ORDER BY e.last_seen_at DESC
LIMIT 20;
```

### 3. 優先度の高い未学習表現
```sql
SELECT expression, meaning, occurrence_count
FROM expressions
WHERE priority >= 4
ORDER BY priority DESC, occurrence_count DESC;
```

## 処理フロー（更新版）

```
1. transcriptを処理
   ↓
2. 単語・熟語を抽出
   ↓
3. 各表現について：
   - expressions テーブルに存在するか確認
   - 存在しない → 新規作成（LLMで優先度・意味・カテゴリ判定）
   - 存在する → expression_occurrences に出現履歴追加
                （トリガーで occurrence_count と last_seen_at を自動更新）
   ↓
4. 出現回数に応じて priority を更新（3回以上で+1、5回以上で+2）
   ↓
5. （オプション）出力先にエクスポート
```

## 次のステップ

- [ ] Go側のRepository実装設計
- [ ] 優先度更新ロジックの詳細決定（閾値など）
- [ ] マイグレーション管理の方針
