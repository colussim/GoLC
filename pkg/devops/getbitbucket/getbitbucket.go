package getbibucket

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/briandowns/spinner"
)

type ProjectBranch struct {
	ProjectKey  string
	RepoSlug    string
	MainBranch  string
	LargestSize int
}
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

type Branch struct {
	Name       string `json:"displayId"`
	Statistics struct {
		Size string `json:"size"`
	} `json:"statistics"`
}

type BranchResponse struct {
	Size          int      `json:"size"`
	Limit         int      `json:"limit"`
	IsLastPage    bool     `json:"isLastPage"`
	Values        []Branch `json:"values"`
	Start         int      `json:"start"`
	NextPageStart int      `json:"nextPageStart"`
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

type Commit struct {
	Hash  string `json:"hash"`
	Links struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
	Type string `json:"type"`
}

type FileInfo struct {
	Path        string `json:"path"`
	Commit      Commit `json:"commit"`
	Type        string `json:"type"`
	Attributes  []int  `json:"attributes,omitempty"`
	EscapedPath string `json:"escaped_path,omitempty"`
	Size        int    `json:"size,omitempty"`
	MimeType    string `json:"mimetype,omitempty"`
	Links       struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
}

type Response1 struct {
	Values  []FileInfo `json:"values"`
	Pagelen int        `json:"pagelen"`
	Page    int        `json:"page"`
	Next    string     `json:"next"`
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

type ExclusionList struct {
	Projectcs map[string]bool `json:"Projects"`
	Repos     map[string]bool `json:"repos"`
}

func formatSize(size int64) string {
	const (
		byteSize = 1.0
		kiloSize = 1024.0
		megaSize = 1024.0 * kiloSize
		gigaSize = 1024.0 * megaSize
	)

	switch {
	case size < kiloSize:
		return fmt.Sprintf("%d B", size)
	case size < megaSize:
		return fmt.Sprintf("%.2f KB", float64(size)/kiloSize)
	case size < gigaSize:
		return fmt.Sprintf("%.2f MB", float64(size)/megaSize)
	default:
		return fmt.Sprintf("%.2f GB", float64(size)/gigaSize)
	}
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
		urlrepos := fmt.Sprintf("%s%s/repositories/%s?q=project.key=\"%s\"", url, apiver, workspace, project.Key)

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
		//fmt.Printf("\t✅ Number of repos: %d\n", len(repos))

		// Get repo with largest branch size
		spin.Prefix = "Analysis of repos Finds the largest branch size..."
		spin.Start()
		for _, repo := range repos {

			isEmpty, err := isRepositoryEmpty(workspace, repo.Slug, accessToken, bitbucketURLBase)
			if err != nil {
				fmt.Printf("❌ Error when Testing if repo is empty %s: %v\n", repo.Name, err)
				spin.Stop()
				continue
			}
			fmt.Println("The repo is :", isEmpty)
			//	os.Exit(1)
			if !isEmpty {

				//https://api.bitbucket.org/2.0/repositories/{workspace}/{repo_slug}/refs/branches

				//urlrepos := fmt.Sprintf("%s%s/repositories/%s/repos/refs/branches", url, apiver, workspace, repo.Slug)

				urlrepos := fmt.Sprintf("%s%s/repositories/%s/%s/refs/branches", url, apiver, workspace, repo.Slug)

				branches, err := CloudAllBranches(urlrepos, accessToken)
				if err != nil {
					fmt.Printf("❌ Error when retrieving branches for repo %s: %v\n", repo.Name, err)
					spin.Stop()
					continue
				}
				// Display Number of branches by repo
				fmt.Printf("\r\t✅ Repo: %s - Number of branches: %d\n", repo.Name, len(branches))

				// Finding the branch with the largest size

				for _, branch := range branches {
					// Display Branch name
					fmt.Printf("\t\t✅ Branche: %s\n", branch.Name)

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

// func GetProjectBitbucketListCloud(url, baseapi, apiver, accessToken, workspace, exlusionfile, project, repo string) ([]Projectc, error) {
func GetProjectBitbucketListCloud(url, baseapi, apiver, accessToken, workspace, exlusionfile, project, repo string) ([]ProjectBranch, error) {

	var largestRepoSize int
	var totalSize int
	var largestRepoProject, largestRepoBranch, largesRepo string
	var importantBranches []ProjectBranch
	var projects []Projectc
	var exclusionList *ExclusionList
	var err1 error
	var emptyRepo int

	//totalSize = 0
	nbRepos := 0
	//emptyRepo := 0
	bitbucketURLBase := fmt.Sprintf("%s/%s/", url, apiver)
	bitbucketURL := fmt.Sprintf("%s%s/workspaces/%s/projects", url, apiver, workspace)

	fmt.Print("\n🔎 Analysis of devops platform objects ...\n")

	spin := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	spin.Prefix = "Get Projects... "
	spin.Color("green", "bold")

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

		importantBranches, nbRepos, emptyRepo = GetReposProjectCloud(projects, url, baseapi, apiver, accessToken, bitbucketURLBase, workspace, nbRepos, exclusionList, spin)

		//importantBranches, nbRepos, emptyRepo = GetReposProject(projects, url, baseapi, apiver, accessToken, url,workspace, nbRepos, exclusionList, spin)

		//func GetReposProjectCloud(projects []Projectc, url, baseapi, apiver, accessToken, bitbucketURLBase,workspace string, nbRepos int, exclusionList *ExclusionList, spin *spinner.Spinner) ([]ProjectBranch, int, int) {

	}

	largestRepoSize = 0
	largestRepoBranch = ""
	largestRepoProject = ""
	largesRepo = ""

	for _, branch := range importantBranches {
		//	fmt.Printf("Projet: %s, Repo: %s, Branche: %s, Taille: %d\n", branch.ProjectKey, branch.RepoSlug, branch.MainBranch, branch.LargestSize)

		if branch.LargestSize > largestRepoSize {
			largestRepoSize = branch.LargestSize
			largestRepoBranch = branch.MainBranch
			largestRepoProject = branch.ProjectKey
			largesRepo = branch.RepoSlug
		}
		totalSize += branch.LargestSize
	}
	totalSizeMB := formatSize(int64(totalSize))
	largestRepoSizeMB := formatSize(int64(largestRepoSize))

	fmt.Printf("\n✅ The largest repo is <%s> in the project <%s> with the branch <%s> and a size of %s\n", largesRepo, largestRepoProject, largestRepoBranch, largestRepoSizeMB)
	fmt.Printf("\r✅ Total size of your organization's repositories: %s\n", totalSizeMB)
	fmt.Printf("\r✅ Total repositories analyzed: %d - Find empty : %d\n", nbRepos-emptyRepo, emptyRepo)
	os.Exit(1)

	return importantBranches, nil
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

func CloudAllBranches(url string, accessToken string) ([]Branch, error) {
	var allBranches []Branch
	for {
		branchesResp, err := CloudBranches(url, accessToken)
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

func CloudBranches(url string, accessToken string) (*BranchResponse, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	//fmt.Println(url)
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

func isProjectExcluded(exclusionList *ExclusionList, project string) bool {
	_, excluded := exclusionList.Projectcs[project]
	return excluded
}

func isRepoExcluded(exclusionList *ExclusionList, repo string) bool {
	_, excluded := exclusionList.Repos[repo]
	return excluded
}
func isRepositoryEmpty(workspace, repoSlug, accessToken, bitbucketURLBase string) (bool, error) {

	urlFiles := fmt.Sprintf("%srepositories/%s/%s/src/main/", bitbucketURLBase, workspace, repoSlug)

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

func fetchFiles(url string, accessToken string) (*Response1, error) {
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

	var filesResp Response1
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

	var wg sync.WaitGroup
	wg.Add(len(filesResp.Children.Values))

	totalSize := 0
	resultCh := make(chan int)

	for _, file := range filesResp.Children.Values {
		go func(fileInfo File) {
			defer wg.Done()
			if fileInfo.Type == "FILE" {
				resultCh <- fileInfo.Size
			} else if fileInfo.Type == "DIRECTORY" {
				dirSize, err := fetchDirectorySize(projectKey, repoSlug, branchName, fileInfo.Path.Components, accessToken, bitbucketURLBase, apiver)
				if err != nil {
					fmt.Println("Error:", err)
					return
				}
				resultCh <- dirSize
			}
		}(file)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	for size := range resultCh {
		totalSize += size
	}

	return totalSize, nil
}

func fetchDirectorySize(projectKey string, repoSlug string, branchName string, components []string, accessToken string, bitbucketURLBase string, apiver string) (int, error) {
	url := fmt.Sprintf("%srest/api/%s/projects/%s/repos/%s/browse/%s?at=refs/heads/%s", bitbucketURLBase, apiver, projectKey, repoSlug, strings.Join(components, "/"), branchName)

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

	var wg sync.WaitGroup
	wg.Add(len(filesResp.Children.Values))

	totalSize := 0
	resultCh := make(chan int)

	for _, file := range filesResp.Children.Values {
		go func(fileInfo File) {
			defer wg.Done()
			if fileInfo.Type == "FILE" {
				resultCh <- fileInfo.Size
			} else if fileInfo.Type == "DIRECTORY" {
				subdirSize, err := fetchDirectorySize(projectKey, repoSlug, branchName, append(components, fileInfo.Path.Components...), accessToken, bitbucketURLBase, apiver)
				if err != nil {
					fmt.Println("Error:", err)
					return
				}
				resultCh <- subdirSize
			}
		}(file)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	for size := range resultCh {
		totalSize += size
	}

	return totalSize, nil
}
