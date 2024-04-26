package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type GitHubContent struct {
	Name   string `json:"name"`
	Size   int    `json:"size"`
	Type   string `json:"type"`
	Url    string `json:"url"`
	GitUrl string `json:"git_url"`
}

func getRemoteBranchSize(owner, repo, branch, path, token string) (int, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, path)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	q := req.URL.Query()
	q.Add("ref", branch)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", "token "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("GitHub API request failed with status code: %d", resp.StatusCode)
	}

	var contents []GitHubContent

	if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
		return 0, err
	}

	totalSize := 0
	for _, content := range contents {
		if content.Type == "dir" {
			size, err := getRemoteBranchSize(owner, repo, branch, path+"/"+content.Name, token)
			if err != nil {
				return 0, err
			}
			totalSize += size
		} else {
			totalSize += content.Size
		}
	}

	return totalSize, nil
}

func main() {
	owner := "colussim"
	repo := "gcloc_m"
	branch := "main"
	token1 := "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"
	path := ""
	var startTime time.Time

	startTime = time.Now()
	size, err := getRemoteBranchSize(owner, repo, branch, path, token1)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60

	fmt.Printf("Estimated size of branch '%s': %d bytes\n", branch, size)
	fmt.Printf("\n\n✅ Time elapsed : %02d:%02d:%02d\n", hours, minutes, seconds)
}
