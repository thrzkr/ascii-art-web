package z01

import (
	"os"
	"path/filepath"
	"strings"
)

func GenerateASCII(text, banner string) (string, error) {
	content, err := os.ReadFile(filepath.Join("banners", banner+".txt"))
	if err != nil {
		return "", err
	}

	lines := strings.Split(strings.ReplaceAll(string(content), "\r\n", "\n"), "\n")
	if len(lines) != 855 {
		return "", os.ErrInvalid
	}

	var result strings.Builder
	for _, line := range strings.Split(normalizeText(text), "\n") {
		for row := 0; row < 8; row++ {
			for _, char := range line {
				index := (int(char)-32)*9 + row + 1
				result.WriteString(lines[index])
			}
			result.WriteByte('\n')
		}
	}
	return result.String(), nil
}
