package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type FileData struct {
	Results []LanguageData `json:"Results"`
}

type LanguageData struct {
	Language  string `json:"Language"`
	CodeLines int    `json:"CodeLines"`
}

func main() {
	// Chemin du répertoire contenant les fichiers JSON
	directory := "../../Results" // Remplacez ceci par le chemin de votre répertoire

	// Dictionnaire pour stocker le nombre de lignes de code par langage
	ligneDeCodeParLangage := make(map[string]int)

	// Parcours de tous les fichiers dans le répertoire
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".json" {
			// Lecture du fichier JSON
			fileData, err := ioutil.ReadFile(path)
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
		return nil
	})
	if err != nil {
		fmt.Println("Erreur lors de la lecture des fichiers :", err)
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
		fmt.Println("Erreur lors de la création du fichier JSON de sortie :", err)
		return
	}

	err = ioutil.WriteFile("resultat.json", outputData, 0644)
	if err != nil {
		fmt.Println("Erreur lors de l'écriture dans le fichier JSON de sortie :", err)
		return
	}

	fmt.Println("Les résultats ont été enregistrés dans resultat.json")
}
