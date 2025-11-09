# 英語表現抽出システム - 要件整理

## 解決したい課題
Google Meetの英語会議で出てきた表現・単語・フレーズを学習したいが：
- 会議中にメモを取る余裕がない
- 後から文字起こしを見直すのは時間がかかる
- どの表現が重要か判断が難しい
- 同じような表現を何度も調べている

## ゴール
会議後に、重要な英語表現が自動抽出されて、学習しやすい形で保存されている状態

## 実現方法の選択肢

### 方針: CLIツールで最小限の実装

**基本フロー：**
```bash
# transcriptファイルを渡すと、英語表現を抽出してCSV出力
$ extract-expressions transcript.txt

# 出力例: expressions.csv
# Expression,Meaning,Category,Context,Importance
# "circle back","後で戻って確認する","ビジネス","Let's circle back on this next week",4
# ...

# オプションで出力先を指定
$ extract-expressions transcript.txt --output notion   # Notionに直接登録
$ extract-expressions transcript.txt --output sheets   # Google Sheetsに登録
$ extract-expressions transcript.txt --output csv     # CSV出力（デフォルト）
```

### 実装の選択肢

#### Option A: シンプルなPythonスクリプト + Claude API
**構成：**
- 単一のPythonスクリプト（100行程度）
- Claude API（Anthropic）を直接呼び出し
- 設定ファイルでプロンプトをカスタマイズ可能

**メリット：**
- 最小限の実装
- Claude APIは使い慣れている
- プロンプトの調整が容易
- すぐに動く

**デメリット：**
- Claude APIキーが必要（会社承認が必要かも）
- Embedding機能はない（重複チェックは文字列一致のみ）

#### Option B: Go CLI + Vertex AI
**構成：**
- Go製のCLIツール
- Vertex AI（Gemini + Text Embedding）
- 出力先プラグイン方式（Notion/Sheets/CSV）

**メリット：**
- 会社で承認されているVertex AIを使用
- Embedding使えるので高精度な重複チェック可能
- バイナリ1つで配布可能
- 型安全

**デメリット：**
- 実装量が多い（300-500行程度）
- 初回セットアップに時間がかかる

#### Option C: 完全手動（スクリプトなし）
**構成：**
- transcriptをテキストエディタで開く
- Claude Code or Web Claudeに貼り付け
- プロンプトで指示して出力をコピペ

**メリット：**
- 開発不要
- 今すぐ使える
- プロンプトを毎回調整できる

**デメリット：**
- 毎回手作業
- フォーマットが安定しない
- 重複チェック不可

## 推奨アプローチ

### Phase 1: まず手動で試す（今日）
1. transcriptのサンプルをClaudeに貼り付け
2. プロンプトを試行錯誤して、良い出力形式を見つける
3. 1-2回実際に使ってみる

### Phase 2: プロンプトが固まったらCLI化（1週間後）
1. Phase 1で確立したプロンプトをスクリプトに埋め込む
2. Option AまたはBのどちらかで実装
3. CSV出力だけでまず動かす

### Phase 3: 必要なら出力先を拡張（必要に応じて）
1. Notion API連携を追加
2. 重複チェック機能を追加
3. その他便利機能

## 議論ポイント

### 1. どのくらいの頻度で会議がある？
- 週1回程度 → 手動でも十分かも
- 週3回以上 → CLI化の価値あり
- 毎日 → 完全自動化も検討

### 2. transcript 1つあたりの量は？
- 短い（5分の会議） → 手動でもOK
- 中くらい（30分） → CLIが便利
- 長い（1時間以上） → CLI必須

### 3. Claude API vs Vertex AI、どっちが使える？
- Claude APIが使える → Option A（最速）
- Vertex AIのみ → Option B
- 両方ダメ → Web Claudeで手動

### 4. いつから使い始めたい？
- 今日 → Option C（手動）で試す
- 今週中 → Option A
- 来週以降 → Option B

## ユースケース整理（ユーザー回答）

### 1. 使用頻度
- **毎日（月〜金）**
- 30分〜1時間のMTG
- → CLI化は必須レベル

### 2. 抽出したい内容
- 単語（プログラムで全抽出）
- 熟語・慣用表現（AIで抽出）
- すべてに優先度付け（AIで判定）

