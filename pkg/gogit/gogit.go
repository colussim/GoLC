package gogit

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/briandowns/spinner"
	git "github.com/go-git/go-git/v5"
)

func Getrepos(src string) (string, error) {
	spinner := newSpinner(fmt.Sprintf(" Extracting files from %s", src))
	spinner.Start()
	defer spinner.Stop()

	suffix, err := randomSuffix()
	if err != nil {
		return "", err
	}

	dst := filepath.Join(os.TempDir(), fmt.Sprintf("gcloc-extract-%s", suffix))
	//pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	_, err = git.PlainClone(dst, false, &git.CloneOptions{
		URL:      src,
		Progress: os.Stdout,
	})

	if err != nil {
		fmt.Println("\n-- Stack: gogit.Getrepos Git Clone -- :", err)
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

func newSpinner(text string) *spinner.Spinner {
	return spinner.New(
		spinner.CharSets[14],
		100*time.Millisecond,
		spinner.WithSuffix(text),
	)
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
