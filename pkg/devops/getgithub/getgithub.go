package getgithub

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/colussim/GoLC/pkg/utils"
	"github.com/google/go-github/v62/github"
)

type ExclusionList struct {
	Repos map[string]bool `json:"repos"`
}

type ParamsReposGithub struct {
	Repos         []Repository
	URL           string
	BaseAPI       string
	Apiver        string
	AccessToken   string
	Organization  string
	NBRepos       int
	ExclusionList *ExclusionList
	Spin          *spinner.Spinner
	Branch        string
	Branchescount int
}
type Repository struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Path          string `json:"full_name"`
	SizeR         int64  `json:"size"`
	Language      string `json:"language"`
	DefaultBranch string `json:"default_branch"`
	Archived      bool   `json:"archived"`
	LOC           map[string]int
}

type ProjectBranch struct {
	Org         string
	RepoSlug    string
	MainBranch  string
	LargestSize int
}

type AnalysisResult struct {
	NumRepositories int
	ProjectBranches []ProjectBranch
}

type TreeItem struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
	Type string `json:"type"`
	Sha  string `json:"sha"`
	Size int    `json:"size,omitempty"`
}

type TreeResponse struct {
	Sha       string     `json:"sha"`
	Url       string     `json:"url"`
	Tree      []TreeItem `json:"tree"`
	Truncated bool       `json:"truncated"`
}

type Branch struct {
	Name      string     `json:"name"`
	Commit    CommitInfo `json:"commit"`
	Protected bool       `json:"protected"`
}

type CommitInfo struct {
	Sha string `json:"sha"`
	URL string `json:"url"`
}

type BranchInfo struct {
	Name     string
	Activity int
}

type Lastanalyse struct {
	LastRepos  string
	LastBranch string
}

// const apigit = "X-GitHub-Api-Version"
const PrefixMsg = "Get Repo(s)..."

