package getbitbucketdc

import (
    "context"
    "fmt"
    "os"
    "sort"
    "strings"
    "time"

    bitbucket "github.com/gfleury/go-bitbucket-v1"
    "github.com/briandowns/spinner"
    "github.com/colussim/GoLC/pkg/utils"
)

type ProjectBranch struct {
    ProjectKey  string
    RepoSlug    string
    MainBranch  string
    LargestSize int64
}

type AnalysisResult struct {
    NumProjects     int
    NumRepositories int
    ProjectBranches []ProjectBranch
}

func newBitbucketClient(baseURL, token string) *bitbucket.APIClient {
    config := bitbucket.NewConfiguration(baseURL)
    config.AddDefaultHeader("Authorization", "Bearer "+token)
    return bitbucket.NewAPIClient(config)
}

// Recursively walk the file tree for a branch and sum file sizes
func getBranchSizeRecursive(ctx context.Context, client *bitbucket.APIClient, projectKey, repoSlug, branchID string) (int64, error) {
    return walkDirectory(ctx, client, projectKey, repoSlug, branchID, "")
}

func walkDirectory(ctx context.Context, client *bitbucket.APIClient, projectKey, repoSlug, branchID, path string) (int64, error) {
    var total int64
    opts := &bitbucket.APIGetBrowseOpts{
        At:    bitbucket.OptionalString("refs/heads/" + branchID),
        Path:  bitbucket.OptionalString(path),
        Limit: bitbucket.OptionalInt32(100),
    }
    resp, _, err := client.BrowseApi.GetBrowse(ctx, projectKey, repoSlug, opts)
    if err != nil {
        return 0, err
    }
    for _, child := range resp.Children.Values {
        switch child.Type {
        case "FILE":
            total += int64(child.Size)
        case "DIRECTORY":
            subdir := child.Path.ToString
            subTotal, err := walkDirectory(ctx, client, projectKey, repoSlug, branchID, subdir)
            if err != nil {
                return total, err
            }
            total += subTotal
        }
    }
    return total, nil
}

// Fetch all projects, respecting exclusion list
func fetchAllProjects(ctx context.Context, client *bitbucket.APIClient, exclusionList *utils.ExclusionList) ([]bitbucket.Project, error) {
    var allProjects []bitbucket.Project
    start := int32(0)
    for {
        opts := &bitbucket.APIGetProjectsOpts{
            Limit: bitbucket.OptionalInt32(100),
            Start: bitbucket.OptionalInt32(start),
        }
        resp, _, err := client.ProjectsApi.GetProjects(ctx, opts)
        if err != nil {
            return nil, err
        }
        for _, p := range resp.Values {
            if exclusionList != nil && exclusionList.Projects[p.Key] {
                continue
            }
            allProjects = append(allProjects, p)
        }
        if resp.IsLastPage {
            break
        }
        start = resp.NextPageStart
    }
    return allProjects, nil
}

// Fetch all repos for a project, respecting exclusion list
func fetchAllRepos(ctx context.Context, client *bitbucket.APIClient, projectKey string, exclusionList *utils.ExclusionList) ([]bitbucket.Repository, error) {
    var allRepos []bitbucket.Repository
    start := int32(0)
    for {
        opts := &bitbucket.APIGetRepositoriesOpts{
            Limit: bitbucket.OptionalInt32(100),
            Start: bitbucket.OptionalInt32(start),
        }
        resp, _, err := client.RepositoriesApi.GetRepositories(ctx, projectKey, opts)
        if err != nil {
            return nil, err
        }
        for _, r := range resp.Values {
            key := r.Project.Key + "/" + r.Slug
            if exclusionList != nil && exclusionList.Repos[key] {
                continue
            }
            allRepos = append(allRepos, r)
        }
        if resp.IsLastPage {
            break
        }
        start = resp.NextPageStart
    }
    return allRepos, nil
}

// Fetch all branches for a repo
func fetchAllBranches(ctx context.Context, client *bitbucket.APIClient, projectKey, repoSlug string) ([]bitbucket.Branch, error) {
    var allBranches []bitbucket.Branch
    start := int32(0)
    for {
        opts := &bitbucket.APIGetBranchesOpts{
            Limit: bitbucket.OptionalInt32(100),
            Start: bitbucket.OptionalInt32(start),
        }
        resp, _, err := client.BranchesApi.GetBranches(ctx, projectKey, repoSlug, opts)
        if err != nil {
            return nil, err
        }
        allBranches = append(allBranches, resp.Values...)
        if resp.IsLastPage {
            break
        }
        start = resp.NextPageStart
    }
    return allBranches, nil
}

