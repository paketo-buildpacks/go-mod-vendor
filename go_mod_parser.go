package gomodvendor

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

type GoModParser struct{}

func NewGoModParser() GoModParser {
	return GoModParser{}
}

func (p GoModParser) ParseVersion(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to parse go.mod: %w", err)
	}

	versionNumber := `\d+\.\d+`
	expression := fmt.Sprintf(`go (%s)`, versionNumber)
	re := regexp.MustCompile(expression)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		matches := re.FindStringSubmatch(scanner.Text())

		if len(matches) == 2 {
			return fmt.Sprintf(">= %s", matches[1]), nil
		}
	}

	return "", nil
}