func loadExclusionList(filename string) (*ExclusionList, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	exclusionList := &ExclusionList{
		Repos: make(map[string]bool),
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		exclusionList.Repos[line] = true

	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return exclusionList, nil
}

// Function to randomly select k branch elements of an array
func selectRandomBranches(branches []Branch, k int) []Branch {
	source := rand.NewSource(time.Now().UnixNano())
	randGen := rand.New(source)

	if k >= len(branches) {
		return branches
	}

	selectedBranches := make([]Branch, k)

	copy(selectedBranches, branches)
	for i := len(selectedBranches) - 1; i > 0; i-- {
		j := randGen.Intn(i + 1)
		selectedBranches[i], selectedBranches[j] = selectedBranches[j], selectedBranches[i]
	}

	return selectedBranches
}

// Function to get the main branch (main or master)
func getMainOrMasterBranch(branches []Branch) Branch {

	for _, branch := range branches {
		if branch.Name == "main" || branch.Name == "master" {
			return branch
		}
	}

	return Branch{}
}

func extractResetTime(errorMessage string) string {
	re := regexp.MustCompile(`\[rate reset in (\d+h\d+m\d+s)\]`)
	matches := re.FindStringSubmatch(errorMessage)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func SaveResult(result AnalysisResult) error {
	// Open or create the file
	file, err := os.Create("Results/config/analysis_repos_github.json")
	if err != nil {
		fmt.Println("❌ Error creating Analysis file:", err)
		return err
	}
	defer file.Close()

	// Create a JSON encoder
	encoder := json.NewEncoder(file)

	// Encode the result and write it to the file
	if err := encoder.Encode(result); err != nil {
		fmt.Println("❌ Error encoding JSON file <Results/config/analysis_repos_github.json> :", err)
		return err
	}

	fmt.Println("✅ Result saved successfully!")
	return nil
}

func SaveBranch(branch []Branch) error {
	// Open or create the file
	file, err := os.Create("Results/config/analysis_branch_github.json")
	if err != nil {
		fmt.Println("❌ Error creating Analysis Branch file:", err)
		return err
	}
	defer file.Close()

	// Create a JSON encoder
	encoder := json.NewEncoder(file)

	// Encode the Branch and write it to the file
	if err := encoder.Encode(branch); err != nil {
		fmt.Println("❌ Error encoding JSON file <Results/config/analysis_branch_github.json> :", err)
		return err
	}

	fmt.Println("✅ Branch saved successfully!")
	return nil
}

func SaveRepos(repos []Repository) error {
	// Open or create the file
	file, err := os.Create("Results/config/analysis_repos_github.json")
	if err != nil {
		fmt.Println("❌ Error creating Analysis Repos file:", err)
		return err
	}
	defer file.Close()

	// Create a JSON encoder
	encoder := json.NewEncoder(file)

	// Encode the Branch and write it to the file
	if err := encoder.Encode(repos); err != nil {
		fmt.Println("❌ Error encoding JSON file <Results/config/analysis_repos_github.json> :", err)
		return err
	}

	fmt.Println("✅ Repos saved successfully!")
	return nil
}

func SaveLast(last Lastanalyse) error {
	// Open or create the file
	file, err := os.Create("Results/config/analysis_last_github.json")
	if err != nil {
		fmt.Println("❌ Error creating Analysis Last file:", err)
		return err
	}
	defer file.Close()

	// Create a JSON encoder
	encoder := json.NewEncoder(file)

	// Encode the Branch and write it to the file
	if err := encoder.Encode(last); err != nil {
		fmt.Println("❌ Error encoding JSON file <Results/config/analysis_last_github.json> :", err)
		return err
	}

	fmt.Println("✅ Last saved successfully!")
	return nil
}

/* func GetReposGithub1(parms ParamsReposGithub) ([]ProjectBranch, int, int) {

var largestRepoSize int
var largestRepoBranch string
var importantBranches []ProjectBranch
var message4 string
cpt := 1
emptyRepo := 0
result := AnalysisResult{}

parms.Spin.Stop()

spin1 := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
//spin1.Prefix = PrefixMsg
spin1.Color("green", "bold")

message4 = "Repo(s)"

fmt.Printf("\t  ✅ The number of %s found is: %d\n", message4, parms.NBRepos)

ctx := context.Background()
client := github.NewClient(nil).WithAuthToken(parms.AccessToken)

for _, repo := range parms.Repos {
	largestRepoSize = 0
	largestRepoBranch = ""
	var branches []Branch
	var branchesAPI []Branch
	var branchInfos []BranchInfo
	//var TopBranches []Branch
	var Nobranch int = 0

	// Test if Repository is empty
	isEmpty, err := isRepositoryEmpty(parms.URL, parms.Apiver, parms.Organization, repo.Name, parms.AccessToken)
	if err != nil {
		fmt.Printf("❌ Error when Testing if repo is empty %s: %v\n", repo.Name, err)
		//spin1.Stop()
		continue
	}

	if !isEmpty {

		// Test if we pass branch name as a parameter in the config file
		if len(parms.Branch) == 0 {

			urlrepos := fmt.Sprintf("%srepos/%s/%s/branches?per_page=100&page=1", parms.URL, parms.Organization, repo.Name)
			branchesAPI, err = GithubAllBranches(urlrepos, parms.AccessToken, parms.Apiver)
			if err != nil {
				fmt.Printf("❌ Error when retrieving branches for repo %s: %v\n", repo.Name, err)
				//spin1.Stop()
				continue
			}

			if len(branchesAPI) > parms.Branchescount {

				/*	branches = selectRandomBranches(branchesAPI, parms.Branchescount)
					// Also add the main branch (main or master) to the selection
					branches = append(branches, getMainOrMasterBranch(branchesAPI))*/

// Browse branches to get their activity (number of commits)
/*	for _, branch := range branches {
						messageB := fmt.Sprintf("\t   Analysis top branch(es) in repository <%s> ...", repo.Name)
						spin1.Prefix = messageB
						spin1.Start()
						// Retrieve branch commits
						commits, _, err := client.Repositories.ListCommits(ctx, parms.Organization, repo.Name, &github.CommitsListOptions{
							SHA: branch.Name,
						})
						if err != nil {
							log.Printf("❌ Error retrieving commits for branch %s : %v", branch.Name, err)
							continue
						}

						// Store branch information
						branchInfos = append(branchInfos, BranchInfo{
							Name:     branch.Name,
							Activity: len(commits), // Number of commits on branch
						})
					}

					// Sort branches based on their activity (number of commits)
					sort.Slice(branchInfos, func(i, j int) bool {
						return branchInfos[i].Activity > branchInfos[j].Activity
					})
					spin1.Stop()

				} else {

					branches = append(branchesAPI, getMainOrMasterBranch(branchesAPI))

				}

			} else {

				branches, err = ifExistBranches(parms.Organization, repo.Name, parms.Branch, parms.AccessToken, parms.URL, parms.Apiver)
				if err != nil {
					fmt.Printf("❗️ The branch <%s> for repository %s not exist - check your config.json file : \n", parms.Branch, repo.Name)
					Nobranch = 1
					continue

				}

			}
			if Nobranch == 0 {

				// Finding the branch with the largest size
				if len(branches) > 1 {
					if len(branchesAPI) > parms.Branchescount {
						fmt.Printf("\r\t\t✅ %d Repo: %s - Number of branches: %d - Maxi top five analysis is : %d\n", cpt, repo.Name, len(branchesAPI), parms.Branchescount+1)
					} else {
						fmt.Printf("\r\t\t✅ %d Repo: %s - Number of branches: %d \n", cpt, repo.Name, len(branches))
					}
					//for _, branch := range branches {
					for i := 0; i < parms.Branchescount+1 && i < len(branches); i++ {
						messageB := fmt.Sprintf("\t   Analysis branch <%s> size...", branches[i].Name)
						spin1.Prefix = messageB
						spin1.Start()

						size, err := fetchBranchSizeGithub(parms.Organization, repo.Name, branches[i].Name, parms.AccessToken, parms.URL, parms.Apiver)
						messageF := ""
						spin1.FinalMSG = messageF

						spin1.Stop()
						if err != nil {
							fmt.Printf("❌ Error retrieving branch <%s> size: %v", branches[i].Name, err)
							spin1.Stop()
							os.Exit(1)
						}

						if size > largestRepoSize {
							largestRepoSize = size
							//largestRepoProject = project.Name
							largestRepoBranch = branches[i].Name
						}

					}
				} else {
					fmt.Printf("\r\t\t✅ %d Repo: %s - Number of branches: %d \n", cpt, repo.Name, len(branchesAPI))

					size1, err1 := fetchBranchSizeGithub(parms.Organization, repo.Name, branches[0].Name, parms.AccessToken, parms.URL, parms.Apiver)

					if err1 != nil {
						fmt.Println("\n❌ Error retrieving branch size:", err1)
						spin1.Stop()
						os.Exit(1)
					}
					largestRepoSize = size1
					largestRepoBranch = branches[0].Name
				}

				importantBranches = append(importantBranches, ProjectBranch{
					Org:         parms.Organization,
					RepoSlug:    repo.Name,
					MainBranch:  largestRepoBranch,
					LargestSize: largestRepoSize,
				})
				Nobranch = 0
			}
		} else {
			emptyRepo++
			Nobranch = 0
		}
		cpt++
	}

	//result.NumProjects = len(parms.Projects)
	result.NumRepositories = parms.NBRepos
	result.ProjectBranches = importantBranches

	// Save Result of Analysis

	file, err := os.Create("Results/config/analysis_repos_github.json")
	if err != nil {
		fmt.Println("❌ Error creating Analysis file:", err)
		return importantBranches, emptyRepo, parms.NBRepos
	}
	defer file.Close()
	encoder := json.NewEncoder(file)

	err = encoder.Encode(result)
	if err != nil {
		fmt.Println("Error encoding JSON file <Results/config/analysis_repos_github.json> :", err)
		return importantBranches, emptyRepo, parms.NBRepos
	}
	return importantBranches, emptyRepo, parms.NBRepos
}*/

func GetReposGithub(parms ParamsReposGithub) ([]ProjectBranch, int, int, int) {

	var largestRepoSize int
	var TotalBranches int
	var largestRepoBranch string
	var importantBranches []ProjectBranch
	var message4 string
	cpt := 1
	emptyRepo := 0
	result := AnalysisResult{}

	parms.Spin.Stop()

	spin1 := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	//spin1.Prefix = PrefixMsg
	spin1.Color("green", "bold")

	message4 = "Repo(s)"

	fmt.Printf("\t  ✅ The number of %s found is: %d\n", message4, parms.NBRepos)

	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(parms.AccessToken)

	for _, repo := range parms.Repos {
		largestRepoSize = 0
		largestRepoBranch = ""
		var branches []Branch
		//var branchesAPI []Branch
		var branchInfos []BranchInfo
		//var TopBranches []Branch
		var Nobranch int = 0

		// Test if Repository is empty
		isEmpty, err := isRepositoryEmpty(parms.URL, parms.Apiver, parms.Organization, repo.Name, parms.AccessToken)
		if err != nil {
			fmt.Printf("❌ Error when Testing if repo is empty %s: %v\n", repo.Name, err)
			//spin1.Stop()
			continue
		}

		if !isEmpty {

			// Test if we pass branch name as a parameter in the config file
			if len(parms.Branch) == 0 {

				urlrepos := fmt.Sprintf("%srepos/%s/%s/branches?per_page=100&page=1", parms.URL, parms.Organization, repo.Name)
				branches, err = GithubAllBranches(urlrepos, parms.AccessToken, parms.Apiver)
				if err != nil {
					fmt.Printf("❌ Error when retrieving branches for repo %s: %v\n", repo.Name, err)
					//spin1.Stop()
					continue
				}
				TotalBranches = TotalBranches + len(branches)

				// Browse branches to get their activity (number of commits)
				for _, branch := range branches {
					messageB := fmt.Sprintf("\t   Analysis top branch(es) in repository <%s> ...", repo.Name)
					spin1.Prefix = messageB
					spin1.Start()
					// Retrieve branch commits
					commits, _, err := client.Repositories.ListCommits(ctx, parms.Organization, repo.Name, &github.CommitsListOptions{
						SHA: branch.Name,
					})
					if err != nil {
						if strings.Contains(err.Error(), "API rate limit of") {
							resetTime := extractResetTime(err.Error())
							fmt.Printf("\n❗️ Sorry, you have exceeded the GitHub API call rate limit. Please wait a few moments before trying again : %s.\n", resetTime)
							fmt.Printf("❌ Stop step %d analysis for the repository %s and the branch %s", cpt, repo.Name, branch.Name)

							/*	importantBranches = append(importantBranches, ProjectBranch{
									Org:         parms.Organization,
									RepoSlug:    repo.Name,
									MainBranch:  largestRepoBranch,
									LargestSize: largestRepoSize,
								})
								result.NumRepositories = parms.NBRepos
								result.ProjectBranches = importantBranches*/

							// Save Result of Analysis
							err := SaveResult(result)
							if err != nil {
								fmt.Println("❌ Error Save Result of Analysis :", err)
								os.Exit(1)

							}

							// Save Repos
							err = SaveRepos(parms.Repos)
							if err != nil {
								fmt.Println("❌ Error Save Repos of Analysis :", err)
								os.Exit(1)

							}

							// Save Branches
							err = SaveBranch(branches)
							if err != nil {
								fmt.Println("❌ Error Save Branch of Analysis :", err)
								os.Exit(1)
							}

							var lastanalyse Lastanalyse

							lastanalyse.LastRepos = repo.Name
							lastanalyse.LastBranch = branch.Name
							// Save Last
							err = SaveLast(lastanalyse)
							if err != nil {
								fmt.Println("❌ Error Save last of Analysis :", err)
								os.Exit(1)
							}

							os.Exit(1)
						}
						continue
					}

					// Store branch information
					branchInfos = append(branchInfos, BranchInfo{
						Name:     branch.Name,
						Activity: len(commits), // Number of commits on branch
					})
				}

				// Sort branches based on their activity (number of commits)
				sort.Slice(branchInfos, func(i, j int) bool {
					return branchInfos[i].Activity > branchInfos[j].Activity
				})
				spin1.Stop()

			} else {

				branches, err = ifExistBranches(parms.Organization, repo.Name, parms.Branch, parms.AccessToken, parms.URL, parms.Apiver)
				if err != nil {
					fmt.Printf("❗️ The branch <%s> for repository %s not exist - check your config.json file : \n", parms.Branch, repo.Name)
					Nobranch = 1
					continue

				}

			}
			if Nobranch == 0 {

				// Finding the branch with the largest size
				//	if len(branches) > 1 {

				fmt.Printf("\r\t\t✅ %d Repo: %s - Number of branches: %d \n", cpt, repo.Name, len(branches))

				//	largestRepoSize = size1
				largestRepoBranch = branchInfos[0].Name
				/*	} else {
					fmt.Printf("\r\t\t✅ %d Repo: %s - Number of branches: %d \n", cpt, repo.Name, len(branches))

					size1, err1 := fetchBranchSizeGithub(parms.Organization, repo.Name, branchInfos[0].Name, parms.AccessToken, parms.URL, parms.Apiver)

					if err1 != nil {
						fmt.Println("\n❌ Error retrieving branch size:", err1)
						spin1.Stop()
						os.Exit(1)
					}
					largestRepoSize = size1
					largestRepoBranch = branchInfos[0].Name
				}*/

				importantBranches = append(importantBranches, ProjectBranch{
					Org:         parms.Organization,
					RepoSlug:    repo.Name,
					MainBranch:  largestRepoBranch,
					LargestSize: largestRepoSize,
				})
				Nobranch = 0
			}
		} else {
			emptyRepo++
			Nobranch = 0
		}
		cpt++
	}

	//result.NumProjects = len(parms.Projects)
	result.NumRepositories = parms.NBRepos
	result.ProjectBranches = importantBranches

	// Save Result of Analysis
	err := SaveResult(result)
	if err != nil {
		fmt.Println("❌ Error Save Result of Analysis :", err)
		os.Exit(1)

	}

	/*	file, err := os.Create("Results/config/analysis_repos_github.json")
		if err != nil {
			fmt.Println("❌ Error creating Analysis file:", err)
			return importantBranches, emptyRepo, parms.NBRepos, TotalBranches
		}
		defer file.Close()
		encoder := json.NewEncoder(file)

		err = encoder.Encode(result)
		if err != nil {
			fmt.Println("❌ Error encoding JSON file <Results/config/analysis_repos_github.json> :", err)
			return importantBranches, emptyRepo, parms.NBRepos, TotalBranches
		}*/
	return importantBranches, emptyRepo, parms.NBRepos, TotalBranches
}

// Get Infos for all Repositories in Organization
func GetRepoGithubList(url, baseapi, apiver, accessToken, organization, exlusionfile, repos, branchmain string, topbranchescount int) ([]ProjectBranch, error) {

	var largestRepoSize int
	var totalSize int
	var largestRepoBranch, largesRepo string
	//var repositories []Repository
	var importantBranches []ProjectBranch
	var exclusionList *ExclusionList
	var err1 error
	var emptyRepo int
	var TotalBranches int
	nbRepos := 0

	fmt.Print("\n🔎 Analysis of devops platform objects ...\n")

	spin := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	spin.Prefix = PrefixMsg
	spin.Color("green", "bold")
	spin.Start()
	// Test if exclusion file exist
	if exlusionfile == "0" {
		exclusionList = &ExclusionList{
			Repos: make(map[string]bool),
		}

	} else {
		exclusionList, err1 = loadExclusionList(exlusionfile)
		if err1 != nil {
			fmt.Printf("\n❌ Error Read Exclusion File <%s>: %v", exlusionfile, err1)
			spin.Stop()
			return nil, err1
		}

	}

	if len(repos) == 0 {

		urlrepo := fmt.Sprintf("%sorgs/%s/repos?type=all&recurse_submodules=false&per_page=1000&page=1", url, organization)

		repositories, err := fetchRepositoriesAllGithub(urlrepo, accessToken, apiver, exclusionList)
		if err != nil {
			fmt.Printf("❌ Error fetching repositories: %v\n", err)
			return importantBranches, nil
		}

		parms := ParamsReposGithub{
			Repos:         repositories,
			URL:           url,
			BaseAPI:       baseapi,
			Apiver:        apiver,
			AccessToken:   accessToken,
			Organization:  organization,
			NBRepos:       len(repositories),
			ExclusionList: exclusionList,
			Spin:          spin,
			Branch:        branchmain,
			Branchescount: topbranchescount,
		}

		importantBranches, emptyRepo, nbRepos, TotalBranches = GetReposGithub(parms)
		//fmt.Printf("Total repositories in %s: %d\n", organization, len(repositories))

	}

	largestRepoSize = 0
	largestRepoBranch = ""
	largesRepo = ""

	for _, branch := range importantBranches {
		if branch.LargestSize > largestRepoSize {
			largestRepoSize = branch.LargestSize
			largestRepoBranch = branch.MainBranch
			largesRepo = branch.RepoSlug
		}
		totalSize += branch.LargestSize
	}
	totalSizeMB := utils.FormatSize(int64(totalSize))
	largestRepoSizeMB := utils.FormatSize(int64(largestRepoSize))

	fmt.Printf("\n✅ The largest repo is <%s> in the organization <%s> with the branch <%s> and a size of ≃ %s\n", largesRepo, organization, largestRepoBranch, largestRepoSizeMB)
	fmt.Printf("\r✅ Total size of your organization's repositories ≃ %s\n", totalSizeMB)
	fmt.Printf("\r✅ Total repositories that will be analyzed: %d - Find empty : %d\n", nbRepos-emptyRepo, emptyRepo)
	fmt.Printf("\r✅ Total Branches that will be analyzed: %d\n", TotalBranches)
	//os.Exit(1)
	return importantBranches, nil
}

func isRepositoryEmpty(urlrepo, apiver, org, repo, accessToken string) (bool, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contents", urlrepo, org, repo)

	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-GitHub-Api-Version", apiver)

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	} else if resp.StatusCode == http.StatusOK {
		return true, nil
	} else {
		return true, fmt.Errorf("\n❌ Failed to check repository. Status code: %d", resp.StatusCode)
	}
}

func fetchRepositoriesAllGithub(url, token, apiver string, exclusionList *ExclusionList) ([]Repository, error) {

	var allRepos []Repository

	for {

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "token "+token)
		req.Header.Set("X-GitHub-Api-Version", apiver)

		//  Check HTTP status code
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		// Check HTTP status code
		if resp.StatusCode != http.StatusOK {
			mess := fmt.Sprintf("\n❌ Request <fetchRepositoriesAllGithub> failed with status: %d", resp.StatusCode)
			return nil, fmt.Errorf(mess)
		}

		// Decoding the JSON response
		var repositories []Repository
		if err := json.NewDecoder(resp.Body).Decode(&repositories); err != nil {
			return nil, err
		}

		// Add the current page's repositories to the total list
		// Filter archived repositories
		for _, repo := range repositories {
			if repoIsArchived(&repo) || repoIsExcluded(&repo, exclusionList) {
				continue
			}
			allRepos = append(allRepos, repo)
		}

		//  Check for next page
		nextPage := getNextPage(resp.Header)
		if nextPage == "" {
			break
		}

		url = nextPage

	}

	return allRepos, nil
}

