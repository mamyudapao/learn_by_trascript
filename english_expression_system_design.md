# 英語表現抽出システム設計

## プロジェクト概要
Google Meetのtranscript（文字起こし）とsummary（要約）から、会議で出た英語表現・単語・フレーズを自動抽出し、社内Notionに登録するシステム。

## 利用条件・制約
- 使用できる環境: 社用PCとBYODスマホのみ
- 外部SaaSへのデータ送信は禁止（Vertex AIのみ利用可）
- Vertex AIのAPIキーを会社から発行してもらう想定
- 出力先: 会社のNotion（**API利用可否は未確認**）
- 情報源: Google Meetのtranscript/summary

## 機能要件
1. 会議データ取得（手動アップロード or GCS連携）
2. テキスト前処理（話者分割・ノイズ除去）
3. Vertex AIによる語彙抽出
   - Expression（表現）
   - Meaning（意味）
   - Category（カテゴリ）
   - Context（文脈）
   - Importance（重要度）
4. 重複判定（string一致＋embedding類似）
5. Notion API登録（英語表現データベース）
6. セキュリティ・ログ管理（APIキー保護・処理監査）
7. 将来的な拡張: Slack通知、自動バッチ処理

## MVP構成図
```
[Google Meet]
    ↓ (手動ダウンロード)
[transcript.txt / summary.txt]
    ↓ (手動アップロード or GCS経由)
┌─────────────────────────────────┐
│  処理エンジン (社用PC or Cloud Run) │
├─────────────────────────────────┤
│ 1. テキスト前処理               │
│ 2. Vertex AI API呼び出し        │
│    - Gemini Pro/Flash (抽出)    │
│    - Text Embeddings (類似判定) │
│ 3. 重複チェック                 │
│ 4. Notion API登録               │
└─────────────────────────────────┘
    ↓
[Notion Database (英語表現DB)]
    ↓
[Notion Page (学習記録)]
```

## データフロー（MVP版）
1. transcript/summaryをJSON形式で入力
2. Gemini Pro/Flashでプロンプトベース抽出（5項目）
3. 既存Notion DBから全表現を取得
4. String一致 → 既存ならスキップ
5. Embedding類似度計算（cosine > 0.9ならスキップ）
6. 新規表現のみNotion DBに追加

## 使用API・モデル

### Vertex AI
- **抽出モデル候補**:
  - `gemini-1.5-flash-002` (速度重視・コスト安 - MVP推奨)
  - `gemini-1.5-pro-002` (精度重視)
- **Embeddingモデル**: `text-embedding-004`
  - 768次元ベクトル
  - 多言語対応

### Notion API（利用可能な場合）
- **Database検索**: `POST https://api.notion.com/v1/databases/{db_id}/query`
- **Page作成**: `POST https://api.notion.com/v1/pages`
- **必要な権限**: データベースへの読み書き権限

### 出力先の代替案（Notion API利用不可の場合）

**Option 1: Google Sheets API**
- 会社のGoogle Workspaceで利用可能なケースが多い
- 重複チェックもSheets上で実行可能
- Notion同様の構造化データ管理が可能
- API: `google-api-python-client` + `gspread`

**Option 2: CSVエクスポート→手動インポート**
- 抽出結果をCSV形式で出力
- Notionのデータベースインポート機能で手動取り込み
- 重複チェックはNotion側のデータベースフィルタで対応
- 最もシンプルだが半自動化

**Option 3: Markdown形式で出力**
- 1会議につき1つのMarkdownファイル生成
- ファイルをNotion Webアプリにドラッグ&ドロップ
- 重複チェックは目視または簡易スクリプトで対応
- 構造化は弱いが手軽

**Option 4: ローカルSQLite DB + Web UI**
- SQLiteで英語表現データベースを構築
- 簡易Web UI（Flask/FastAPI）で閲覧・検索
- Notionへは定期的にエクスポートして手動同期
- オフライン環境でも動作可能

**推奨順位:**
1. **Notion API**（第一優先）
2. **Google Sheets API**（次点、自動化可能）
3. CSV出力→手動インポート（最速MVP）
4. ローカルSQLite DB（完全自己完結）
5. Markdown出力（最もシンプル）

**アーキテクチャ方針: 出力先の抽象化**
- 出力先を切り替え可能な設計にする
- インターフェース/抽象クラスで出力ロジックを定義
- 実装クラスで各出力先に対応（Notion, Sheets, CSV, etc.）
- 設定ファイルまたは環境変数で出力先を切り替え

## 実装方針

### 技術スタック
**バックエンド: Go**
- Vertex AI REST APIを直接呼び出し
- パフォーマンス・並行処理に優れる
- Cloud Runへのデプロイが容易
- 型安全性が高い

**フロントエンド: Next.js**
- 英語表現の閲覧・検索UI
- 会議データのアップロード画面
- 抽出結果のプレビュー・承認機能
- レスポンシブデザイン（スマホでも閲覧可能）

### デプロイ構成
- **開発環境**: ローカルでGo API + Next.js dev server
- **本番環境**:
  - バックエンド: Cloud Run（Goコンテナ）
  - フロントエンド: Vercel または Cloud Run
  - または両方をCloud Runで統合

