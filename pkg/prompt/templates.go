package prompt

import "fmt"

// ExtractPhrasesPrompt は熟語・慣用表現を抽出するプロンプト
func ExtractPhrasesPrompt(transcript string) string {
	return fmt.Sprintf(`以下の英語の会議transcriptから、熟語・慣用表現（2語以上の表現）を抽出してください。

# 抽出対象
- ビジネス英語の熟語（例: "circle back", "touch base", "reach out"）
- 技術用語の組み合わせ（例: "code review", "pull request", "API endpoint"）
- 一般的な慣用表現（例: "at the end of the day", "in terms of"）
- コロケーション（例: "make sense", "take a look"）

# 抽出しないもの
- 単語1つだけの表現
- 固有名詞（人名、会社名など）
- 通常の文法的な組み合わせ（例: "the meeting", "a project"）

# 出力形式
各熟語を1行につき1つ、以下のJSON形式で出力してください：
{"phrase": "熟語", "context": "その熟語が使われている元の文"}

例：
{"phrase": "circle back", "context": "Let's circle back on this next week."}
{"phrase": "code review", "context": "We need to do a code review before merging."}

重要: 各行は必ず正しいJSON形式にしてください。配列全体を[]で囲む必要はありません。

# Transcript
%s

# 抽出結果（JSON形式、1行1熟語）`, transcript)
}

// PrioritizeExpressionsPrompt は表現に優先度とカテゴリを付けるプロンプト
func PrioritizeExpressionsPrompt(expressions []string, transcript string) string {
	exprList := ""
	for i, expr := range expressions {
		exprList += fmt.Sprintf("%d. %s\n", i+1, expr)
	}

	return fmt.Sprintf(`以下の英語表現について、優先度とカテゴリを判定してください。

# 判定基準

## 優先度（1〜5）
5: ソフトウェアエンジニアとして必須の表現
4: 仕事・技術的な議論でよく使う表現
3: ビジネス英語として重要な表現
2: 知っておくと便利な表現
1: 日常会話レベルの表現

## カテゴリ
- engineering: ソフトウェアエンジニアリング関連
- business: ビジネス・仕事関連
- casual: 雑談・日常会話

# 表現リスト
%s

# 文脈（参考）
%s

# 出力形式
各表現を1行につき1つ、以下のJSON形式で出力してください：
{"expression": "表現", "meaning": "日本語の意味", "priority": 優先度数値, "category": "カテゴリ"}

例：
{"expression": "deprecate", "meaning": "非推奨にする", "priority": 5, "category": "engineering"}
{"expression": "touch base", "meaning": "連絡を取る", "priority": 3, "category": "business"}

重要: 各行は必ず正しいJSON形式にしてください。配列全体を[]で囲む必要はありません。

# 判定結果（JSON形式、1行1表現）`, exprList, transcript)
}
