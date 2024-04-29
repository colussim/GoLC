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

	directory := "Results" // Remplacez ceci par le chemin de votre répertoire

	// Dictionnaire pour stocker le nombre de lignes de code par langage
	ligneDeCodeParLangage := make(map[string]int)

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
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

			if filepath.Ext(path) == ".json" {
				// Lecture du fichier JSON
				fileData, err := os.ReadFile(path)
				if err != nil {
					return err
				}

				// Décodage des données JSON
				var data FileData
				err = json.Unmarshal(fileData, &data)
				if err != nil {
					return err
				}

				// Parcours des résultats de chaque fichier
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
		fmt.Println("❌ Erreur lors de la lecture des fichiers :", err)
		return
	}

	// Création de la structure de sortie
	var resultats []LanguageData
	for lang, total := range ligneDeCodeParLangage {
		resultats = append(resultats, LanguageData{
			Language:  lang,
			CodeLines: total,
		})
	}

	// Écriture des résultats dans un fichier JSON
	outputData, err := json.MarshalIndent(resultats, "", "  ")
	if err != nil {
		fmt.Println("❌ Erreur lors de la création du fichier JSON de sortie :", err)
		return
	}
	outputFile := "Results/code_lines_by_language.json"
	err = os.WriteFile(outputFile, outputData, 0644)
	if err != nil {
		fmt.Println("❌ Erreur lors de l'écriture dans le fichier JSON de sortie :", err)
		return
	}

	fmt.Println("✅ Results recorded in", outputFile)
	fmt.Println("✅ Next step : run Result")
}
