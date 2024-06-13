package utils

import (
	"bufio"
	"os"
	"strings"
)

type ExclusionRepos1 map[string]bool

// Load repository ignore map from file
func LoadExclusionRepos1(filename string) (ExclusionRepos1, error) {

	ignoreMap := make(ExclusionRepos1)

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		repoName := strings.TrimSpace(scanner.Text())
		if repoName != "" {
			ignoreMap[repoName] = true

		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ignoreMap, nil
}
