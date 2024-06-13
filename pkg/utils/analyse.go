package utils

// Load repository ignore map from file
func loadExclusionRepos(filename string) (ExclusionRepos, error) {

	ignoreMap := make(ExclusionRepos)

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
