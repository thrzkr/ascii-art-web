package z01

import "strings"

func ValidInput(text, banner string) bool {
	if text == "" || len([]rune(text)) > MaxTextLength || len(strings.Split(text, "\n")) > MaxLines {
		return false
	}
	if _, ok := bannerChecksums[banner]; !ok {
		return false
	}
	for _, r := range text {
		if r != '\n' && (r < 32 || r > 126) {
			return false
		}
	}
	return true
}

func normalizeText(text string) string {
	return strings.ReplaceAll(text, "\r\n", "\n")
}
