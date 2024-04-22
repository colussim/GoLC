package utils

import (
	"encoding/json"
	"os"
)

type ExclusionList struct {
	Projects []string `json:"projects"`
	Repos    []string `json:"repos"`
}

func loadExclusionListBit(filePath string) (ExclusionList, error) {
	var exclusionList ExclusionList
	file, err := os.Open(filePath)
	if err != nil {
		return exclusionList, err
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&exclusionList)
	if err != nil {
		return exclusionList, err
	}

	return exclusionList, nil
}
