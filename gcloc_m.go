package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/colussim/go-cloc/internal/constants"
	"github.com/colussim/go-cloc/pkg/devops/getgithub"
	"github.com/colussim/go-cloc/pkg/devops/getgitlab"
	getbibucketdc "github.com/colussim/go-cloc/pkg/devops/v1/getbitbucketdc"
	"github.com/colussim/go-cloc/pkg/gcloc"
	"github.com/colussim/go-cloc/pkg/utils"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type Repository struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	DefaultBranch string `json:"default_branch"`
	Path          string `json:"path"`
}

type ProjectBranch struct {
	ProjectKey  string
	RepoSlug    string
	MainBranch  string
	LargestSize int
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

// Declare a struct for Config fields
type Configuration struct {
	Users        string
	AccessToken  string
	Organization string
	DevOps       string
	Repos        string
	Branch       string
	Url          string
	Apiver       string
	Baseapi      string
}

type Report struct {
	TotalFiles      int `json:",omitempty"`
	TotalLines      int
	TotalBlankLines int
	TotalComments   int
	TotalCodeLines  int
	Results         interface{}
}

type Repository1 interface {
	GetID() int
	GetName() string
	GetPath() string
}

type Project1 interface {
	GetID() int
	GetName() string
}

func (p Project) GetName() string {
	return p.Name
}

func (p Project) GetId() int {
	return p.ID
}

type GithubRepository struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	DefaultBranch string `json:"default_branch"`
	Path          string `json:"full_name"`
	SizeR         int64  `json:"size"`
}

func (r GithubRepository) GetName() string {
	return r.Name
}
func (r GithubRepository) GetPath() string {
	return r.Path
}
func (r GithubRepository) GetID() int {
	return r.ID
}

// Implémentation pour getgitlab.Repository
type GitlabRepository struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	DefaultBranch string `json:"default_branch"`
	Path          string `json:"path_with_namespace"`
	Empty         bool   `json:"empty_repo"`
}

func (r GitlabRepository) GetName() string {
	return r.Name
}
func (r GitlabRepository) GetPath() string {
	return r.Path
}
func (r GitlabRepository) GetID() int {
	return r.ID
}

// Implémentation pour getbitbucket.Repository
type BitbucketRepository struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	DefaultBranch string `json:"default_branch"`
	Path          string `json:"path_with_namespace"`
	Empty         bool   `json:"empty_repo"`
}

func (r BitbucketRepository) GetName() string {
	return r.Name
}
func (r BitbucketRepository) GetPath() string {
	return r.Path
}
func (r BitbucketRepository) GetID() int {
	return r.ID
}

// Read Config file : Config.json
func GetConfig(configjs Configuration) Configuration {

	fconfig, err := os.ReadFile("config.json")
	if err != nil {
		panic("❌ Problem with the configuration file : config.json")
		os.Exit(1)
	}
	json.Unmarshal(fconfig, &configjs)
	return configjs
}

// Parse Result Files in JSON Format
func parseJSONFile(filePath, reponame string) int {
	file, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("❌ Error reading file:", err)
	}

	var report Report
	err = json.Unmarshal(file, &report)
	if err != nil {
		fmt.Println("❌ Error parsing JSON:", err)
	}

	//fmt.Printf("\nTotal Lines Of Code : %d\n\n", report.TotalCodeLines)

	return report.TotalCodeLines
}

func AnalyseReposList(DestinationResult string, Users string, AccessToken string, DevOps string, Organization string, repolist []Repository1) (cpt int) {
	var pathBranches string
	for _, repo := range repolist {
		fmt.Printf("\nAnalyse Repository : %s\n", repo.GetName())

		// Get All Branches in Repository
		branches, err := getgitlab.GetRepositoryBranches(AccessToken, repo.GetName(), repo.GetID())
		if err != nil {
			fmt.Println("\nError Analyse Repository Branches: ", err)
			os.Exit(1)
		}

		for _, branch := range branches {
			fmt.Printf("    - %s\n", branch.Name)

			pathBranches = fmt.Sprintf("?ref=%s", branch.Name)

			pathToScan := fmt.Sprintf("https://%s:%s@%s.com/%s%s", Users, AccessToken, DevOps, repo.GetPath(), pathBranches)
			fmt.Println("Scan PATH :", pathToScan)
			outputFileName := fmt.Sprintf("Result_%s", repo.GetName())

			params := gcloc.Params{
				Path:              pathToScan,
				ByFile:            false,
				ExcludePaths:      []string{},
				ExcludeExtensions: []string{},
				IncludeExtensions: []string{},
				OrderByLang:       false,
				OrderByFile:       false,
				OrderByCode:       false,
				OrderByLine:       false,
				OrderByBlank:      false,
				OrderByComment:    false,
				Order:             "DESC",
				OutputName:        outputFileName,
				OutputPath:        DestinationResult,
				ReportFormats:     []string{"json"},
			}

			gc, err := gcloc.NewGCloc(params, constants.Languages)
			if err != nil {
				fmt.Println("\nError Analyse Repositories: ", err)
				os.Exit(1)
			}

			gc.Run()
			cpt++

			// Remove Repository Directory
			err1 := os.RemoveAll(gc.Repopath)
			if err != nil {
				fmt.Printf("Error deleting Repository Directory: %v\n", err1)
				return
			}

		}
	}
	return cpt
}