func GithubAllBranches(url, AccessToken, apiver string) ([]Branch, error) {

	client := http.Client{}
	var branches []Branch

	for {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		req.Header.Set("Authorization", "token "+AccessToken)
		req.Header.Set("X-GitHub-Api-Version", apiver)

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("\n❌ Failed to list branches. Status code: %d", resp.StatusCode)
		}

		var branchList []Branch
		err = json.NewDecoder(resp.Body).Decode(&branchList)
		if err != nil {
			return nil, err
		}
		branches = append(branches, branchList...)

		nextPageURL := getNextPage(resp.Header)
		if nextPageURL == "" {
			break
		}
		url = nextPageURL
	}

	return branches, nil
}

func fetchBranchSizeGithub(org, repoName, branchName, accessToken, urlrepo, apiver string) (int, error) {
	branchNameEncoded := url.QueryEscape(branchName)
	url := fmt.Sprintf("%srepos/%s/%s/git/trees/%s?recursive=1&per_page=100&page=1", urlrepo, org, repoName, branchNameEncoded)

	totalBranchSize := 0

	for {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return 0, err
		}
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("X-GitHub-Api-Version", apiver)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return 0, err
		}
		defer resp.Body.Close()

		// Check HTTP status code
		if resp.StatusCode != http.StatusOK {
			mess := fmt.Sprintf("\n❌ Request failed with status: %d - requets : %s ", resp.StatusCode, url)
			return 0, fmt.Errorf(mess)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0, err
		}

		var treeResponse TreeResponse
		err = json.Unmarshal(body, &treeResponse)
		if err != nil {
			return 0, err
		}

		// Calculate branch size
		for _, item := range treeResponse.Tree {
			if item.Type == "blob" {
				totalBranchSize += item.Size
			}
		}

		nextPageURL := getNextPage(resp.Header)
		if nextPageURL == "" {
			break
		}
		url = nextPageURL
	}

	return totalBranchSize, nil
}

