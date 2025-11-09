package extractor

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/mamyudapao/learn-by-transcript/internal/models"
)

// WordExtractor は単語抽出を行う
type WordExtractor struct {
	minWordLength int
}

// NewWordExtractor は新しいWordExtractorを作成
func NewWordExtractor() *WordExtractor {
	return &WordExtractor{
		minWordLength: 2, // 最小2文字以上の単語のみ抽出
	}
}

// Extract はテキストから単語を抽出
func (e *WordExtractor) Extract(text string) []string {
	// テキストをトークン化
	tokens := tokenize(text)

	// 重複除去のためのmap
	wordSet := make(map[string]bool)

	for _, token := range tokens {
		// 英単語のみを抽出（アルファベットのみで構成されているか確認）
		if !isEnglishWord(token) {
			continue
		}

		// 小文字化
		word := strings.ToLower(token)

		// 最小文字数フィルタ
		if len(word) < e.minWordLength {
			continue
		}

		// ストップワード除外（基本的な接続詞・冠詞など）
		if isStopWord(word) {
			continue
		}

		wordSet[word] = true
	}

	// mapからスライスに変換
	words := make([]string, 0, len(wordSet))
	for word := range wordSet {
		words = append(words, word)
	}

	return words
}

// ExtractWithContext はテキストから単語と文脈を抽出
func (e *WordExtractor) ExtractWithContext(text string) []*models.Expression {
	words := e.Extract(text)

	expressions := make([]*models.Expression, 0, len(words))
	for _, word := range words {
		// 単語が使われている文脈を抽出（その単語を含む文）
		context := findContext(text, word)

		expr := &models.Expression{
			Expression: word,
			Type:       string(models.TypeWord),
			Context:    context,
		}
		expressions = append(expressions, expr)
	}

	return expressions
}

// tokenize はテキストをトークンに分割
func tokenize(text string) []string {
	// 記号や空白で分割
	re := regexp.MustCompile(`[\s\p{P}]+`)
	tokens := re.Split(text, -1)

	// 空文字を除去
	result := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if token != "" {
			result = append(result, token)
		}
	}

	return result
}

// isEnglishWord は英単語かどうかを判定
func isEnglishWord(word string) bool {
	if len(word) == 0 {
		return false
	}

	for _, r := range word {
		// アルファベット以外が含まれていたらfalse
		if !unicode.IsLetter(r) {
			return false
		}
		// 英語のアルファベット範囲外ならfalse
		if r < 'A' || (r > 'Z' && r < 'a') || r > 'z' {
			return false
		}
	}

	return true
}

// isStopWord はストップワード（除外する一般的な単語）かどうかを判定
func isStopWord(word string) bool {
	// 基本的なストップワードリスト
	stopWords := map[string]bool{
		"a": true, "an": true, "the": true,
		"is": true, "am": true, "are": true, "was": true, "were": true, "be": true, "been": true, "being": true,
		"have": true, "has": true, "had": true, "having": true,
		"do": true, "does": true, "did": true, "doing": true,
		"will": true, "would": true, "should": true, "could": true, "may": true, "might": true, "must": true, "can": true,
		"and": true, "or": true, "but": true, "nor": true,
		"if": true, "then": true, "else": true,
		"of": true, "at": true, "by": true, "for": true, "with": true, "about": true, "against": true,
		"between": true, "into": true, "through": true, "during": true, "before": true, "after": true,
		"above": true, "below": true, "to": true, "from": true, "up": true, "down": true, "in": true,
		"out": true, "on": true, "off": true, "over": true, "under": true,
		"i": true, "you": true, "he": true, "she": true, "it": true, "we": true, "they": true,
		"me": true, "him": true, "her": true, "us": true, "them": true,
		"my": true, "your": true, "his": true, "its": true, "our": true, "their": true,
		"this": true, "that": true, "these": true, "those": true,
		"what": true, "which": true, "who": true, "when": true, "where": true, "why": true, "how": true,
		"all": true, "each": true, "every": true, "both": true, "few": true, "more": true, "most": true, "other": true,
		"some": true, "such": true, "no": true, "not": true, "only": true, "own": true, "same": true, "so": true,
		"than": true, "too": true, "very": true, "yes": true,
	}

	return stopWords[word]
}

// findContext は単語が使われている文脈（文）を抽出
func findContext(text, word string) string {
	// テキストを文に分割
	sentences := splitIntoSentences(text)

	// 単語を含む最初の文を返す
	wordLower := strings.ToLower(word)
	for _, sentence := range sentences {
		if strings.Contains(strings.ToLower(sentence), wordLower) {
			// 文を整形して返す
			return strings.TrimSpace(sentence)
		}
	}

	// 見つからない場合は空文字列
	return ""
}

// splitIntoSentences はテキストを文に分割
func splitIntoSentences(text string) []string {
	// ピリオド、疑問符、感嘆符で分割
	re := regexp.MustCompile(`[.!?]+`)
	sentences := re.Split(text, -1)

	result := make([]string, 0, len(sentences))
	for _, sentence := range sentences {
		trimmed := strings.TrimSpace(sentence)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
