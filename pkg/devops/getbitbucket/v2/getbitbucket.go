package getbibucketv2

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/ktrysmt/go-bitbucket"
)

type ProjectBranch struct {
	Org         string
	ProjectKey  string
	RepoSlug    string
	MainBranch  string
	LargestSize int
}

type SummaryStats struct {
	LargestRepo       string
	LargestRepoBranch string
	NbRepos           int
	EmptyRepo         int
	TotalExclude      int
	TotalArchiv       int
	TotalBranches     int
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

type ProjectcsResponse struct {
	Values []Projectc `json:"values"`
	Next   string     `json:"next"`
}

type ExclusionList struct {
	Projectcs map[string]bool `json:"Projects"`
	Repos     map[string]bool `json:"repos"`
}

type ParamsProjectBitbucket struct {
	Client           *bitbucket.Client
	Projects         []Projectc
	Workspace        string
	URL              string
	BaseAPI          string
	APIVersion       string
	AccessToken      string
	BitbucketURLBase string
	Organization     string
	Exclusionlist    *ExclusionList
	Spin             *spinner.Spinner
	Period           int
	Stats            bool
	DefaultB         bool
}

type Response1 struct {
	Values  []FileInfo `json:"values"`
	Pagelen int        `json:"pagelen"`
	Page    int        `json:"page"`
	Next    string     `json:"next"`
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
type Commit struct {
	Hash  string `json:"hash"`
	Links struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
	Type string `json:"type"`
}

const PrefixMsg = "Get Projects..."

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

func isRepoExcluded(exclusionList *ExclusionList, repoKey string) bool {
	_, excluded := exclusionList.Repos[repoKey]
	return excluded
}
func isProjectExcluded(exclusionList *ExclusionList, projectKey string) bool {
	_, excluded := exclusionList.Projectcs[projectKey]
	return excluded
}

func GetProjectBitbucketListCloud(platformConfig map[string]interface{}, exclusionFile string) ([]ProjectBranch, error) {

	var totalExclude, totalArchiv, emptyRepo, TotalBranches int
	var nbRepos int

	var largestRepoBranch, largesRepo string
	var importantBranches []ProjectBranch
	var projects []Projectc
	var exclusionList *ExclusionList
	var err error
	var totalSize int
	//	result := AnalysisResult{}

	fmt.Print("\n🔎 Analysis of devops platform objects ...\n")

	spin := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	spin.Prefix = "Processing"
	spin.Color("green", "bold")

	exclusionList, err = loadExclusionFileOrCreateNew(exclusionFile)
	if err != nil {
		fmt.Printf("\n❌ Error Read Exclusion File <%s>: %v", exclusionFile, err)
		spin.Stop()
		return nil, err
	}

	client := bitbucket.NewOAuthbearerToken(platformConfig["AccessToken"].(string))

	project := platformConfig["Project"].(string)
	repos := platformConfig["Repos"].(string)
	bitbucketURLBase := fmt.Sprintf("%s%s/", platformConfig["Url"].(string), platformConfig["Apiver"].(string))

	if len(project) == 0 && len(repos) == 0 {
		projects, err = getAllProjects(client, platformConfig["Workspace"].(string), exclusionList)
		if err != nil {
			fmt.Println("\r❌ Error Get All Projects:", err)
			spin.Stop()
			return nil, err
		}

		//importantBranches, nbRepos, err = getAllProjects(client, platformConfig["Workspace"].(string), exclusionList)
	} /*else if len(project) > 0 && len(repos) == 0 {
		importantBranches, nbRepos, err = getProject(client, exclusionList, project)
	} else if len(project) > 0 && len(repos) > 0 {
		importantBranches, nbRepos, err = getProjectRepos(client, exclusionList, project, repos)
	} else {
		spin.Stop()
		fmt.Println("❌ Error Project name is empty")
		return nil, fmt.Errorf("project name is empty")
	}

	if err != nil {
		spin.Stop()
		return nil, err
	}*/

	spin.Stop()
	params := getCommonParams(client, platformConfig, projects, exclusionList, spin, bitbucketURLBase)
	importantBranches, emptyRepo, nbRepos, TotalBranches, totalExclude, totalArchiv = getRepoAnalyse(params)

	largestRepoBranch, largesRepo = findLargestRepository(importantBranches, &totalSize)

	/*result.NumProjects = 1
	result.NumRepositories = nbRepos
	result.ProjectBranches = importantBranches

	if err := saveAnalysisResult("Results/config/analysis_repos_bitbucketdc.json", result); err != nil {
		return importantBranches, nil
	}*/

	stats := SummaryStats{
		LargestRepo:       largesRepo,
		LargestRepoBranch: largestRepoBranch,
		NbRepos:           nbRepos,
		EmptyRepo:         emptyRepo,
		TotalExclude:      totalExclude,
		TotalArchiv:       totalArchiv,
		TotalBranches:     TotalBranches,
	}

	printSummary(params.Organization, stats)
	os.Exit(1)
	return importantBranches, nil
}

func findLargestRepository(importantBranches []ProjectBranch, totalSize *int) (string, string) {
	var largestRepoSize int
	var largestRepoBranch, largesRepo string

	for _, branch := range importantBranches {
		if branch.LargestSize > largestRepoSize {
			largestRepoSize = branch.LargestSize
			largestRepoBranch = branch.MainBranch
			largesRepo = branch.RepoSlug
		}
		*totalSize += branch.LargestSize
	}
	//return largestRepoSize, largestRepoBranch, largesRepo
	return largestRepoBranch, largesRepo
}

func printSummary(Org string, stats SummaryStats) {
	fmt.Printf("\n✅ The largest Repository is <%s> in the organization <%s> with the branch <%s> \n", stats.LargestRepo, Org, stats.LargestRepoBranch)
	fmt.Printf("\r✅ Total Repositories that will be analyzed: %d - Find empty : %d - Excluded : %d - Archived : %d\n", stats.NbRepos-stats.EmptyRepo-stats.TotalExclude-stats.TotalArchiv, stats.EmptyRepo, stats.TotalExclude, stats.TotalArchiv)
	fmt.Printf("\r✅ Total Branches that will be analyzed: %d\n", stats.TotalBranches)
}

func loadExclusionFileOrCreateNew(exclusionFile string) (*ExclusionList, error) {
	if exclusionFile == "0" {
		return &ExclusionList{
			Projectcs: make(map[string]bool),
			Repos:     make(map[string]bool),
		}, nil
	}
	return loadExclusionList(exclusionFile)
}

func getCommonParams(client *bitbucket.Client, platformConfig map[string]interface{}, project []Projectc, exclusionList *ExclusionList, spin *spinner.Spinner, bitbucketURLBase string) ParamsProjectBitbucket {
	return ParamsProjectBitbucket{
		Client:           client,
		Projects:         project,
		Workspace:        platformConfig["Workspace"].(string),
		URL:              platformConfig["Url"].(string),
		BaseAPI:          platformConfig["Baseapi"].(string),
		APIVersion:       platformConfig["Apiver"].(string),
		AccessToken:      platformConfig["AccessToken"].(string),
		BitbucketURLBase: bitbucketURLBase,
		Organization:     platformConfig["Organization"].(string),
		Exclusionlist:    exclusionList,
		Spin:             spin,
		Period:           int(platformConfig["Period"].(float64)),
		Stats:            platformConfig["Stats"].(bool),
		DefaultB:         platformConfig["DefaultBranch"].(bool),
	}
}

func getAllProjects(client *bitbucket.Client, workspace string, exclusionList *ExclusionList) ([]Projectc, error) {

	var projects []Projectc

	projectsRes, err := client.Workspaces.Projects(workspace)
	if err != nil {
		return nil, err
	}

	for _, project := range projectsRes.Items {
		if isProjectExcluded(exclusionList, project.Key) {
			continue
		}

		projects = append(projects, Projectc{
			Key:         project.Key,
			UUID:        project.Uuid,
			IsPrivate:   project.Is_private,
			Name:        project.Name,
			Description: project.Description,
		})
	}

	return projects, nil
}

func getRepoAnalyse(params ParamsProjectBitbucket) ([]ProjectBranch, int, int, int, int, int) {

	var emptyRepos = 0
	var totalexclude = 0
	var importantBranches []ProjectBranch
	var NBRrepo, TotalBranches int
	NBRrepos := 0
	cptarchiv := 0

	cpt := 1

	message4 := "Repo(s)"

	spin1 := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	spin1.Prefix = PrefixMsg
	spin1.Color("green", "bold")

	params.Spin.Start()
	messageF := fmt.Sprintf("✅ The number of project(s) to analyze is %d\n", len(params.Projects))
	params.Spin.FinalMSG = messageF
	params.Spin.Stop()

	// Get Repository in each Project
	for _, project := range params.Projects {

		fmt.Printf("\n\t🟢  Analyse Projet: %s \n", project.Name)
		//largestRepoSize = 0
		//largestRepoBranch = ""

		emptyOrArchivedCount, excludedCount, repos, err := listReposForProject(params, project.Key)
		if err != nil {
			fmt.Println("\r❌ Get Repos for each Project:", err)
			spin1.Stop()
			continue
		}
		emptyRepos = emptyRepos + emptyOrArchivedCount
		totalexclude = totalexclude + excludedCount

		spin1.Stop()

		for _, repo := range repos {
			fmt.Println("Reponame:", repo.Name)
			largestRepoBranch, repoBranches, nbrb, err := analyzeRepoBranches(params, repo, cpt, spin1)
			if err != nil {
				largestRepoBranch = repo.Mainbranch.Name

			}

			importantBranches = append(importantBranches, ProjectBranch{
				Org:         params.Organization,
				RepoSlug:    repo.Slug,
				MainBranch:  largestRepoBranch,
				LargestSize: len(repoBranches),
			})
			TotalBranches += nbrb

			cpt++
		}
		if emptyOrArchivedCount > 0 {
			NBRrepo = len(repos) + emptyOrArchivedCount
			fmt.Printf("\t  ✅ The number of %s found is: %d - Find empty %d:\n", message4, NBRrepo, emptyOrArchivedCount)
		} else {
			NBRrepo = len(repos)
			fmt.Printf("\t  ✅ The number of %s found is: %d\n", message4, NBRrepo)
		}

		NBRrepos += NBRrepo

	}

	return importantBranches, emptyRepos, NBRrepos, TotalBranches, totalexclude, cptarchiv

}

// List Project in Workspace
func listReposForProject(parms ParamsProjectBitbucket, projectKey string) (int, int, []*bitbucket.Repository, error) {
	var allRepos []*bitbucket.Repository
	var excludedCount, emptyOrArchivedCount int

	page := 1
	for {
		reposRes, err := parms.Client.Repositories.ListProject(&bitbucket.RepositoriesOptions{
			Owner:   parms.Workspace,
			Project: projectKey,
			Page:    &page,
		})
		if err != nil {
			return 0, 0, nil, err
		}

		emptyOrArchivedCount, excludedCount, allRepos = listRepos(parms, reposRes)

		if len(reposRes.Items) < int(reposRes.Pagelen) {
			break
		}

		page++
	}

	return emptyOrArchivedCount, excludedCount, allRepos, nil
}

// List Repository in each Project
func listRepos(parms ParamsProjectBitbucket, reposRes *bitbucket.RepositoriesRes) (int, int, []*bitbucket.Repository) {
	var allRepos []*bitbucket.Repository
	var excludedCount, emptyOrArchivedCount int

	for _, repo := range reposRes.Items {
		if isRepoExcluded(parms.Exclusionlist, repo.Slug) {
			excludedCount++
			continue
		}

		isEmpty, err := isRepositoryEmpty(parms.Workspace, repo.Slug, repo.Mainbranch.Name, parms.AccessToken, parms.BitbucketURLBase)
		if err != nil {
			fmt.Printf("❌ Error when Testing if repo is empty %s: %v\n", repo.Slug, err)

		}
		if isEmpty {
			emptyOrArchivedCount++
			continue
		}

		allRepos = append(allRepos, &repo)

	}
	return emptyOrArchivedCount, excludedCount, allRepos
}

// Test is Repository is empty
func isRepositoryEmpty(workspace, repoSlug, mainbranch, accessToken, bitbucketURLBase string) (bool, error) {

	urlMain := fmt.Sprintf("%srepositories/%s/%s/src/%s/?pagelen=100", bitbucketURLBase, workspace, repoSlug, mainbranch)

	// Récupérer les fichiers de la branche principale
	filesResp, err := fetchFiles(urlMain, accessToken)
	if err != nil {
		return false, fmt.Errorf("❌ Error when testing if repo: %s is empty - Function: %s - %v", repoSlug, "getbibucket-isRepositoryEmpty", err)
	}

	if filesResp == nil {
		urlMaster := fmt.Sprintf("%srepositories/%s/%s/src/master/?pagelen=100", bitbucketURLBase, workspace, repoSlug)
		filesResp, err = fetchFiles(urlMaster, accessToken)
		if err != nil {
			return false, fmt.Errorf("❌ Error when testing if repo: %s is empty - Function: %s - %v", repoSlug, "getbibucket-isRepositoryEmpty", err)
		}
	}

	if len(filesResp.Values) == 0 {
		return true, nil
	}

	return false, nil
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

func analyzeRepoBranches(parms ParamsProjectBitbucket, repo *bitbucket.Repository, cpt int, spin1 *spinner.Spinner) (string, []*bitbucket.RepositoryBranch, int, error) {
	var branches []*bitbucket.RepositoryBranch
	//var branchPushes map[string]int

	spin1.Prefix = "Analyzing branches"
	spin1.Start()

	// Get all branches for the repository
	repoBranches, err := getAllBranches(parms.Client, parms.Workspace, repo.Slug)
	if err != nil {
		spin1.Stop()
		return "", nil, 0, err
	}

	// Determine the largest branch based on the number of commits
	largestRepoBranch := determineLargestBranch(parms, repo, branches)
	if err != nil {
		spin1.Stop()
		//fmt.Println("error determineLargestBranch")
		return "", nil, 1, err
	}

	spin1.Stop()

	// Print analysis summary
	fmt.Printf("\t✅ Repo %d: %s - Number of branches: %d - Largest Branch: %s\n", cpt, repo.Slug, len(repoBranches), largestRepoBranch)

	return largestRepoBranch, branches, len(repoBranches), nil
}

func getAllBranches(client *bitbucket.Client, workspace, repoSlug string) ([]*bitbucket.RepositoryBranch, error) {
	// Get all branches for the repository
	branchesRes, err := client.Repositories.Repository.ListBranches(&bitbucket.RepositoryBranchOptions{
		Owner:    workspace,
		RepoSlug: repoSlug,
	})
	if err != nil {
		return nil, err
	}
	// Convert branchesRes.Values to []*bitbucket.RepositoryBranch
	branches := make([]*bitbucket.RepositoryBranch, len(branchesRes.Branches))
	for i, branch := range branchesRes.Branches {
		branches[i] = &branch
	}

	return branches, nil
}

func getBranchCommitCount(client *bitbucket.Client, workspace, repoSlug, branchName string) (int, error) {

	commitsRes, err := client.Repositories.Commits.GetCommits(&bitbucket.CommitsOptions{
		Owner:    workspace,
		RepoSlug: repoSlug,
		Include:  branchName,
	})

	if err != nil {
		return 0, err
	}
	commitsMap, ok := commitsRes.(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("❌ unexpected response format")
	}

	values, ok := commitsMap["values"].([]interface{})
	if !ok {
		return 0, fmt.Errorf("❌ unexpected response format")
	}

	return len(values), nil
}

func countBranchPushes(client *bitbucket.Client, workspace, repoSlug string, period int) (map[string]int, error) {
	// Implement counting of branch pushes based on events
	// (This functionality depends on whether Bitbucket provides an API for retrieving push events for branches)
	// For now, return an empty map.
	return make(map[string]int), nil
}

func determineLargestBranch(parms ParamsProjectBitbucket, repo *bitbucket.Repository, branches []*bitbucket.RepositoryBranch) string {
	var largestRepoBranch string
	var maxCommits int

	for _, branch := range branches {
		commits, err := getBranchCommitCount(parms.Client, parms.Workspace, repo.Slug, branch.Name)
		if err != nil {
			fmt.Printf("❌ Error when retrieving commit count for branch %s in repo %s: %v\n", branch.Name, repo.Slug, err)
			continue
		}
		if commits > maxCommits {
			maxCommits = commits
			largestRepoBranch = branch.Name
		}
	}

	if largestRepoBranch == "" {
		largestRepoBranch = repo.Mainbranch.Name
	}

	return largestRepoBranch
}

/*func getDefaultBranch(client *bitbucket.Client, workspace, repoSlug string) (string, error) {
	// Get the repository details to fetch the default branch
	repoDetails, err := client.Repositories.Get(workspace, repoSlug)
	if err != nil {
		return "", err
	}

	// Return the name of the default branch
	return repoDetails.Mainbranch.Name, nil
}*/

func sortBranchesByCommits(branchPushes map[string]int) []string {
	// Create a slice to hold branch names
	var branches []string

	// Iterate over the branchPushes map and append branch names to the slice
	for branch := range branchPushes {
		branches = append(branches, branch)
	}

	// Sort the branches by the number of commits in descending order
	sort.Slice(branches, func(i, j int) bool {
		return branchPushes[branches[i]] > branchPushes[branches[j]]
	})

	return branches
}

func saveAnalysisResult(filepath string, result AnalysisResult) error {
	file, err := os.Create(filepath)
	if err != nil {
		fmt.Println("❌ Error creating Analysis file:", err)
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(result); err != nil {
		fmt.Println("Error encoding JSON file <", filepath, ">:", err)
		return err
	}

	return nil
}
