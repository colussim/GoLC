package getazure

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
)

type ProjectBranch struct {
	Org         string
	Namespace   string
	RepoSlug    string
	MainBranch  string
	LargestSize int
}

/*type ExclusionList struct {
	Repos map[string]bool `json:"repos"`
}*/

type AnalysisResult struct {
	NumRepositories int
	ProjectBranches []ProjectBranch
}

type ExclusionList struct {
	Projects map[string]bool
	Repos    map[string]bool
}

type AnalyzeProject struct {
	Project       core.TeamProjectReference
	AzureClient   core.Client
	Context       context.Context
	ExclusionList *ExclusionList
	Spin1         *spinner.Spinner
	Org           string
}

type ParamsProjectAzure struct {
	Client         core.Client
	Context        context.Context
	Projects       []core.TeamProjectReference
	URL            string
	AccessToken    string
	ApiURL         string
	Organization   string
	Exclusionlist  *ExclusionList
	Excludeproject int
	Spin           *spinner.Spinner
	Period         int
	Stats          bool
	DefaultB       bool
	SingleRepos    string
	SingleBranch   string
}

// RepositoryMap represents a map of repositories to ignore
type ExclusionRepos map[string]bool

const PrefixMsg = "Get Project(s)..."
const MessageErro1 = "/\n❌ Failed to list projects for organization %s: %v\n"
const Message1 = "\t  ✅ The number of %s found is: %d\n"
const Message2 = "\t   Analysis top branch(es) in project <%s> ..."
const Message3 = "\r\t\t✅ %d Project: %s - Number of branches: %d - largest Branch: %s \n"
const Message4 = "Project(s)"

// Load repository ignore map from file
/*func loadExclusionRepos(filename string) (ExclusionRepos, error) {

	ignoreMap := make(ExclusionRepos)

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		repoName := strings.TrimSpace(scanner.Text())
		if repoName != "" {
			ignoreMap[repoName] = true

		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ignoreMap, nil
}*/

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
			// Exclusion de projet
			exclusionList.Projects[parts[0]] = true
		} else if len(parts) == 2 {
			// Exclusion de répertoire
			exclusionList.Repos[line] = true
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return exclusionList, nil
}

func loadExclusionFileOrCreateNew(exclusionFile string) (*ExclusionList, error) {
	if exclusionFile == "0" {
		return &ExclusionList{
			Projects: make(map[string]bool),
			Repos:    make(map[string]bool),
		}, nil
	}
	return loadExclusionList(exclusionFile)
}

func isRepoExcluded(exclusionList *ExclusionList, projectKey, repoKey string) bool {
	_, repoExcluded := exclusionList.Repos[projectKey+"/"+repoKey]
	return repoExcluded
}

// Fonction pour vérifier si un projet est exclu
func isProjectExcluded(exclusionList *ExclusionList, projectKey string) bool {
	_, projectExcluded := exclusionList.Projects[projectKey]
	return projectExcluded
}

// Function to check if a project should be excluded from analysis
func isExcluded(projectName string, exclusionList map[string]bool) bool {

	if _, ok := exclusionList[projectName]; ok {
		return true
	}

	// Check for subdomain match
	for excludedRepo := range exclusionList {
		if strings.HasPrefix(projectName, excludedRepo) {
			return true
		}
	}

	return false

}

func getAllProjects(ctx context.Context, coreClient core.Client, exclusionList *ExclusionList) ([]core.TeamProjectReference, error) {
	var allProjects []core.TeamProjectReference

	// Récupération de la première page
	responseValue, err := coreClient.GetProjects(ctx, core.GetProjectsArgs{})
	if err != nil {
		return nil, err
	}

	//index := 0
	for responseValue != nil {
		// Ajouter les projets de la page actuelle
		allProjects = append(allProjects, responseValue.Value...)

		// Si continuationToken a une valeur, il y a au moins une autre page de projets à obtenir
		if responseValue.ContinuationToken != "" {
			continuationToken, err := strconv.Atoi(responseValue.ContinuationToken)
			if err != nil {
				return nil, err
			}

			// Récupération de la page suivante des projets
			projectArgs := core.GetProjectsArgs{
				ContinuationToken: &continuationToken,
			}
			responseValue, err = coreClient.GetProjects(ctx, projectArgs)
			if err != nil {
				return nil, err
			}
		} else {
			// S'il n'y a pas de continuationToken, il n'y a plus de pages à récupérer
			responseValue = nil
		}
	}

	return allProjects, nil
}

