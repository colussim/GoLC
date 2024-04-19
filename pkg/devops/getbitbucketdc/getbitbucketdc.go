package getbibucketdc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/briandowns/spinner"
)

type ProjectBranch struct {
	ProjectKey  string
	RepoSlug    string
	MainBranch  string
	LargestSize int
}

type ProjectResponse struct {
	Size          int       `json:"size"`
	Limit         int       `json:"limit"`
	IsLastPage    bool      `json:"isLastPage"`
	Values        []Project `json:"values"`
	Start         int       `json:"start"`
	NextPageStart int       `json:"nextPageStart"`
}

type Project struct {
	Key   string `json:"key"`
	Name  string `json:"name"`
	Links struct {
		Self []struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
}

type RepoResponse struct {
	Size          int    `json:"size"`
	Limit         int    `json:"limit"`
	IsLastPage    bool   `json:"isLastPage"`
	Values        []Repo `json:"values"`
	Start         int    `json:"start"`
	NextPageStart int    `json:"nextPageStart"`
}

type Repo struct {
	Slug    string `json:"slug"`
	Name    string `json:"name"`
	Project struct {
		Key string `json:"key"`
	} `json:"project"`
	Links struct {
		Self []struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
}

type BranchResponse struct {
	Size          int      `json:"size"`
	Limit         int      `json:"limit"`
	IsLastPage    bool     `json:"isLastPage"`
	Values        []Branch `json:"values"`
	Start         int      `json:"start"`
	NextPageStart int      `json:"nextPageStart"`
}

type Branch struct {
	Name       string `json:"displayId"`
	Statistics struct {
		Size string `json:"size"`
	} `json:"statistics"`
}

type FileResponse struct {
	Path          Path     `json:"path"`
	Revision      string   `json:"revision"`
	Children      Children `json:"children"`
	Start         int      `json:"start"`
	IsLastPage    bool     `json:"isLastPage"`
	NextPageStart int      `json:"nextPageStart"`
}

type Path struct {
	Components []string `json:"components"`
	Name       string   `json:"name"`
	ToString   string   `json:"toString"`
}

type Children struct {
	Size   int    `json:"size"`
	Limit  int    `json:"limit"`
	Values []File `json:"values"`
}

type File struct {
	Path      Path   `json:"path"`
	ContentID string `json:"contentId"`
	Type      string `json:"type"`
	Size      int    `json:"size"`
}

func GetProjectBitbucketList(url, baseapi, apiver, accessToken string) ([]ProjectBranch, error) {

	var largestRepoSize int
	var largestRepoProject, largestRepoBranch string
	var importantBranches []ProjectBranch

	totalSize := 0
	nbRepos := 0
	emptyRepo := 0
	bitbucketURLBase := "http://ec2-18-194-139-24.eu-central-1.compute.amazonaws.com:7990/"
	bitbucketURL := fmt.Sprintf("%s%s%s/project", url, baseapi, apiver)

	// Get All Projects

	spin := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	spin.Prefix = "Get Projects... "
	spin.Color("green", "bold")
	spin.Start()

	projects, err := fetchAllProjects(bitbucketURL, accessToken)
	if err != nil {
		fmt.Println("❌ Error Get All Projects:", err)
		return
	}
	spin.Stop()
	fmt.Printf("\r✅ Number of projects: %d\n", len(projects))

	// Get Repos for each Project
	spin.Prefix = "Get Repos for each Project..."
	spin.Start()
	for _, project := range projects {
		// Display Project Name
		//	fmt.Printf("✅ Projet: %s - Key: %s\n", project.Name, project.Key)
		largestRepoSize = 0
		largestRepoBranch = ""
		largestRepoProject = ""

		urlrepos := fmt.Sprintf("%s%s/%s/projects/%s/repos", url, baseapi, apiver, project.Key)

		// Get Repos for each Project

		repos, err := fetchAllRepos(urlrepos, accessToken)
		if err != nil {
			fmt.Println("❌ Get Repos for each Project:", err)
			continue
		}
		spin.Stop()
		nbRepos += len(repos)
		// Display size of Repo
		// fmt.Printf("\t✅ Number of repos: %d\n", len(repos))

		// Get repo with largest branch size
		spin.Prefix = "Analysis of repos Finds the largest branch size..."
		spin.Start()
		for _, repo := range repos {

			//  isRepositoryEmpty(projectKey, repoSlug, accessToken, bitbucketURLBase, apiver string)
			isEmpty, err := isRepositoryEmpty(project.Key, repo.Slug, accessToken, bitbucketURLBase, apiver)
			if err != nil {
				fmt.Printf("❌ Error when Testing if repo is empty %s: %v\n", repo.Name, err)
				continue
			}
			if !isEmpty {

				urlrepos := fmt.Sprintf("%s%s%s/projects/%s/repos/%s/branches", url, baseapi, apiver, project.Key, repo.Slug)

				branches, err := fetchAllBranches(urlrepos, accessToken)
				if err != nil {
					fmt.Printf("❌ Error when retrieving branches for repo %s: %v\n", repo.Name, err)
					continue
				}
				// Display Number of branches by repo
				// fmt.Printf("\r\t✅ Repo: %s - Number of branches: %d\n", repo.Name, len(branches))

				// Finding the branch with the largest size
				for _, branch := range branches {
					// Display Branch name
					// fmt.Printf("\t\t✅ Branche: %s\n", branch.Name)

					size, err := fetchBranchSize(project.Key, repo.Slug, branch.Name, accessToken, url, apiver)
					if err != nil {
						fmt.Println("❌ Error retrieving branch size:", err)
						continue
					}
					// Display size of branch
					// fmt.Printf("\t\t\t✅ Size of branch: %s \n", sizemb)

					if size > largestRepoSize {
						largestRepoSize = size
						largestRepoProject = project.Name
						largestRepoBranch = branch.Name
					}

				}
				importantBranches = append(importantBranches, ProjectBranch{
					ProjectKey:  project.Key,
					RepoSlug:    repo.Slug,
					MainBranch:  largestRepoBranch,
					LargestSize: largestRepoSize,
				})
			} else {
				emptyRepo++
			}
		}
		spin.Stop()

	}
	largestRepoSize = 0
	largestRepoBranch = ""
	largestRepoProject = ""

	for _, branch := range importantBranches {
		//	fmt.Printf("Projet: %s, Repo: %s, Branche: %s, Taille: %d\n", branch.ProjectKey, branch.RepoSlug, branch.MainBranch, branch.LargestSize)

		if branch.LargestSize > largestRepoSize {
			largestRepoSize = branch.LargestSize
			largestRepoBranch = branch.MainBranch
			largestRepoProject = branch.ProjectKey
		}
		totalSize += branch.LargestSize
	}
	totalSizeMB := float64(totalSize) / 1048576.0
	largestRepoSizeMB := float64(largestRepoSize) / 1048576.0

	fmt.Printf("\r✅ The largest repo is in the %s with the branch %s and a size of %2.f Mo\n", largestRepoProject, largestRepoBranch, largestRepoSizeMB)
	fmt.Printf("\r✅ Total size of your organization's repositories: %2.f Mo\n", totalSizeMB)
	fmt.Printf("\r✅ Total repositories analyzed: %d - Find empty : %d\n", nbRepos-emptyRepo, emptyRepo)

	return importantBranches, nil
}

func fetchAllProjects(url string, accessToken string) ([]Project, error) {
	var allProjects []Project
	for {
		projectsResp, err := fetchProjects(url, accessToken)
		if err != nil {
			return nil, err
		}
		allProjects = append(allProjects, projectsResp.Values...)
		if projectsResp.IsLastPage {
			break
		}
		url = fmt.Sprintf("%s?start=%d", url, projectsResp.NextPageStart)
	}
	return allProjects, nil
}

func fetchProjects(url string, accessToken string) (*ProjectResponse, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var projectsResp ProjectResponse
	err = json.Unmarshal(body, &projectsResp)
	if err != nil {
		return nil, err
	}

	return &projectsResp, nil
}

func fetchAllRepos(url string, accessToken string) ([]Repo, error) {
	var allRepos []Repo
	for {
		reposResp, err := fetchRepos(url, accessToken)
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, reposResp.Values...)
		if reposResp.IsLastPage {
			break
		}
		url = fmt.Sprintf("%s?start=%d", url, reposResp.NextPageStart)
	}
	return allRepos, nil
}

func fetchRepos(url string, accessToken string) (*RepoResponse, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var reposResp RepoResponse
	err = json.Unmarshal(body, &reposResp)
	if err != nil {
		return nil, err
	}

	return &reposResp, nil
}

func fetchAllBranches(url string, accessToken string) ([]Branch, error) {
	var allBranches []Branch
	for {
		branchesResp, err := fetchBranches(url, accessToken)
		if err != nil {
			return nil, err
		}
		allBranches = append(allBranches, branchesResp.Values...)
		if branchesResp.IsLastPage {
			break
		}
		url = fmt.Sprintf("%s?start=%d", url, branchesResp.NextPageStart)
	}
	return allBranches, nil
}

func fetchBranches(url string, accessToken string) (*BranchResponse, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var branchesResp BranchResponse
	err = json.Unmarshal(body, &branchesResp)
	if err != nil {
		return nil, err
	}

	return &branchesResp, nil
}
func calculateTotalSize(files []File) int {
	totalSize := 0
	for _, file := range files {
		totalSize += file.Size
	}
	return totalSize
}

func fetchAllFiles(url string, accessToken string) ([]File, error) {
	var allFiles []File
	for {
		filesResp, err := fetchFiles(url, accessToken)
		if err != nil {
			return nil, err
		}
		allFiles = append(allFiles, filesResp.Children.Values...)
		if filesResp.IsLastPage {
			break
		}
		url = fmt.Sprintf("%s?start=%d", url, filesResp.NextPageStart)
	}
	return allFiles, nil
}

func fetchFiles(url string, accessToken string) (*FileResponse, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var filesResp FileResponse
	err = json.Unmarshal(body, &filesResp)
	if err != nil {
		return nil, err
	}

	return &filesResp, nil
}

func fetchBranchSize(projectKey string, repoSlug string, branchName string, accessToken string, bitbucketURLBase string, apiver string) (int, error) {
	url := fmt.Sprintf("%srest/api/%s/projects/%s/repos/%s/browse?at=refs/heads/%s", bitbucketURLBase, apiver, projectKey, repoSlug, branchName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var filesResp FileResponse
	err = json.Unmarshal(body, &filesResp)
	if err != nil {
		return 0, err
	}

	totalSize := 0
	for _, file := range filesResp.Children.Values {
		totalSize += file.Size
	}

	return totalSize, nil
}

func isRepositoryEmpty(projectKey, repoSlug, accessToken, bitbucketURLBase, apiver string) (bool, error) {
	urlFiles := fmt.Sprintf("%srest/api/%s/projects/%s/repos/%s/browse", bitbucketURLBase, apiver, projectKey, repoSlug)
	filesResp, err := fetchFiles(urlFiles, accessToken)
	if err != nil {
		return false, fmt.Errorf("❌ Error when testing if repo : %s is empty - Function :%s - %v", repoSlug, "getbibucketdc-isRepositoryEmpty", err)
	}
	if filesResp.Children.Size == 0 {
		//fmt.Println("Repo %s is empty", repoSlug)

		return true, nil
	}

	return false, nil
}
