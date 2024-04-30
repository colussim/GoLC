package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FileData struct {
	Results []LanguageData `json:"Results"`
}

type LanguageData struct {
	Language  string `json:"Language"`
	CodeLines int    `json:"CodeLines"`
}

func main() {

	directory := "Results"

	ligneDeCodeParLangage := make(map[string]int)

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// If the file is not a directory and its name starts with "Result_", then
		if !info.IsDir() && strings.HasPrefix(info.Name(), "Result_") {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if filepath.Ext(path) == ".json" {
				// Reading the JSON file
				fileData, err := os.ReadFile(path)
				if err != nil {
					return err
				}

				// JSON data decoding
				var data FileData
				err = json.Unmarshal(fileData, &data)
				if err != nil {
					return err
				}

				// Browse results for each file
				for _, result := range data.Results {
					language := result.Language
					codeLines := result.CodeLines
					ligneDeCodeParLangage[language] += codeLines
				}
			}
		}
		return nil
	})
	if err != nil {
		fmt.Println("❌ Error reading files :", err)
		return
	}

	// Create output structure
	var resultats []LanguageData
	for lang, total := range ligneDeCodeParLangage {
		resultats = append(resultats, LanguageData{
			Language:  lang,
			CodeLines: total,
		})
	}
	// Writing results to a JSON file
	outputData, err := json.MarshalIndent(resultats, "", "  ")
	if err != nil {
		fmt.Println("❌ Error creating output JSON file :", err)
		return
	}
	outputFile := "Results/code_lines_by_language.json"
	err = os.WriteFile(outputFile, outputData, 0644)
	if err != nil {
		fmt.Println("❌ Error writing to output JSON file :", err)
		return
	}

	fmt.Println("✅ Results recorded in", outputFile)
	fmt.Println("✅ Next step : run Result")
}
