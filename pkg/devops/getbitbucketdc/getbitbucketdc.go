package getbibucketdc

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

type ExclusionList struct {
	Projects map[string]bool `json:"projects"`
	Repos    map[string]bool `json:"repos"`
}

func loadExclusionList(filename string) (*ExclusionList, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	exclusionList := &ExclusionList{
		Projects: make(map[string]bool),
		Repos:    make(map[string]bool),
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "/")
		if len(parts) == 1 {
			// Get Projet
			exclusionList.Projects[parts[0]] = true
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

func GetReposProject(projects []Project, url, baseapi, apiver, accessToken, bitbucketURLBase string, nbRepos int, exclusionList *ExclusionList, spin *spinner.Spinner) ([]ProjectBranch, int, int) {

	var largestRepoSize int
	var largestRepoBranch string
	var importantBranches []ProjectBranch
	emptyRepo := 0

	for _, project := range projects {
		// Display Project Name
		//	fmt.Printf("✅ Projet: %s - Key: %s\n", project.Name, project.Key)
		largestRepoSize = 0
		largestRepoBranch = ""

		urlrepos := fmt.Sprintf("%s%s%s/projects/%s/repos", url, baseapi, apiver, project.Key)

		// Get Repos for each Project

		repos, err := fetchAllRepos(urlrepos, accessToken, exclusionList)
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

func GetRepos(project string, repos []Repo, url, baseapi, apiver, accessToken, bitbucketURLBase string, exclusionList *ExclusionList, spin *spinner.Spinner) ([]ProjectBranch, int, int) {

	var largestRepoSize int
	var largestRepoBranch string
	var importantBranches []ProjectBranch
	emptyRepo := 0
	nbRepos := 1

	spin.Prefix = "Analysis of repos Finds the largest branch size..."
	spin.Start()

	isEmpty, err := isRepositoryEmpty(project, repos[0].Slug, accessToken, bitbucketURLBase, apiver)
	if err != nil {
		fmt.Printf("❌ Error when Testing if repo is empty %s: %v\n", repos[0].Name, err)
		spin.Stop()
		os.Exit(1)
	}
	if !isEmpty {

		urlrepos := fmt.Sprintf("%s%s%s/projects/%s/repos/%s/branches", url, baseapi, apiver, project, repos[0].Slug)

		branches, err := fetchAllBranches(urlrepos, accessToken)
		if err != nil {
			fmt.Printf("❌ Error when retrieving branches for repo %s: %v\n", repos[0].Name, err)
			spin.Stop()
			os.Exit(1)
		}
		// Display Number of branches by repo
		// fmt.Printf("\r\t✅ Repo: %s - Number of branches: %d\n", repos[0].Name, len(branches))

		// Finding the branch with the largest size

		for _, branch := range branches {
			// Display Branch name
			// fmt.Printf("\t\t✅ Branche: %s\n", branch.Name)

			size, err := fetchBranchSize(project, repos[0].Slug, branch.Name, accessToken, url, apiver)
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
			ProjectKey:  project,
			RepoSlug:    repos[0].Slug,
			MainBranch:  largestRepoBranch,
			LargestSize: largestRepoSize,
		})
	} else {
		fmt.Println("❌ Repo is empty:", repos[0].Name)
		return importantBranches, nbRepos, emptyRepo
	}

	spin.Stop()
	return importantBranches, nbRepos, emptyRepo

}

func GetProjectBitbucketList(url, baseapi, apiver, accessToken, exlusionfile, project, repo string) ([]ProjectBranch, error) {

	var largestRepoSize int
	var totalSize int
	var largestRepoProject, largestRepoBranch string
	var importantBranches []ProjectBranch
	var exclusionList *ExclusionList
	var err1 error

	totalSize = 0
	nbRepos := 0
	emptyRepo := 0
	bitbucketURLBase := "http://ec2-18-194-139-24.eu-central-1.compute.amazonaws.com:7990/"
	bitbucketURL := fmt.Sprintf("%s%s%s/projects", url, baseapi, apiver)

	// Get All Projects

	spin := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	spin.Prefix = "Get Projects... "
	spin.Color("green", "bold")
	spin.Start()

	if exlusionfile == "0" {
		exclusionList = &ExclusionList{
			Projects: make(map[string]bool),
			Repos:    make(map[string]bool),
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

		projects, err := fetchAllProjects(bitbucketURL, accessToken, exclusionList)
		if err != nil {
			fmt.Println("\r❌ Error Get All Projects:", err)
			spin.Stop()
			return nil, err
		}
		spin.Stop()

		importantBranches, nbRepos, emptyRepo = GetReposProject(projects, url, baseapi, apiver, accessToken, bitbucketURLBase, nbRepos, exclusionList, spin)

	} else if len(project) > 0 && len(repo) == 0 {

		bitbucketURLProject := fmt.Sprintf("%s/%s", bitbucketURL, project)

		projects, err := fetchOnelProjects(bitbucketURLProject, accessToken, exclusionList)
		if err != nil {
			fmt.Printf("\n❌ Error Get Project:%s - %v", project, err)
			spin.Stop()
			return nil, err
		}
		spin.Stop()

		if len(projects) == 0 {
			fmt.Printf("\n❌ Error Project:%s not exist - %v", project, err)
			spin.Stop()
			return nil, err
		} else {
			importantBranches, nbRepos, emptyRepo = GetReposProject(projects, url, baseapi, apiver, accessToken, bitbucketURLBase, nbRepos, exclusionList, spin)

		}
	} else if len(project) > 0 && len(repo) > 0 {

		bitbucketURLProject := fmt.Sprintf("%s/%s/repos/%s", bitbucketURL, project, repo)
		Repos, err := fetchOneRepos(bitbucketURLProject, accessToken, exclusionList)
		if err != nil {
			fmt.Printf("\n❌ Error Get Repo:%s/%s - %v", project, repo, err)
			spin.Stop()
			return nil, err
		}
		fmt.Printf("Taille : %d", len(Repos))
		spin.Stop()

		importantBranches, nbRepos, emptyRepo = GetRepos(project, Repos, url, baseapi, apiver, accessToken, bitbucketURLBase, exclusionList, spin)

	} else {
		spin.Stop()
		fmt.Println("❌ Error Project name is empty")
		os.Exit(1)
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
	totalSizeMB := formatSize(int64(totalSize))
	largestRepoSizeMB := formatSize(int64(largestRepoSize))

	fmt.Printf("\n✅ The largest repo is in the %s with the branch %s and a size of %s\n", largestRepoProject, largestRepoBranch, largestRepoSizeMB)
	fmt.Printf("\r✅ Total size of your organization's repositories: %s\n", totalSizeMB)
	fmt.Printf("\r✅ Total repositories analyzed: %d - Find empty : %d\n", nbRepos-emptyRepo, emptyRepo)

	return importantBranches, nil
}

func fetchAllProjects(url string, accessToken string, exclusionList *ExclusionList) ([]Project, error) {
	var allProjects []Project
	for {
		projectsResp, err := fetchProjects(url, accessToken, true)
		if err != nil {
			return nil, err
		}
		projectResponse := projectsResp.(*ProjectResponse)
		//allProjects = append(allProjects, projectsResp.Values...)
		for _, project := range projectResponse.Values {

			if len(exclusionList.Projects) == 0 && len(exclusionList.Repos) == 0 {
				allProjects = append(allProjects, project)
			} else {
				if !isProjectExcluded(exclusionList, project.Key) {
					allProjects = append(allProjects, project)
				}
			}
		}

		if projectResponse.IsLastPage {
			break
		}
		url = fmt.Sprintf("%s?start=%d", url, projectResponse.NextPageStart)
	}
	return allProjects, nil
}

func fetchOnelProjects(url string, accessToken string, exclusionList *ExclusionList) ([]Project, error) {
	var allProjects []Project

	projectsResp, err := fetchProjects(url, accessToken, false)
	if err != nil {
		return nil, err
	}
	project := projectsResp.(*Project)

	if len(*&project.Key) == 0 {
		fmt.Println("\n❌ Error Project not exist")
		os.Exit(1)
	}

	if len(exclusionList.Projects) == 0 && len(exclusionList.Repos) == 0 {
		allProjects = append(allProjects, *project)
	} else {
		if !isProjectExcluded(exclusionList, project.Key) {
			allProjects = append(allProjects, *project)
		}
	}

	return allProjects, nil
}

func fetchOneRepos(url string, accessToken string, exclusionList *ExclusionList) ([]Repo, error) {
	var allRepos []Repo

	reposResp, err := fetchRepos(url, accessToken, false)
	if err != nil {
		return nil, err
	}
	repo := reposResp.(*Repo)

	if len(*&repo.Name) == 0 {
		fmt.Println("\n❌ Error Repo or Project not exist")
		os.Exit(1)
	}

	KEYTEST := repo.Project.Key + "/" + repo.Slug

	if len(exclusionList.Projects) == 0 && len(exclusionList.Repos) == 0 {
		allRepos = append(allRepos, *repo)
	} else {
		if !isRepoExcluded(exclusionList, KEYTEST) {
			allRepos = append(allRepos, *repo)
		}
	}

	return allRepos, nil
}

func fetchProjects(url string, accessToken string, isProjectResponse bool) (interface{}, error) {
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
		projectsResp = &ProjectResponse{}
	} else {
		projectsResp = &Project{}
	}

	err = json.Unmarshal(body, &projectsResp)
	if err != nil {
		return nil, err
	}

	return projectsResp, nil

}

func isProjectExcluded(exclusionList *ExclusionList, project string) bool {
	_, excluded := exclusionList.Projects[project]
	return excluded
}

func isRepoExcluded(exclusionList *ExclusionList, repo string) bool {
	_, excluded := exclusionList.Repos[repo]
	return excluded
}

func fetchAllRepos(url string, accessToken string, exclusionList *ExclusionList) ([]Repo, error) {
	var allRepos []Repo
	for {
		reposResp, err := fetchRepos(url, accessToken, true)
		if err != nil {
			return nil, err
		}
		ReposResponse := reposResp.(*RepoResponse)
		for _, repo := range ReposResponse.Values {
			KEYTEST := repo.Project.Key + "/" + repo.Slug

			if len(exclusionList.Projects) == 0 && len(exclusionList.Repos) == 0 {
				allRepos = append(allRepos, repo)
			} else {
				if !isRepoExcluded(exclusionList, KEYTEST) {
					allRepos = append(allRepos, repo)
				}
			}

		}

		if ReposResponse.IsLastPage {
			break
		}
		url = fmt.Sprintf("%s?start=%d", url, ReposResponse.NextPageStart)
	}
	return allRepos, nil
}

// func fetchRepos(url string, accessToken string,isProjectResponse bool) (*RepoResponse, error) {
func fetchRepos(url string, accessToken string, isProjectResponse bool) (interface{}, error) {
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
		reposResp = &RepoResponse{}
	} else {
		reposResp = &Repo{}
	}

	err = json.Unmarshal(body, &reposResp)
	if err != nil {
		return nil, err
	}

	return reposResp, nil

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
