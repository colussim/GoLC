package getgithub

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
)

const baseURL = "api.github.com"

type Repository struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Path          string `json:"full_name"`
	SizeR         int64  `json:"size"`
	Language      string `json:"language"`
	DefaultBranch string `json:"default_branch"`
}

type ExclusionList struct {
	Repos map[string]bool `json:"repos"`
}

type ParamsReposGithub struct {
	Repos         []Repository
	URL           string
	BaseAPI       string
	AccessToken   string
	Organization  string
	NBRepos       int
	ExclusionList *ExclusionList
	Spin          *spinner.Spinner
	Branch        string
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

func GetReposGithub(parms ParamsReposGithub) ([]ProjectBranch, int) {

	var largestRepoSize int
	var largestRepoBranch string
	var importantBranches []ProjectBranch
	var message4 string
	emptyRepo := 0
	result := AnalysisResult{}

	parms.Spin.Stop()

	message4 = "Repo(s)"

	fmt.Printf("\t  ✅ The number of %s found is: %d\n", message4, parms.NBRepos)

	for _, repo := range parms.Repos {
		largestRepoSize = 0
		largestRepoBranch = ""
		var branches []Branch
		var Nobranch int = 0

		isEmpty, err := isRepositoryEmpty(parms.Workspace, repo.Slug, parms.AccessToken, parms.BitbucketURLBase)
		if err != nil {
			fmt.Printf("❌ Error when Testing if repo is empty %s: %v\n", repo.Name, err)
			spin1.Stop()
			continue
		}

		if !isEmpty {

			if len(parms.Branch) == 0 {

				urlrepos := fmt.Sprintf("%s%s/repositories/%s/%s/refs/branches/?pagelen=100", parms.URL, parms.APIVersion, parms.Workspace, repo.Slug)

				branches, err = CloudAllBranches(urlrepos, parms.AccessToken)
				if err != nil {
					fmt.Printf("❌ Error when retrieving branches for repo %s: %v\n", repo.Name, err)
					spin1.Stop()
					continue
				}
			} else {
				urlrepos := fmt.Sprintf("%s%s/repositories/%s/%s/refs/branches/%s", parms.URL, parms.APIVersion, parms.Workspace, repo.Slug, parms.Branch)

				branches, err = ifExistBranches(urlrepos, parms.AccessToken)
				if err != nil {
					fmt.Printf("❗️ The branch <%s> for repository %s not exist - check your config.json file : \n", parms.Branch, repo.Name)
					Nobranch = 1
					continue

				}
			}
			if Nobranch == 0 {
				// Display Number of branches by repo
				fmt.Printf("\r\t✅ Repo: %s - Number of branches: %d\n", repo.Name, len(branches))

				// Finding the branch with the largest size
				if len(branches) > 1 {
					for _, branch := range branches {
						messageB := fmt.Sprintf("\t   Analysis branch <%s> size...", branch.Name)
						spin1.Prefix = messageB
						spin1.Start()

						size, err := fetchBranchSize(parms.Workspace, repo.Slug, branch.Name, parms.AccessToken, parms.URL, parms.APIVersion)
						messageF = ""
						spin1.FinalMSG = messageF

						spin1.Stop()
						if err != nil {
							fmt.Println("❌ Error retrieving branch size:", err)
							spin1.Stop()
							os.Exit(1)
						}

						if size > largestRepoSize {
							largestRepoSize = size
							//largestRepoProject = project.Name
							largestRepoBranch = branch.Name
						}

					}
				} else {
					size1, err1 := fetchBranchSize1(parms.Workspace, repo.Slug, parms.AccessToken, parms.URL, parms.APIVersion)

					if err1 != nil {
						fmt.Println("\n❌ Error retrieving branch size:", err1)
						spin1.Stop()
						os.Exit(1)
					}
					largestRepoSize = size1
					largestRepoBranch = branches[0].Name
				}

				importantBranches = append(importantBranches, ProjectBranch{
					ProjectKey:  project.Key,
					RepoSlug:    repo.Slug,
					MainBranch:  largestRepoBranch,
					LargestSize: largestRepoSize,
				})
				Nobranch = 0
			}
		} else {
			emptyRepo++
			Nobranch = 0
		}
	}

	result.NumProjects = len(parms.Projects)
	result.NumRepositories = parms.NBRepos
	result.ProjectBranches = importantBranches

	// Save Result of Analysis
	file, err := os.Create("Results/config/analysis_repos.json")
	if err != nil {
		fmt.Println("❌ Error creating Analysis file:", err)
		return importantBranches, parms.NBRepos, emptyRepo
	}
	defer file.Close()
	encoder := json.NewEncoder(file)

	err = encoder.Encode(result)
	if err != nil {
		fmt.Println("Error encoding JSON file <Results/config/analysis_repos.json> :", err)
		return importantBranches, parms.NBRepos, emptyRepo
	}
	return importantBranches, emptyRepo
}

// Get Infos for all Repositories in Organization for Main Branch
func GetRepoGithubList(url, baseapi, accessToken, organization, exlusionfile, repos, branchmain string) ([]Repository, error) {

	var repositories []Repository
	var exclusionList *ExclusionList
	var err1 error
	//nbRepos := 0

	fmt.Print("\n🔎 Analysis of devops platform objects ...\n")

	spin := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	spin.Prefix = PrefixMsg
	spin.Color("green", "bold")

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

		urlrepo := fmt.Sprintf("%sorgs/%s/repos?type=all&recurse_submodules=false&per_page=30&page=1", url, organization)

		repositories, err := fetchRepositoriesAllGithub(urlrepo, accessToken, exclusionList)
		if err != nil {
			fmt.Printf("❌ Error fetching repositories: %v\n", err)
			return repositories, nil
		}

		parms := ParamsReposGithub{
			Repos:         repositories,
			URL:           url,
			BaseAPI:       baseapi,
			AccessToken:   accessToken,
			Organization:  organization,
			NBRepos:       len(repositories),
			ExclusionList: exclusionList,
			Spin:          spin,
			Branch:        branchmain,
		}

		importantBranches, emptyRepo = GetReposGithub(parms)
		fmt.Printf("Total repositories in %s: %d\n", organization, len(repositories))
		os.Exit(1)
	}

	return repositories, nil
}

func fetchRepositoriesAllGithub(url, token string, exclusionList *ExclusionList) ([]Repository, error) {

	var allRepos []Repository

	for {

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "token "+token)

		//  Check HTTP status code
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		// Check HTTP status code
		if resp.StatusCode != http.StatusOK {
			mess := fmt.Sprintf("Request failed with status: %d", resp.StatusCode)
			return nil, fmt.Errorf(mess)
		}

		// Decoding the JSON response
		var repositories []Repository
		if err := json.NewDecoder(resp.Body).Decode(&repositories); err != nil {
			return nil, err
		}

		// Add the current page's repositories to the total list
		allRepos = append(allRepos, repositories...)

		//  Check for next page
		nextPage := getNextPage(resp.Header)
		if nextPage == "" {
			break
		}

		url = nextPage

	}

	return allRepos, nil
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