func AnalyseRun(params gcloc.Params, reponame string) {
	gc, err := gcloc.NewGCloc(params, constants.Languages)
	if err != nil {
		fmt.Println("\nError Analyse Repositories: ", err)
		os.Exit(1)
	}

	gc.Run()
}

func AnalyseRepo(DestinationResult string, Users string, AccessToken string, DevOps string, Organization string, reponame string) (cpt int) {

	//pathToScan := fmt.Sprintf("git::https://%s@%s.com/%s/%s", AccessToken, DevOps, Organization, reponame)
	pathToScan := fmt.Sprintf("https://%s:%s@%s.com/%s/%s", Users, AccessToken, DevOps, Organization, reponame)

	outputFileName := fmt.Sprintf("Result_%s", reponame)
	params := gcloc.Params{
		Path:              pathToScan,
		ByFile:            false,
		ExcludePaths:      []string{},
		ExcludeExtensions: []string{},
		IncludeExtensions: []string{},
		OrderByLang:       false,
		OrderByFile:       false,
		OrderByCode:       false,
		OrderByLine:       false,
		OrderByBlank:      false,
		OrderByComment:    false,
		Order:             "DESC",
		OutputName:        outputFileName,
		OutputPath:        DestinationResult,
		ReportFormats:     []string{"json"},
	}
	gc, err := gcloc.NewGCloc(params, constants.Languages)
	if err != nil {
		fmt.Println("\nError Analyse Repositories: ", err)
		os.Exit(1)
	}

	gc.Run()
	cpt++

	// Remove Repository Directory
	err1 := os.RemoveAll(gc.Repopath)
	if err != nil {
		fmt.Printf("❌ Error deleting Repository Directory: %v\n", err1)
		return
	}

	return cpt
}