func ifExistBranches(org, repo, branch, accessToken, urlb, apiver string) ([]Branch, error) {

	url := fmt.Sprintf("%srepos/%s/%s/branches/%s", urlb, org, repo, branch)

	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-GitHub-Api-Version", apiver)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	} else if resp.StatusCode == http.StatusOK {
		var branch Branch
		err := json.NewDecoder(resp.Body).Decode(&branch)
		if err != nil {
			return nil, err
		}
		return []Branch{branch}, nil
	} else {
		return nil, fmt.Errorf("\n❌ Failed to check branch existence. Status code: %d", resp.StatusCode)
	}
}

// manage pagination
func getNextPage(header http.Header) string {
	linkHeader := header.Get("Link")
	if linkHeader == "" {
		return ""
	}

	links := strings.Split(linkHeader, ",")
	for _, link := range links {
		parts := strings.Split(strings.TrimSpace(link), ";")
		if len(parts) == 2 && strings.Contains(parts[1], `rel="next"`) {
			return strings.Trim(parts[0], "<>")
		}
	}

	return ""
}

// Function to check if a repository is archived
func repoIsArchived(repo *Repository) bool {

	return repo != nil && repo.Path != "" && repo.Archived
}

// Function to check whether a deposit is in the exclusion list
func repoIsExcluded(repo *Repository, exclusionList *ExclusionList) bool {
	if exclusionList == nil || exclusionList.Repos == nil {
		return false
	}
	// Checks if the deposit is in the exclusion map
	_, exists := exclusionList.Repos[repo.Path]
	return exists
}

func fetchLanguagesGithub(org, repoName, accessToken, urlrepo string) (map[string]int, error) {

	url := fmt.Sprintf("%srepos/%s/%s/languages", urlrepo, org, repoName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch languages. Status code: %d", resp.StatusCode)
	}

	var languages map[string]int
	if err := json.NewDecoder(resp.Body).Decode(&languages); err != nil {
		return nil, err
	}

	return languages, nil
}

/* func updateRepositoryWithLanguages(repo *Repository, languages map[string]int) {

	repo.LOC = make(map[string]int)

	for lang, loc := range languages {

		langLower := strings.ToLower(lang)

		if _, ok := Languages[langLower]; ok {
			Languages.language
			repo.LOC[langLower] = loc
		}
	}
}*/