func GetRepoAzureList(platformConfig map[string]interface{}, exclusionFile string) ([]ProjectBranch, error) {

	var importantBranches []ProjectBranch
	var totalExclude, totalArchiv, emptyRepo, TotalBranches, exludedprojects int
	var nbRepos int
	//	var emptyRepos, archivedRepos int
	//	var TotalBranches int = 0 // Counter Number of Branches on All Repositories
	var exclusionList *ExclusionList
	var err error

	//var totalExclude, totalArchiv, emptyRepo, TotalBranches, exludedprojects int
	//var nbRepos int

	//	var totalSize int

	//	excludedProjects := 0
	//	result := AnalysisResult{}

	// Calculating the period
	//	until := time.Now()
	//	since := until.AddDate(0, int(platformConfig["Period"].(float64)), 0)
	ApiURL := platformConfig["Url"].(string) + platformConfig["Organization"].(string)

	fmt.Print("\n🔎 Analysis of devops platform objects ...\n")

	spin := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	spin.Prefix = PrefixMsg
	spin.Color("green", "bold")
	spin.Start()

	exclusionList, err = loadExclusionFileOrCreateNew(exclusionFile)
	if err != nil {
		fmt.Printf("\n❌ Error Read Exclusion File <%s>: %v", exclusionFile, err)
		spin.Stop()
		return nil, err
	}

	// Create a connection to your organization
	connection := azuredevops.NewPatConnection(ApiURL, platformConfig["AccessToken"].(string))
	ctx := context.Background()

	// Create a client to interact with the Core area
	coreClient, err := core.NewClient(ctx, connection)
	if err != nil {
		log.Fatal(err)
	}

	gitClient, err := git.NewClient(ctx, connection)
	if err != nil {
		log.Fatalf("Erreur lors de la création du client Git: %v", err)
	}

	if platformConfig["DefaultBranch"].(bool) {
		//	cpt := 1

		/* --------------------- Analysis all projects with a default branche  ---------------------  */
		if platformConfig["Project"].(string) == "" {
			//pageSize := 100

			projects, err := getAllProjects(ctx, coreClient, exclusionList)
			exludedprojects := 0
			if err != nil {
				fmt.Println("\n\nError quit")
				log.Fatalf(MessageErro1, platformConfig["Organization"].(string), err)
			}
			spin.Stop()
			spin1 := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
			spin1.Color("green", "bold")

			fmt.Printf(Message1, Message4, len(projects))

			params := getCommonParams(ctx, coreClient, platformConfig, projects, exclusionList, exludedprojects, spin, ApiURL)
			importantBranches, emptyRepo, nbRepos, TotalBranches, totalExclude, totalArchiv, err = getRepoAnalyse(params, gitClient)
			if err != nil {
				spin.Stop()
				return nil, err
			}
			fmt.Println("debug %d %d %d %d %d %d", totalExclude, totalArchiv, emptyRepo, TotalBranches, exludedprojects, nbRepos)

			for _, project := range projects {
				fmt.Printf("Projet: %s (ID: %s)\n", *project.Name, *project.Id)
			}

		}

	}
	os.Exit(1)

	return importantBranches, nil
}

func getCommonParams(ctx context.Context, client core.Client, platformConfig map[string]interface{}, project []core.TeamProjectReference, exclusionList *ExclusionList, excludeproject int, spin *spinner.Spinner, apiURL string) ParamsProjectAzure {
	return ParamsProjectAzure{
		Client:   client,
		Context:  ctx,
		Projects: project,

		URL: platformConfig["Url"].(string),

		AccessToken:    platformConfig["AccessToken"].(string),
		ApiURL:         apiURL,
		Organization:   platformConfig["Organization"].(string),
		Exclusionlist:  exclusionList,
		Excludeproject: excludeproject,
		Spin:           spin,
		Period:         int(platformConfig["Period"].(float64)),
		Stats:          platformConfig["Stats"].(bool),
		DefaultB:       platformConfig["DefaultBranch"].(bool),
		SingleRepos:    platformConfig["Repos"].(string),
		SingleBranch:   platformConfig["Branch"].(string),
	}
}