// Find the largest branch by file size
func findLargestBranch(ctx context.Context, client *bitbucket.APIClient, projectKey, repoSlug string, branches []bitbucket.Branch, spin *spinner.Spinner) (string, int64, error) {
    type branchSize struct {
        Name string
        Size int64
    }
    var sizes []branchSize
    for _, branch := range branches {
        spin.Prefix = fmt.Sprintf("\t   Analysis branch <%s> size...", branch.DisplayID)
        spin.Start()
        size, err := getBranchSizeRecursive(ctx, client, projectKey, repoSlug, branch.ID)
        spin.Stop()
        if err != nil {
            continue
        }
        sizes = append(sizes, branchSize{Name: branch.DisplayID, Size: size})
    }
    sort.Slice(sizes, func(i, j int) bool { return sizes[i].Size > sizes[j].Size })
    if len(sizes) > 0 {
        return sizes[0].Name, sizes[0].Size, nil
    }
    return "", 0, nil
}

// Main analysis entry point
func GetProjectBitbucketList(platformConfig map[string]interface{}, exclusionFile string) ([]ProjectBranch, error) {
    loggers := utils.NewLogger()
    bitbucketURLBase := platformConfig["Url"].(string)
    token := platformConfig["AccessToken"].(string)
    exclusionList, err := loadOrCreateExclusionList(exclusionFile)
    if err != nil {
        loggers.Errorf("\n❌ Error Reading Exclusion File <%s>: %v\n", exclusionFile, err)
        return nil, err
    }
    client := newBitbucketClient(bitbucketURLBase, token)
    ctx := context.Background()
    spin := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
    spin.Prefix = "Get Projects... "
    spin.Color("green", "bold")

    projects, err := fetchAllProjects(ctx, client, exclusionList)
    if err != nil {
        return nil, err
    }

    var importantBranches []ProjectBranch
    var nbRepos int

    for _, project := range projects {
        loggers.Infof("\t🟢  Analyse Projet: %s ", project.Name)
        repos, err := fetchAllRepos(ctx, client, project.Key, exclusionList)
        if err != nil {
            loggers.Errorf("\r❌ Get Repos for each Project:%v", err)
            continue
        }
        nbRepos += len(repos)
        loggers.Infof("\t  ✅ The number of Repo(s) found is: %d", len(repos))
        for _, repo := range repos {
            branches, err := fetchAllBranches(ctx, client, project.Key, repo.Slug)
            if err != nil || len(branches) == 0 {
                loggers.Warningf("❗️ No branches found for repository %s\n", repo.Slug)
                continue
            }
            loggers.Infof("\t   ✅ Repo: <%s> - Number of branches: %d", repo.Name, len(branches))
            largestBranch, largestSize, err := findLargestBranch(ctx, client, project.Key, repo.Slug, branches, spin)
            if err != nil {
                loggers.Errorf("❌ Error processing repo %s: %v\n", repo.Name, err)
                continue
            }
            loggers.Infof("\t     ✅ The largest branch of the repo is <%s> of size : %s", largestBranch, utils.FormatSize(largestSize))
            importantBranches = append(importantBranches, ProjectBranch{
                ProjectKey:  project.Key,
                RepoSlug:    repo.Slug,
                MainBranch:  largestBranch,
                LargestSize: largestSize,
            })
        }
    }

    result := AnalysisResult{
        NumProjects:     len(projects),
        NumRepositories: nbRepos,
        ProjectBranches: importantBranches,
    }
    if err := saveAnalysisResult1("Results/config/analysis_repos.json", result); err != nil {
        loggers.Errorf("❌ Error creating Analysis file:%v", err)
        return importantBranches, err
    }
    return importantBranches, nil
}

// Save analysis result as JSON
func saveAnalysisResult1(filePath string, result AnalysisResult) error {
    file, err := os.Create(filePath)
    if err != nil {
        return err
    }
    defer file.Close()
    return utils.JSONEncode(file, result)
}

// Load or create exclusion list
func loadOrCreateExclusionList(exclusionFile string) (*utils.ExclusionList, error) {
    if exclusionFile == "0" {
        return &utils.ExclusionList{
            Projects: make(map[string]bool),
            Repos:    make(map[string]bool),
        }, nil
    }
    return utils.LoadExclusionList(exclusionFile)
}
