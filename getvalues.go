package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Result struct {
	TotalFiles      int           `json:"TotalFiles"`
	TotalLines      int           `json:"TotalLines"`
	TotalBlankLines int           `json:"TotalBlankLines"`
	TotalComments   int           `json:"TotalComments"`
	TotalCodeLines  int           `json:"TotalCodeLines"`
	Results         []LanguageRes `json:"Results"`
}

type LanguageRes struct {
	Language   string `json:"Language"`
	Files      int    `json:"Files"`
	Lines      int    `json:"Lines"`
	BlankLines int    `json:"BlankLines"`
	Comments   int    `json:"Comments"`
	CodeLines  int    `json:"CodeLines"`
}

func formatCodeLines(numLines float64) string {
	if numLines >= 1000000 {
		return fmt.Sprintf("%.2fM", numLines/1000000)
	} else if numLines >= 1000 {
		return fmt.Sprintf("%.2fK", numLines/1000)
	} else {
		return fmt.Sprintf("%.0f", numLines)
	}
}

func main() {
	// Chemin du répertoire contenant les fichiers JSON
	//resultDir := "/chemin/vers/le/repertoire/Result"
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err)
	}
	DestinationResult := pwd + "/Results"

	// Variables pour garder une trace du projet et du repo avec le TotalCodeLines le plus élevé
	var maxTotalCodeLines int
	var maxProject, maxRepo string

	// List files in the directory
	files, err := os.ReadDir(DestinationResult)
	if err != nil {
		fmt.Println("❌ Error listing files:", err)
		os.Exit(1)
	}

	// Initialiser la somme de TotalCodeLines
	totalCodeLinesSum := 0

	for _, file := range files {
		// Vérifier si le fichier est un fichier JSON
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			// Lire le contenu du fichier JSON
			filePath := filepath.Join(DestinationResult, file.Name())
			jsonData, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Printf("Erreur lors de la lecture du fichier %s: %v\n", file.Name(), err)
				continue
			}

			// Analyser le contenu JSON dans une structure Result
			var result Result
			err = json.Unmarshal(jsonData, &result)
			if err != nil {
				fmt.Printf("Erreur lors de l'analyse du contenu JSON du fichier %s: %v\n", file.Name(), err)
				continue
			}

			// Ajouter TotalCodeLines à la somme totale
			totalCodeLinesSum += result.TotalCodeLines

			// Vérifier si ce repo a un TotalCodeLines plus élevé que le maximum actuel
			if result.TotalCodeLines > maxTotalCodeLines {
				maxTotalCodeLines = result.TotalCodeLines
				// Extraire le nom du projet et du repo du nom du fichier
				parts := strings.Split(strings.TrimSuffix(file.Name(), ".json"), "_")
				maxProject = parts[1]
				maxRepo = parts[2]
			}
		}
	}

	maxTotalCodeLines1 := formatCodeLines(float64(maxTotalCodeLines))
	totalCodeLinesSum1 := formatCodeLines(float64(totalCodeLinesSum))

	// Afficher le TotalCodeLines maximum avec le nom du projet et du repo correspondants
	fmt.Printf("Le repo avec le TotalCodeLines le plus élevé est dans le projet %s, repo %s, avec un TotalCodeLines de %s\n", maxProject, maxRepo, maxTotalCodeLines1)

	// Afficher la somme totale de TotalCodeLines
	fmt.Printf("La somme totale de TotalCodeLines de tous les fichiers est de %s\n", totalCodeLinesSum1)
}
