package getbibucket

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
)

type Projectc struct {
	Key         string `json:"key"`
	UUID        string `json:"uuid"`
	IsPrivate   bool   `json:"is_private"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Links       struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
}

type ProjectRepo struct {
	Type string `json:"type"`
	Key  string `json:"key"`
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

type ProjectcsResponse struct {
	Values  []Projectc `json:"values"`
	PageLen int        `json:"pagelen"`
	Size    int        `json:"size"`
	Page    int        `json:"page"`
	Next    string     `json:"next"`
}

type Reposc struct {
	Name        string      `json:"name"`
	Slug        string      `json:"slug"`
	Description string      `json:"description"`
	Size        int         `json:"size"`
	Language    string      `json:"language"`
	Project     ProjectRepo `json:"project"`
}

type ReposResponse struct {
	Values  []Reposc `json:"values"`
	Pagelen int      `json:"pagelen"`
	Size    int      `json:"size"`
	Page    int      `json:"page"`
	Next    string   `json:"next"`
}

type ExclusionList struct {
	Projectcs map[string]bool `json:"Projects"`
	Repos     map[string]bool `json:"repos"`
}

func loadExclusionList(filename string) (*ExclusionList, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	exclusionList := &ExclusionList{
		Projectcs: make(map[string]bool),
		Repos:     make(map[string]bool),
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "/")
		if len(parts) == 1 {
			// Get Projet
			exclusionList.Projectcs[parts[0]] = true
		} else if len(parts) == 2 {
			// Get Repos
			exclusionList.Repos[line] = true
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return exclusionList, nil
}

func GetReposProjectCloud(projects []Projectc, url, baseapi, apiver, accessToken, bitbucketURLBase, workspace string, nbRepos int, exclusionList *ExclusionList, spin *spinner.Spinner) ([]ProjectBranch, int, int) {

	var largestRepoSize int
	var largestRepoBranch string
	var importantBranches []ProjectBranch
	emptyRepo := 0

	for _, project := range projects {
		// Display Project Name
		//	fmt.Printf("✅ Projet: %s - Key: %s\n", project.Name, project.Key)
		largestRepoSize = 0
		largestRepoBranch = ""

		//'https://api.bitbucket.org/2.0/repositories/sonarsource?q=project.key="SAMPLES"'
		urlrepos := fmt.Sprintf("%s%s/repositories/%s?q=project.key='%s'", url, apiver, workspace, project.Key)

		// Get Repos for each Project

		repos, err := CloudAllRepos(urlrepos, accessToken, exclusionList)
		if err != nil {
			fmt.Println("\r❌ Get Repos for each Project:", err)
			spin.Stop()
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

			isEmpty, err := isRepositoryEmpty(project.Key, repo.Slug, accessToken, bitbucketURLBase, apiver)
			if err != nil {
				fmt.Printf("❌ Error when Testing if repo is empty %s: %v\n", repo.Name, err)
				spin.Stop()
				continue
			}
			if !isEmpty {

				urlrepos := fmt.Sprintf("%s%s%s/projects/%s/repos/%s/branches", url, baseapi, apiver, project.Key, repo.Slug)

				branches, err := fetchAllBranches(urlrepos, accessToken)
				if err != nil {
					fmt.Printf("❌ Error when retrieving branches for repo %s: %v\n", repo.Name, err)
					spin.Stop()
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
						spin.Stop()
						continue
					}
					// Display size of branch
					// fmt.Printf("\t\t\t✅ Size of branch: %s \n", sizemb)

					if size > largestRepoSize {
						largestRepoSize = size
						//largestRepoProject = project.Name
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

	return importantBranches, nbRepos, emptyRepo
}

func GetProjectBitbucketListCloud(url, baseapi, apiver, accessToken, workspace, exlusionfile, project, repo string) ([]Projectc, error) {

	/*var largestRepoSize int
	var totalSize int
	var largestRepoProject, largestRepoBranch string
	var importantBranches []ProjectBranch*/
	var projects []Projectc
	var exclusionList *ExclusionList
	var err1 error

	//totalSize = 0
	//nbRepos := 0
	//emptyRepo := 0
	bitbucketURL := fmt.Sprintf("%s%s/workspaces/%s/projects", url, apiver, workspace)

	// Get All Projects

	spin := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	spin.Prefix = "Get Projects... "
	spin.Color("green", "bold")
	spin.Start()

	if exlusionfile == "0" {
		exclusionList = &ExclusionList{
			Projectcs: make(map[string]bool),
			Repos:     make(map[string]bool),
		}

	} else {
		exclusionList, err1 = loadExclusionList(exlusionfile)
		if err1 != nil {
			fmt.Printf("\n❌ Error Read Exclusion File <%s>: %v", exlusionfile, err1)
			spin.Stop()
			return nil, err1
		}

	}

	if len(project) == 0 && len(repo) == 0 {

		projects, err1 = CloudAllProjects(bitbucketURL, accessToken, exclusionList)
		if err1 != nil {
			fmt.Println("\r❌ Error Get All Projects:", err1)
			spin.Stop()
			return nil, err1
		}
		spin.Stop()

		//importantBranches, nbRepos, emptyRepo = GetReposProject(projects, url, baseapi, apiver, accessToken, url,workspace, nbRepos, exclusionList, spin)

		//func GetReposProjectCloud(projects []Projectc, url, baseapi, apiver, accessToken, bitbucketURLBase,workspace string, nbRepos int, exclusionList *ExclusionList, spin *spinner.Spinner) ([]ProjectBranch, int, int) {

	}
	return projects, nil
}

func CloudAllProjects(url string, accessToken string, exclusionList *ExclusionList) ([]Projectc, error) {
	var allProjects []Projectc

	for url != "" {
		projectsResp, err := CloudProjects(url, accessToken, true)
		if err != nil {
			return nil, err
		}
		projectResponse := projectsResp.(*ProjectcsResponse)

		for _, project := range projectResponse.Values {
			if len(exclusionList.Projectcs) == 0 && len(exclusionList.Repos) == 0 {
				allProjects = append(allProjects, project)
			} else {
				if !isProjectExcluded(exclusionList, project.Key) {
					allProjects = append(allProjects, project)
				}
			}
		}

		url = projectResponse.Next
	}

	return allProjects, nil
}

func CloudProjects(url string, accessToken string, isProjectResponse bool) (interface{}, error) {
	var projectsResp interface{}

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

	if isProjectResponse {
		projectsResp = &ProjectcsResponse{}
	} else {
		projectsResp = &Projectc{}
	}

	err = json.Unmarshal(body, &projectsResp)
	if err != nil {
		return nil, err
	}

	return projectsResp, nil
}

func CloudAllRepos(url string, accessToken string, exclusionList *ExclusionList) ([]Reposc, error) {
	var allRepos []Reposc
	for url != "" {
		reposResp, err := CloudRepos(url, accessToken, true)
		if err != nil {
			return nil, err
		}
		ReposResponse := reposResp.(*ReposResponse)
		for _, repo := range ReposResponse.Values {
			KEYTEST := repo.Project.Key + "/" + repo.Slug

			if len(exclusionList.Projectcs) == 0 && len(exclusionList.Repos) == 0 {
				allRepos = append(allRepos, repo)
			} else {
				if !isRepoExcluded(exclusionList, KEYTEST) {
					allRepos = append(allRepos, repo)
				}
			}

		}

		url = ReposResponse.Next
	}
	return allRepos, nil
}

func CloudRepos(url string, accessToken string, isProjectResponse bool) (interface{}, error) {
	var reposResp interface{}

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

	if isProjectResponse {
		reposResp = &ReposResponse{}
	} else {
		reposResp = &Reposc{}
	}

	err = json.Unmarshal(body, &reposResp)
	if err != nil {
		return nil, err
	}

	return reposResp, nil

}

func isProjectExcluded(exclusionList *ExclusionList, project string) bool {
	_, excluded := exclusionList.Projectcs[project]
	return excluded
}

func isRepoExcluded(exclusionList *ExclusionList, repo string) bool {
	_, excluded := exclusionList.Repos[repo]
	return excluded
}
func isRepositoryEmpty(projectKey, repoSlug, accessToken, bitbucketURLBase, apiver string) (bool, error) {
	urlFiles := fmt.Sprintf("%srest/api/%s/projects/%s/repos/%s/browse", bitbucketURLBase, apiver, projectKey, repoSlug)
	filesResp, err := fetchFiles(urlFiles, accessToken)
	if err != nil {
		return false, fmt.Errorf("❌ Error when testing if repo : %s is empty - Function :%s - %v", repoSlug, "getbibucket-isRepositoryEmpty", err)
	}
	if filesResp.Children.Size == 0 {
		//fmt.Println("Repo %s is empty", repoSlug)

		return true, nil
	}

	return false, nil
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
