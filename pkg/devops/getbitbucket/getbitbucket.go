package getbibucket

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
)

func GetProjectBitbucketListCloud(url, baseapi, apiver, accessToken, exlusionfile, project, repo string) ([]ProjectBranch, error) {

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

	}
}