func main() {

	var config1 Configuration
	var AppConfig = GetConfig(config1)
	var largestLineCounter int
	var nameRepos2 string
	var NumberRepos int
	var fileexclusion = ".clocignore"

	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err)
	}
	DestinationResult := pwd + "/Results"
	_, err = os.Stat(DestinationResult)
	if err == nil {
		err := os.RemoveAll(DestinationResult)
		if err != nil {
			fmt.Printf("Error deleting directory: %s\n", err)
			os.Exit(1)
		}
		if err := os.MkdirAll(DestinationResult, os.ModePerm); err != nil {
			panic(err)
		}

	} else if os.IsNotExist(err) {
		if err := os.MkdirAll(DestinationResult, os.ModePerm); err != nil {
			panic(err)
		}

	}

	GlobalReport := DestinationResult + "/GlobalReport.txt"
	// Create Global Report File
	file, err := os.Create(GlobalReport)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// Test whether the analysis is for one or several repositories AppConfig.Repos != 0 -> 1 repository else more
	// Analyse 1 repository
	if len(AppConfig.Repos) != 0 {
		fmt.Printf("\nAnalyse %s Repository %s in Organization: %s\n", AppConfig.DevOps, AppConfig.Repos, AppConfig.Organization)
		NumberRepos = AnalyseRepo(DestinationResult, AppConfig.Users, AppConfig.AccessToken, AppConfig.DevOps, AppConfig.Organization, AppConfig.Repos)

	} else {
		// Analyse more repositories

		switch devops := AppConfig.DevOps; devops {
		case "github":
			var EmptyR = 0
			repositories, err := getgithub.GetRepoGithubList(AppConfig.AccessToken, AppConfig.Organization)
			if err != nil {
				fmt.Printf("Error Get Info Repositories in organization '%s' : '%s'", AppConfig.Organization, err)
				return
			}
			var repoList []Repository1

			for _, repo := range repositories {

				repoItem := GithubRepository{
					ID:            repo.ID,
					Name:          repo.Name,
					DefaultBranch: repo.DefaultBranch,
					Path:          repo.Path,
					SizeR:         repo.SizeR,
				}
				if repo.SizeR > 0 {
					ExcludeRepo, _ := utils.CheckCLOCignoreFile(fileexclusion, repo.Name)
					if !ExcludeRepo {
						repoList = append(repoList, repoItem)
					} else {
						fmt.Printf("\nRepository '%s' in Organization: '%s' is not Analyse because it is excluded\n", repo.Name, AppConfig.Organization)
						EmptyR = EmptyR + 1
					}
				} else {
					fmt.Print("\nRepository '%s' in Organization: '%s' is not Analyse because it is empty\n", repo.Name, AppConfig.Organization)
					EmptyR = EmptyR + 1
				}
			}

			NumberRepos1 := int(uintptr(len(repositories))) - EmptyR
			fmt.Printf("\nAnalyse '%s' Repositories('%d') in Organization: '%s'\n", AppConfig.DevOps, NumberRepos1, AppConfig.Organization)

			NumberRepos = AnalyseReposList(DestinationResult, AppConfig.Users, AppConfig.AccessToken, AppConfig.DevOps, AppConfig.Organization, repoList)

		case "gitlab":
			var EmptyR = 0
			repositories, err := getgitlab.GetRepoGitlabList(AppConfig.AccessToken, AppConfig.Organization)
			if err != nil {
				fmt.Printf("❌ Error Get Info Repositories in organization '%s' : '%s'", AppConfig.Organization, err)
				return
			}

			var repoList []Repository1
			for _, repo := range repositories {
				repoItem := GitlabRepository{
					ID:            repo.ID,
					Name:          repo.Name,
					DefaultBranch: repo.DefaultBranch,
					Path:          repo.Path,
					Empty:         repo.Empty,
				}
				if !repo.Empty {
					ExcludeRepo, _ := utils.CheckCLOCignoreFile(fileexclusion, repo.Name)
					if !ExcludeRepo {
						repoList = append(repoList, repoItem)
					} else {
						fmt.Print("\nRepository '%s' in Organization: '%s' is not Analyse because it is excluded\n", repo.Name, AppConfig.Organization)
						EmptyR = EmptyR + 1
					}
				} else {
					fmt.Printf("\nRepository '%s' in Organization:'%s' is not Analyse because it is empty\n", repo.Name, AppConfig.Organization)
					EmptyR = EmptyR + 1
				}
			}

			NumberRepos1 := int(uintptr(len(repositories))) - EmptyR
			fmt.Printf("\nAnalyse '%s' Repositories('%d') in Organization: '%s'\n", AppConfig.DevOps, NumberRepos1, AppConfig.Organization)

			NumberRepos = AnalyseReposList(DestinationResult, AppConfig.Users, AppConfig.AccessToken, AppConfig.DevOps, AppConfig.Organization, repoList)

		case "bitbucket_dc":

			projects, err := getbibucketdc.GetProjectBitbucketList(AppConfig.Url, AppConfig.Baseapi, AppConfig.Apiver, AppConfig.AccessToken)
			if err != nil {
				fmt.Printf("❌ Error Get Info Projects in Bitbucket server '%s' : ", err)
				return
			}

			//	var projectList []Project1
			for _, project := range projects {
				fmt.Printf("Projet: %s, Repo: %s, Branche: %s, Taille: %d\n", project.ProjectKey, project.RepoSlug, project.MainBranch, project.LargestSize)

			}
			NumberProject1 := int(uintptr(len(projects)))
			fmt.Printf("\nNumber of Projects : '%d'\n", NumberProject1)

		}
	}

	spin := spinner.New(spinner.CharSets[35], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
	spin.Suffix = " \nAnalyse Report..."
	spin.Start()

	// List files in the directory
	fileInfos, err := os.ReadDir(DestinationResult)
	if err != nil {
		fmt.Println("Error listing files:", err)
		return
	}

	// Loop through each file
	for _, fileInfo := range fileInfos {
		if !fileInfo.IsDir() && filepath.Ext(fileInfo.Name()) == ".json" {
			filePath := filepath.Join(DestinationResult, fileInfo.Name())
			nameRepos := strings.Split(fileInfo.Name(), "_")
			nameRepos1 := strings.Split(nameRepos[1], ".")
			TotalCodeLines := parseJSONFile(filePath, nameRepos1[0])
			if TotalCodeLines > largestLineCounter {
				largestLineCounter = TotalCodeLines
				nameRepos2 = nameRepos1[0]
			}
		}
	}
	spin.Stop()

	p := message.NewPrinter(language.English)
	s := strings.Replace(p.Sprintf("%d", largestLineCounter), ",", " ", -1)

	message0 := fmt.Sprintf("\nNumber of Repository analyzed in Organization '%s' is '%d' \n", AppConfig.Organization, NumberRepos)
	message1 := fmt.Sprintf("In Organization '%s' the largest number of line of code is <'%s'> and the repository is <'%s'>\n\nReports are located in the Results directory", AppConfig.Organization, s, nameRepos2)
	message2 := message0 + message1
	fmt.Println(message0)
	fmt.Println(message1)

	// Write message in Gobal Report File
	_, err = file.WriteString(message2)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}