func getRepoAnalyse(params ParamsProjectAzure, gitClient git.Client) ([]ProjectBranch, int, int, int, int, int, error) {

	var emptyRepos = 0
	var totalexclude = 0
	var importantBranches []ProjectBranch
	var NBRrepo, TotalBranches int
	var messageF = ""
	NBRrepos := 0
	cptarchiv := 0

	//cpt := 1

	message4 := "Repo(s)"

	spin1 := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	spin1.Prefix = PrefixMsg
	spin1.Color("green", "bold")

	params.Spin.Start()
	if params.Excludeproject > 0 {
		messageF = fmt.Sprintf("✅ The number of project(s) to analyze is %d - Excluded : %d\n", len(params.Projects), params.Excludeproject)
	} else {
		messageF = fmt.Sprintf("✅ The number of project(s) to analyze is %d\n", len(params.Projects))
	}
	params.Spin.FinalMSG = messageF
	params.Spin.Stop()

	// Get Repository in each Project
	for _, project := range params.Projects {

		fmt.Printf("\n\t🟢  Analyse Projet: %s \n", project.Name)

		emptyOrArchivedCount, excludedCount, repos, err := listReposForProject(params, *project.Name, gitClient)
		if err != nil {
			if len(params.SingleRepos) == 0 {
				fmt.Println("\r❌ Get Repos for each Project:", err)
				spin1.Stop()
				continue
			} else {
				errmessage := fmt.Sprintf(" Get Repo %s for Project %s %v", params.SingleRepos, *project.Name, err)
				spin1.Stop()
				return importantBranches, emptyRepos, NBRrepos, TotalBranches, totalexclude, cptarchiv, fmt.Errorf(errmessage)
			}
		}
		emptyRepos = emptyRepos + emptyOrArchivedCount
		totalexclude = totalexclude + excludedCount

		spin1.Stop()
		if emptyOrArchivedCount > 0 {
			NBRrepo = len(repos) + emptyOrArchivedCount
			fmt.Printf("\t  ✅ The number of %s found is: %d - Find empty %d:\n", message4, NBRrepo, emptyOrArchivedCount)
		} else {
			NBRrepo = len(repos)
			fmt.Printf("\t  ✅ The number of %s found is: %d\n", message4, NBRrepo)
		}

		/*	for _, repo := range repos {

			largestRepoBranch, repobranches, brsize, err := analyzeRepoBranches(params, repo, cpt, spin1)
			if err != nil {
				largestRepoBranch = repo.Mainbranch.Name

			}

			importantBranches = append(importantBranches, ProjectBranch{
				Org:         params.Organization,
				ProjectKey:  project.Key,
				RepoSlug:    repo.Slug,
				MainBranch:  largestRepoBranch,
				LargestSize: brsize,
			})
			TotalBranches += len(repobranches)

			cpt++
		}*/

		NBRrepos += NBRrepo

	}
	os.Exit(1)
	return importantBranches, emptyRepos, NBRrepos, TotalBranches, totalexclude, cptarchiv, nil

}

// getRepositories récupère la liste des références de dépôts pour un projet donné en gérant la pagination
func getRepositories(ctx context.Context, gitClient git.Client, projectID string, exclusionList *ExclusionList) (int, int, int, []git.GitRepository, error) {
	var allRepos []git.GitRepository
	var archivedCount, emptyCount, excludedCount int
	pageSize := 100 // Nombre maximum de dépôts par page

	for skip := 0; ; skip += pageSize {
		repos, err := gitClient.GetRepositories(ctx, git.GetRepositoriesArgs{
			Project: &projectID,
			Top:     &pageSize,
			Skip:    &skip,
		})
		if err != nil {
			return 0, 0, 0, nil, err
		}

		if len(*repos) == 0 {
			break // Pas de dépôts sur cette page, fin de la pagination
		}

		for _, repo := range *repos {
			repoName := *repo.Name

			if isRepoExcluded(exclusionList, projectID, repoName) {
				excludedCount++
				continue // Ignorer le dépôt exclu
			}

			// Vérifier si le dépôt est archivé
			isArchived := repo.IsDisabled != nil && *repo.IsDisabled
			if isArchived {
				archivedCount++
				continue // Ignorer le dépôt archivé
			}

			// Obtenir les commits du dépôt pour vérifier s'il est vide
			commits, err := gitClient.GetCommits(ctx, git.GetCommitsArgs{
				RepositoryId: repo.Id,
				Project:      &projectID,
			})
			if err != nil {
				return 0, 0, 0, nil, err
			}

			if len(*commits) == 0 {
				emptyCount++
				continue // Ignorer le dépôt vide
			}

			allRepos = append(allRepos, repo)
		}
	}

	return archivedCount, emptyCount, excludedCount, allRepos, nil
}

