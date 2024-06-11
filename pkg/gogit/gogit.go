package gogit

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/capability"
	"github.com/go-git/go-git/v5/plumbing/transport"
	//"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func Getrepos(src, branch, token string) (string, error) {

	suffix, err := randomSuffix()
	if err != nil {
		return "", err
	}

	dst := filepath.Join(os.TempDir(), fmt.Sprintf("gcloc-extract-%s", suffix))
	//pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	log.SetOutput(os.Stderr)

	transport.UnsupportedCapabilities = []capability.Capability{
		capability.ThinPack,
	}

	_, err = git.PlainClone(dst, false, &git.CloneOptions{
		/*Auth: &http.BasicAuth{
			Username: "x-token-auth",
			Password: """,
		},*/
		URL: src,

		ReferenceName: plumbing.NewBranchReferenceName(branch),
		//ReferenceName: plumbing.ReferenceName(branch),

		SingleBranch: true,
		Depth:        1,
	})

	if err != nil {
		fmt.Printf("\n--❌ Stack: gogit.Getrepos Git Branch %s - %s-- Source: %s -", plumbing.Main, err, src)
	}

	symLink, err := isSymLink(dst)
	if err != nil {
		return "", err
	}

	if symLink {
		origin, err := os.Readlink(dst)
		if err != nil {
			return "", err
		}

		return origin, nil
	}

	return dst, nil
}

func randomSuffix() (string, error) {
	randBytes := make([]byte, 16)
	_, err := rand.Read(randBytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(randBytes), nil
}

func isSymLink(path string) (bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return false, err
	}

	return info.Mode()&os.ModeSymlink != 0, nil
}