### 3. 処理フロー
```
transcript
  ↓
[1. 単語抽出（プログラム）]
  ↓
[2. 熟語・慣用表現抽出（AI）]
  ↓
[3. 優先度付け（AI）]
  ↓
[4. DB保存（SQLite or その他）]
  ↓
[5. 出力（切り替え可能）]
  → Notion / Sheets / CSV / その他
```

### 4. 出力先
- 切り替え可能にしたい
- 候補: Notion, Google Sheets, CSV, Anki, その他

## 設計上の重要ポイント

### 中間DB層の導入
- 抽出結果を一旦DBに保存
- 出力先を自由に切り替えられる
- 重複チェックもDB層で実施
- 後から再出力も可能

### 2段階のAI処理
1. **熟語・慣用表現の抽出**:
   - 単語抽出（プログラム）だけでは拾えない表現をAIで検出
   - 例: "circle back", "touch base", "at the end of the day"

2. **優先度付け**:
   - すべての表現に対して重要度をAIが判定
   - 学習優先度を自動スコアリング

### 単語抽出のアプローチ
**疑問点：**
- 単語の「全抽出」とは？
  - 品詞でフィルタ？（名詞・動詞・形容詞のみ）
  - 頻度でフィルタ？（低頻度のみ）
  - 既知語を除外？

**例：**
transcript: "We need to deprecate this API endpoint by Q2."
- 単語抽出 → deprecate, API, endpoint, Q2?
- 熟語抽出 → "deprecate API", "by Q2"?

## 決定事項

### 1. 単語抽出の基準
- **全単語を抽出**
- フィルタリングはAIの優先度付けで対応

### 2. DBの選択
- **SQLite**
- ローカルファイル、シンプル
- 重複チェックや履歴管理に使用

### 3. 優先度付けの基準（3段階）
1. **高**: ソフトウェアエンジニアとして働くうえで必要
2. **中**: 仕事・議論で必要
3. **低**: 雑談レベル

### 4. データスキーマ（案）
```sql
CREATE TABLE expressions (
    id INTEGER PRIMARY KEY,
    expression TEXT NOT NULL,      -- 単語または熟語
    type TEXT,                      -- 'word' or 'phrase'
    meaning TEXT,                   -- 日本語の意味
    context TEXT,                   -- 使用された文脈
    priority INTEGER,               -- 1(低) ~ 5(高)
    category TEXT,                  -- 'engineering' / 'business' / 'casual'
    source_meeting TEXT,            -- 会議名/日付
    created_at TIMESTAMP,
    UNIQUE(expression)              -- 重複防止
);
```

## 次のステップ

### Phase 1: プロトタイプで検証（最優先）
1. 実際のtranscriptサンプルを1つ用意
2. 手動でフロー全体を試す：
   - 全単語抽出（プログラム）
   - 熟語抽出（AIにプロンプト）
   - 優先度付け（AIにプロンプト）
3. 出力フォーマットを確認
4. プロンプトを調整・確定

### Phase 2: CLI実装
1. Phase 1で確定したプロンプトを使用
2. Go製CLIツール作成
3. SQLite連携
4. CSV出力まで実装

### Phase 3: 出力先拡張
1. Notion連携
2. その他の出力先

## 次回の議論再開時にやること

### 1. Phase 1の実施（プロトタイプ検証）
- [ ] transcriptサンプルを用意
- [ ] 全単語抽出（プログラム/手動）
- [ ] 熟語抽出プロンプトの試作・検証
- [ ] 優先度付けプロンプトの試作・検証
- [ ] 出力フォーマットの確認

### 2. プロンプトの確定
- [ ] 熟語・慣用表現抽出用プロンプト
- [ ] 優先度付け用プロンプト（3段階: 高/中/低）

### 3. CLI実装の検討
- [x] Go vs Python の最終決定 → **Go**
- [x] LLMモデルの選択 → **Claude Sonnet 4.5**
- [x] プロバイダー選択
  - 開発環境: Anthropic Claude API（直接）
  - 本番環境: Vertex AI経由でClaude Sonnet 4.5
  - → 環境変数で切り替え可能にする
- [ ] 実装スケジュールの策定

## 補足メモ

### なぜこのアプローチか
- 毎日使う（月〜金）→ CLI化必須
- 30分〜1時間のMTG → 処理の自動化が重要
- DB層を挟むことで出力先を柔軟に切り替え可能

### 関連ドキュメント
- `english_expression_system_design.md`: 初期のWebアプリ設計案（現在は方針変更してCLI化）
- `requirements_discussion.md`: 本ファイル（最新の要件・設計）
