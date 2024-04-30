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
type AnalysisResult struct {
	NumProjects     int
	NumRepositories int
	ProjectBranches []ProjectBranch
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
	Name             string   `json:"name"`
	DefaultMergeType string   `json:"default_merge_strategy"`
	MergeStrategies  []string `json:"merge_strategies"`
	Links            struct {
		Self    Link `json:"self"`
		Commits Link `json:"commits"`
		HTML    Link `json:"html"`
	} `json:"links"`
}

type Link struct {
	Href string `json:"href"`
}

type BranchResponse struct {
	Values  []Branch `json:"values"`
	Pagelen int      `json:"pagelen"`
	Size    int      `json:"size"`
	Page    int      `json:"page"`
	Next    string   `json:"next"`
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
type FileInfo1 struct {
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

type FileInfo struct {
	Path     string `json:"path"`
	Commit   Commit `json:"commit"`
	Type     string `json:"type"`
	Size     int    `json:"size,omitempty"`
	MimeType string `json:"mimetype,omitempty"`
	Links    struct {
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
	var message4 string
	emptyRepo := 0
	result := AnalysisResult{}

	spin1 := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	spin1.Prefix = "Get Projects... "
	spin1.Color("green", "bold")

	spin.Start()
	messageF := fmt.Sprintf("✅ The number of project(s) to analyze is %d\n", len(projects))
	spin.FinalMSG = messageF
	spin.Stop()

	for _, project := range projects {

		fmt.Printf("\n\t🟢  Analyse Projet: %s \n", project.Name)
		largestRepoSize = 0
		largestRepoBranch = ""

		urlrepos := fmt.Sprintf("%s%s/repositories/%s?q=project.key=\"%s\"&pagelen=100", url, apiver, workspace, project.Key)

		// Get Repos for each Project

		repos, err := CloudAllRepos(urlrepos, accessToken, exclusionList)
		if err != nil {
			fmt.Println("\r❌ Get Repos for each Project:", err)
			spin.Stop()
			continue
		}
		spin.Stop()

		nbRepos += len(repos)
		if nbRepos > 1 {
			message4 = "Repository"
		} else {
			message4 = "Repositories"
		}

		fmt.Printf("\t  ✅ The number of %s found is: %d\n", message4, len(repos))

		for _, repo := range repos {
			largestRepoSize = 0
			largestRepoBranch = ""

			isEmpty, err := isRepositoryEmpty(workspace, repo.Slug, accessToken, bitbucketURLBase)
			if err != nil {
				fmt.Printf("❌ Error when Testing if repo is empty %s: %v\n", repo.Name, err)
				spin.Stop()
				continue
			}

			if !isEmpty {

				urlrepos := fmt.Sprintf("%s%s/repositories/%s/%s/refs/branches/?pagelen=100", url, apiver, workspace, repo.Slug)

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
					messageB := fmt.Sprintf("\t   Analysis branch <%s> size...", branch.Name)
					spin1.Prefix = messageB
					spin1.Start()

					size, err := fetchBranchSize(workspace, repo.Slug, branch.Name, accessToken, url, apiver)
					messageF = ""
					spin1.FinalMSG = messageF

					spin1.Stop()
					if err != nil {
						fmt.Println("❌ Error retrieving branch size:", err)
						spin.Stop()
						os.Exit(1)
					}

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

	}
	result.NumProjects = len(projects)
	result.NumRepositories = nbRepos
	result.ProjectBranches = importantBranches

	// Save Result of Analysis
	file, err := os.Create("Results/config/analysis_repos.json")
	if err != nil {
		fmt.Println("❌ Error creating Analysis file:", err)
		return importantBranches, nbRepos, emptyRepo
	}
	defer file.Close()
	encoder := json.NewEncoder(file)

	err = encoder.Encode(result)
	if err != nil {
		fmt.Println("Error encoding JSON file <Results/config/analysis_repos.json> :", err)
		return importantBranches, nbRepos, emptyRepo
	}
	return importantBranches, nbRepos, emptyRepo
}

func GetRepos(project string, repos []Repo, url, baseapi, apiver, accessToken, bitbucketURLBase string, exclusionList *ExclusionList, spin *spinner.Spinner) ([]ProjectBranch, int, int) {

	var largestRepoSize int
	var largestRepoBranch string
	var importantBranches []ProjectBranch
	emptyRepo := 0
	nbRepos := 1
	result := AnalysisResult{}

	fmt.Printf("\n🟢 Analyse Projet: %s \n", project)

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
		fmt.Printf("\n\t   ✅ Repo: <%s> - Number of branches: %d\n", repos[0].Name, len(branches))

		// Finding the branch with the largest size

		for _, branch := range branches {
			messageB := fmt.Sprintf("\t   Analysis branch <%s> size...", branch.Name)
			spin.Prefix = messageB
			spin.Start()

			size, err := fetchBranchSize(project, repos[0].Slug, branch.Name, accessToken, url, apiver)
			if err != nil {
				fmt.Println("❌ Error retrieving branch size:", err)
				spin.Stop()
				continue
			}
			messageF := ""
			spin.FinalMSG = messageF

			spin.Stop()

			if size > largestRepoSize {
				largestRepoSize = size
				//largestRepoProject = project.Name
				largestRepoBranch = branch.Name
			}

		}
		fmt.Printf("\t     ✅ The largest branch of the repo is <%s> of size : %s\n", largestRepoBranch, formatSize(int64(largestRepoSize)))

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

	result.NumProjects = 1
	result.NumRepositories = nbRepos
	result.ProjectBranches = importantBranches

	// Save Result of Analysis
	file, err := os.Create("Results/config/analysis_repos.json")
	if err != nil {
		fmt.Println("❌ Error creating Analysis file:", err)
		return importantBranches, nbRepos, emptyRepo
	}
	defer file.Close()
	encoder := json.NewEncoder(file)

	err = encoder.Encode(result)
	if err != nil {
		fmt.Println("Error encoding JSON file <Results/config/analysis_repos.json> :", err)
		return importantBranches, nbRepos, emptyRepo
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
	bitbucketURLBase := fmt.Sprintf("%s%s/", url, apiver)
	bitbucketURL := fmt.Sprintf("%s%s/workspaces/%s/projects/?pagelen=100", url, apiver, workspace)

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

	} else if len(project) > 0 && len(repo) == 0 {

		spin.Start()
		bitbucketURLProject := fmt.Sprintf("%s%s/workspaces/%s/projects/%s", url, apiver, workspace, project)

		projects, err := CloudOnelProjects(bitbucketURLProject, accessToken, exclusionList)
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
			importantBranches, nbRepos, emptyRepo = GetReposProjectCloud(projects, url, baseapi, apiver, accessToken, bitbucketURLBase, workspace, nbRepos, exclusionList, spin)

		}
	} else if len(project) > 0 && len(repo) > 0 {

		spin.Start()
		bitbucketURLProject := fmt.Sprintf("%s/%s/repos/%s", bitbucketURL, project, repo)
		Repos, err := fetchOneRepos(bitbucketURLProject, accessToken, exclusionList)
		if err != nil {
			fmt.Printf("\n❌ Error Get Repo:%s/%s - %v", project, repo, err)
			spin.Stop()
			return nil, err
		}
		fmt.Printf("Taille : %d", len(Repos))
		spin.Stop()

		//	importantBranches, nbRepos, emptyRepo = GetRepos(project, Repos, url, baseapi, apiver, accessToken, bitbucketURLBase, exclusionList, spin)

	} else {
		spin.Stop()
		fmt.Println("❌ Error Project name is empty")
		os.Exit(1)
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
	//os.Exit(1)
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

func CloudOnelProjects(url string, accessToken string, exclusionList *ExclusionList) ([]Projectc, error) {
	var allProjects []Projectc

	projectsResp, err := CloudProjects(url, accessToken, false)
	if err != nil {
		return nil, err
	}
	project := projectsResp.(*Projectc)

	if len(*&project.Key) == 0 {
		fmt.Println("\n❌ Error Project not exist")
		os.Exit(1)
	}

	if len(exclusionList.Projectcs) == 0 && len(exclusionList.Repos) == 0 {
		allProjects = append(allProjects, *project)
	} else {
		if !isProjectExcluded(exclusionList, project.Key) {
			allProjects = append(allProjects, *project)
		}
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

func fetchOneRepos(url string, accessToken string, exclusionList *ExclusionList) ([]Reposc, error) {
	var allRepos []Reposc

	reposResp, err := CloudRepos(url, accessToken, false)
	if err != nil {
		return nil, err
	}
	repo := reposResp.(*Reposc)

	if len(*&repo.Name) == 0 {
		fmt.Println("\n❌ Error Repo or Project not exist")
		os.Exit(1)
	}

	KEYTEST := repo.Project.Key + "/" + repo.Slug

	if len(exclusionList.Projectcs) == 0 && len(exclusionList.Repos) == 0 {
		allRepos = append(allRepos, *repo)
	} else {
		if !isRepoExcluded(exclusionList, KEYTEST) {
			allRepos = append(allRepos, *repo)
		}
	}

	return allRepos, nil
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
	for url != "" {
		branchesResp, err := CloudBranches(url, accessToken)
		if err != nil {
			return nil, err
		}
		allBranches = append(allBranches, branchesResp.Values...)

		url = branchesResp.Next
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

	urlMain := fmt.Sprintf("%srepositories/%s/%s/src/main/?pagelen=100", bitbucketURLBase, workspace, repoSlug)

	// Récupérer les fichiers de la branche principale
	filesResp, err := fetchFiles(urlMain, accessToken)
	if err != nil {
		return false, fmt.Errorf("❌ Error when testing if repo: %s is empty - Function: %s - %v", repoSlug, "getbibucket-isRepositoryEmpty", err)
	}

	// Si la réponse est nulle, essayer avec la branche "master"
	if filesResp == nil {
		urlMaster := fmt.Sprintf("%srepositories/%s/%s/src/master/?pagelen=100", bitbucketURLBase, workspace, repoSlug)
		filesResp, err = fetchFiles(urlMaster, accessToken)
		if err != nil {
			return false, fmt.Errorf("❌ Error when testing if repo: %s is empty - Function: %s - %v", repoSlug, "getbibucket-isRepositoryEmpty", err)
		}
	}

	// Vérifier si la liste de fichiers est vide
	if len(filesResp.Values) == 0 {
		return true, nil
	}

	return false, nil
}

func fetchAllFiles(url string, accessToken string) ([]FileInfo, error) {
	var allFiles []FileInfo

	for url != "" {
		filesResp, err := fetchFiles(url, accessToken)
		if err != nil {
			return nil, err
		}

		allFiles = append(allFiles, filesResp.Values...)
		//if filesResp.IsLastPage {
		//	break
		//}
		//url = fmt.Sprintf("%s?start=%d", url, filesResp.Next)
		url = filesResp.Next
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

	if strings.Contains(string(body), "error") || strings.Contains(string(body), "Commit not found") {
		// Branch does not exist, return nil response
		return nil, nil
	}

	var filesResp Response1
	err = json.Unmarshal(body, &filesResp)
	if err != nil {
		return nil, err
	}

	return &filesResp, nil
}
func fetchBranchSize(workspace, repoSlug, branchName, accessToken, url, apiver string) (int, error) {

	url1 := fmt.Sprintf("%s%s/repositories/%s/%s/src/%s/?pagelen=100", url, apiver, workspace, repoSlug, branchName)

	req, err := http.NewRequest("GET", url1, nil)
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

	var filesResp Response1
	err = json.Unmarshal(body, &filesResp)
	if err != nil {
		return 0, err
	}

	var wg sync.WaitGroup
	wg.Add(len(filesResp.Values))

	totalSize := 0
	resultCh := make(chan int)

	for _, file := range filesResp.Values {
		go func(fileInfo FileInfo) {
			defer wg.Done()
			if fileInfo.Type == "commit_file" {
				resultCh <- fileInfo.Size
			} else if fileInfo.Type == "commit_directory" {
				dirSize, err := fetchDirectorySize(workspace, repoSlug, branchName, fileInfo.Path, accessToken, url, apiver)
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

func fetchDirectorySize(workspace string, repoSlug string, branchName string, components string, accessToken string, url string, apiver string) (int, error) {
	//url := fmt.Sprintf("%srest/api/%s/projects/%s/repos/%s/browse/%s?at=refs/heads/%s", bitbucketURLBase, apiver, projectKey, repoSlug, strings.Join(components, "/"), branchName)

	url1 := fmt.Sprintf("%s%s/reposiories/%s/%s/src/%s/%s/?pagelen=100", url, apiver, workspace, repoSlug, branchName, components)

	req, err := http.NewRequest("GET", url1, nil)
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

	var filesResp Response1
	err = json.Unmarshal(body, &filesResp)
	if err != nil {
		return 0, err
	}

	var wg sync.WaitGroup
	wg.Add(len(filesResp.Values))

	totalSize := 0
	resultCh := make(chan int)

	for _, file := range filesResp.Values {
		go func(fileInfo FileInfo) {
			defer wg.Done()
			if fileInfo.Type == "commit_file" {
				resultCh <- fileInfo.Size
			} else if fileInfo.Type == "commit_directory" {
				subdirSize, err := fetchDirectorySize(workspace, repoSlug, branchName, fileInfo.Path, accessToken, url, apiver)
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