## セキュリティ対策案
- APIキー管理: 環境変数 or Secret Manager
- ログ設計: 処理履歴の監査ログ
- データ保護: 社内ネットワーク内での処理

## 確認事項
1. **出力先**: Notion API利用可能か？不可の場合、上記代替案のどれを採用するか？
   - Google Sheets API
   - CSV出力→手動インポート
   - ローカルSQLite DB
   - Markdown出力
2. **実行環境**: ローカルスクリプト vs Cloud Run、どちらから始めるか？
3. **モデル選択**: Gemini Flash（速い・安い）vs Pro（高精度）、どちらを優先するか？
4. **入力形式**: transcript/summaryはどのような形式で提供されるか？（テキスト、JSON、等）
5. **Google Workspace**: 会社でGoogle Workspaceを使用しているか？（Sheets API利用可否に関連）

## システムアーキテクチャ（Go + Next.js）

### 全体構成図
```
[ユーザー（PC/スマホ）]
    ↓
[Next.js Frontend]
    ↓ (REST API)
[Go Backend API]
    ├─→ [Vertex AI API]
    │   ├── Gemini (表現抽出)
    │   └── Text Embedding (類似判定)
    └─→ [出力先]
        ├── Notion API
        ├── Google Sheets API
        └── CSV Export
```

### プロジェクト構成

#### バックエンド（Go）
```
backend/
├── cmd/
│   └── api/
│       └── main.go                 # エントリーポイント
├── internal/
│   ├── config/
│   │   └── config.go              # 設定管理
│   ├── handler/
│   │   ├── upload.go              # transcript/summaryアップロード
│   │   ├── extract.go             # 表現抽出トリガー
│   │   └── expressions.go         # 表現の取得・検索
│   ├── service/
│   │   ├── preprocessor.go        # テキスト前処理
│   │   ├── extractor.go           # Geminiによる抽出
│   │   ├── embedding.go           # Embedding計算
│   │   └── deduplicator.go        # 重複チェック
│   ├── repository/
│   │   ├── repository.go          # リポジトリインターフェース
│   │   ├── notion.go              # Notion実装
│   │   ├── sheets.go              # Google Sheets実装
│   │   └── csv.go                 # CSV実装
│   ├── models/
│   │   └── expression.go          # データモデル
│   └── middleware/
│       ├── auth.go                # 認証ミドルウェア
│       └── logging.go             # ロギング
├── pkg/
│   └── vertexai/
│       ├── client.go              # Vertex AIクライアント
│       ├── gemini.go              # Gemini API
│       └── embedding.go           # Embedding API
├── go.mod
├── go.sum
└── Dockerfile
```

#### フロントエンド（Next.js）
```
frontend/
├── app/                           # App Router（Next.js 13+）
│   ├── layout.tsx                # ルートレイアウト
│   ├── page.tsx                  # ホーム（ダッシュボード）
│   ├── upload/
│   │   └── page.tsx              # transcript/summaryアップロード
│   ├── expressions/
│   │   ├── page.tsx              # 表現一覧
│   │   └── [id]/
│   │       └── page.tsx          # 表現詳細
│   └── api/                      # API Routes（オプション）
│       └── proxy/
│           └── route.ts          # Goバックエンドへのプロキシ
├── components/
│   ├── upload/
│   │   └── TranscriptUploader.tsx
│   ├── expressions/
│   │   ├── ExpressionList.tsx
│   │   ├── ExpressionCard.tsx
│   │   └── SearchBar.tsx
│   └── common/
│       ├── Header.tsx
│       └── Layout.tsx
├── lib/
│   ├── api.ts                    # Go APIクライアント
│   └── utils.ts
├── public/
├── package.json
└── next.config.js
```

### 主要コンポーネント

**バックエンド（Go）**
- REST API（Gin or Chi等のフレームワーク）
- Vertex AI連携（Gemini, Text Embedding）
- 出力先の抽象化（Interface）
  - Notion実装
  - Google Sheets実装
  - CSV実装
- 重複チェックロジック（String一致 + Embedding類似度）

**フロントエンド（Next.js）**
- transcript/summaryアップロード画面
- 抽出結果のプレビュー・編集
- 英語表現の一覧・検索画面
- レスポンシブデザイン

### 処理フロー
1. ユーザーがフロントエンドからtranscript/summaryをアップロード
2. GoバックエンドがVertex AI Geminiで英語表現を抽出
3. 各表現のEmbeddingを計算
4. 既存データと比較して重複チェック
5. 新規表現のみを選択した出力先（Notion/Sheets）に登録
6. 結果をフロントエンドに返却

## 次のステップ
- [ ] 確認事項の回答
- [ ] 出力先の確定（Notion/Sheets/CSV）
- [ ] データスキーマ設計
- [ ] 重複判定ロジックの詳細設計
- [ ] セキュリティ対策の詳細設計
- [ ] プロンプトエンジニアリング（Geminiへの指示文）
- [ ] MVP開発スケジュール策定
