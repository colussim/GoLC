package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Dictionnaire pour stocker le nombre de lignes de code par langage
	linesByLanguage := make(map[string]int)

	// Parcours des fichiers dans le répertoire Results
	resultsDir := "../Results"
	err := filepath.Walk(resultsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Si le fichier n'est pas un répertoire et son nom commence par "Result_"
		if !info.IsDir() && strings.HasPrefix(info.Name(), "Result_") {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			var data map[string]interface{}
			err = json.NewDecoder(file).Decode(&data)
			if err != nil {
				return err
			}

			results, ok := data["Results"].([]interface{})
			if !ok {
				return fmt.Errorf("invalid format for Results")
			}

			for _, result := range results {
				resultMap, ok := result.(map[string]interface{})
				if !ok {
					return fmt.Errorf("invalid format for result")
				}

				language, ok := resultMap["Language"].(string)
				if !ok {
					return fmt.Errorf("invalid format for language")
				}

				codeLines, ok := resultMap["CodeLines"].(float64)
				if !ok {
					return fmt.Errorf("invalid format for codeLines")
				}

				linesByLanguage[language] += int(codeLines)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Println("Erreur lors du parcours des fichiers:", err)
		return
	}

	// Écriture des résultats dans un fichier JSON
	outputFile := "code_lines_by_language.json"
	outputFilePtr, err := os.Create(outputFile)
	if err != nil {
		fmt.Println("Erreur lors de la création du fichier de sortie:", err)
		return
	}
	defer outputFilePtr.Close()

	err = json.NewEncoder(outputFilePtr).Encode(linesByLanguage)
	if err != nil {
		fmt.Println("Erreur lors de l'écriture des résultats:", err)
		return
	}

	fmt.Println("Résultats enregistrés dans", outputFile)
}

