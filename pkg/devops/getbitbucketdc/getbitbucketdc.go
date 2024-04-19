package getbibucketdc

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type Repository struct {
	Slug  string `json:"slug"`
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Links struct {
		Clone []struct {
			Href string `json:"href"`
			Name string `json:"name"`
		} `json:"clone"`
	} `json:"links"`
}

type Branch struct {
	Name string `json:"name"`
}

type ProjectsResponse struct {
	Values []Project `json:"values"`
}

type Project struct {
	KEY    string `json:"key"`
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Public bool   `json:"public"`
	Type   string `json:"type"`
	Links  Links  `json:"links"`
}

type Links struct {
	Self []SelfLink `json:"self"`
}

type SelfLink struct {
	Href string `json:"href"`
}

// Get Project
func GetProjectBitbucketList(url, accessToken string) ([]Project, error) {

	var projects []Project

	//url = fmt.Sprintf("https://%s/orgs/%s/repos?type=all&recurse_submodules=false", baseURL, organization)
	urlc := url + "rest/api/1.0/projects"
	//fmt.Print("urlc:", urlc)

	page := 1000
	for {
		pro, nextPageURL, err := FetchProjectsBitbucket(urlc, page, accessToken)
		if err != nil {
			log.Fatal(err)
		}
		projects = append(projects, pro...)
		if nextPageURL == "" {
			break
		}
		page++
	}

	return projects, nil
}

// Browsing project
func FetchProjectsBitbucket(url string, page int, accessToken string) ([]Project, string, error) {

	var projectsResponse ProjectsResponse

	url1 := fmt.Sprintf("%s?limit=%d", url, page)

	req, _ := http.NewRequest("GET", url1, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print("-- Stack: getbitbucket.FetchProjects Request API -- ")
		return nil, "", err
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if err := json.Unmarshal(body, &projectsResponse); err != nil {
		return nil, "", err
	}

	// Extract projects from ProjectsResponse
	projects := projectsResponse.Values

	nextPageURL := ""
	linkHeader := resp.Header.Get("Link")

	if linkHeader != "" {
		links := strings.Split(linkHeader, ",")
		for _, link := range links {
			parts := strings.Split(strings.TrimSpace(link), ";")
			if len(parts) == 2 && strings.TrimSpace(parts[1]) == `rel="next"` {
				nextPageURL = strings.Trim(parts[0], "<>")
			}
		}
	}

	return projects, nextPageURL, nil

}

// ProjectRepos fetches the list of repositories in a project
func FetchProjectsReposBitbucket(url, token string, page int) ([]Repository, string, error) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	var repos []Repository
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, "", err
	}

	nextPageURL := ""
	linkHeader := resp.Header.Get("Link")

	if linkHeader != "" {
		links := strings.Split(linkHeader, ",")
		for _, link := range links {
			parts := strings.Split(strings.TrimSpace(link), ";")
			if len(parts) == 2 && strings.TrimSpace(parts[1]) == `rel="next"` {
				nextPageURL = strings.Trim(parts[0], "<>")
			}
		}
	}

	return repos, nextPageURL, nil

}

// GetHTTPCloneURL extracts the HTTP clone URL from the repository links
func GetHTTPCloneURL(repo Repository) string {
	for _, clone := range repo.Links.Clone {
		if strings.ToLower(clone.Name) == "http" {
			return clone.Href
		}
	}
	return ""
}

// Get Project
func GetRepotBitbucketList(url, accessToken, projectKey string) ([]Repository, error) {

	var repos []Repository
	page := 1000

	urlc := fmt.Sprintf("%srest/api/1.0/projects/%s/repos?limit=%d", url, projectKey, page)

	//fmt.Print("urlc:", urlc)

	for {
		pro, nextPageURL, err := FetchProjectsReposBitbucket(urlc, accessToken, page)
		if err != nil {
			log.Fatal(err)
		}
		repos = append(repos, pro...)
		if nextPageURL == "" {
			break
		}
		page++
	}

	return repos, nil
}