func listReposForProject(parms ParamsProjectAzure, projectKey string, gitClient git.Client) (int, int, int, []git.GitRepository, error) {
	var allRepos []git.GitRepository
	var archivedCount, emptyCount, excludedCount int
	pageSize := 100 // Nombre maximum de dépôts par page

	for page := 0; ; page++ {
		// Récupérer la page actuelle des dépôts
		repos, err := gitClient.GetRepositories(parms.Context, git.GetRepositoriesArgs{
			Project: &projectID,
			Top:     &pageSize,
			Skip:    azuredevops.ToIntPtr(pageSize * page),
		})
		if err != nil {
			return 0, 0, 0, nil, err
		}

		// Vérifier si la page actuelle contient des dépôts
		if len(*repos) == 0 {
			break // Pas de dépôts sur cette page, fin de la pagination
		}

		// Parcourir tous les dépôts de la page actuelle
		for _, repo := range *repos {
			repoName := *repo.Name

			// Vérifier si le dépôt est exclu
			if isRepoExcluded(exclusionList, projectID, repoName) {
				excludedCount++
				continue // Ignorer le dépôt exclu
			}

			// Obtenir les détails du dépôt
			repository, err := gitClient.GetRepository(parms.Context, git.GetRepositoryArgs{
				RepositoryId: repo.Id,
				Project:      &projectID,
			})
			if err != nil {
				return 0, 0, 0, nil, err
			}

			// Vérifier si le dépôt est archivé
			isArchived := repository.Properties != nil && repository.Properties["isArchived"] == "true"

			if isArchived {
				archivedCount++
				continue // Ignorer le dépôt archivé
			}

			// Obtenir les commits du dépôt pour vérifier s'il est vide
			commits, err := gitClient.GetCommits(ctx, git.GetCommitsArgs{
				RepositoryId: repo.Id,
				Project:      &projectID,
			})
			if err != nil {
				return 0, 0, 0, nil, err
			}

			if len(*commits) == 0 {
				emptyCount++
				continue // Ignorer le dépôt vide
			}

			// Si le dépôt ne répond pas aux conditions d'exclusion, d'archivage ou de vide,
			// l'ajouter à la liste de tous les dépôts récupérés
			allRepos = append(allRepos, repo)
		}
	}

	return archivedCount, emptyCount, excludedCount, allRepos, nil
}

func listReposForProject1(parms ParamsProjectAzure, projectKey string, gitClient git.Client) (int, int, []git.GitRepository, error) {
	var allRepos []git.GitRepository

	var excludedCount, emptyOrArchivedCount int

	for {
		repos, err := gitClient.GetRepositories(parms.Context, git.GetRepositoriesArgs{
			Project: projectKey,
		})
		if err != nil {
			return nil, err
		}

		allRepos = append(allRepos, repos.Value...)

		if repos.ContinuationToken == "" {
			break
		}

		// Préparer la demande pour la page suivante
		continuationToken := repos.ContinuationToken
		opts := git.GetRepositoriesArgs{
			Project:           projectKey,
			ContinuationToken: &continuationToken,
		}
		repos, err = gitClient.GetRepositories(parms.Context, opts)
		if err != nil {
			return nil, err
		}
	}

	/*	repos, err := gitClient.GetRepositories(parms.Context, git.GetRepositoriesArgs{
			Project: projectKey,
		})
		if err != nil {
			return 0, 0, nil, err
		}*/

	/*

		page := 1
		for {
			repos, err := gitClient.GetRepositories(parms.Context, git.GetRepositoriesArgs{
				Project: projectKey,
			})
			if err != nil {
				return 0, 0, nil, err
			}

			eoc, exc, repos, err := listRepos(parms, projectKey, reposRes)
			if err != nil {
				return 0, 0, nil, err
			}
			emptyOrArchivedCount += eoc
			excludedCount += exc
			allRepos = append(allRepos, repos...)

			if len(reposRes.Items) < int(reposRes.Pagelen) {
				break
			}

			page++
		}*/

	return emptyOrArchivedCount, excludedCount, allRepos, nil
}
