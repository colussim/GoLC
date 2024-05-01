package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sort"
	"strconv"
)

type Language struct {
	Language   string
	CodeLines  int
	Percentage float64
}

type GlobalReport struct {
	Organization           string `json:"Organization"`
	TotalLinesOfCode       string `json:"TotalLinesOfCode"`
	LargestRepository      string `json:"LargestRepository"`
	LinesOfCodeLargestRepo string `json:"LinesOfCodeLargestRepo"`
}

func convertToInteger(str string) (int, error) {
	// Décomposer la chaîne en nombre et unité
	var numStr string
	var unit string
	_, err := fmt.Sscanf(str, "%s%s", &numStr, &unit)
	if err != nil {
		return 0, err
	}

	// Convertir la chaîne en nombre entier
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, err
	}

	// Convertir l'unité en facteur multiplicatif
	var multiplier float64
	switch unit {
	case "K":
		multiplier = 1000
	case "M":
		multiplier = 1000000
	default:
		return 0, fmt.Errorf("unité invalide: %s", unit)
	}

	// Multiplier le nombre par le facteur multiplicatif
	result := int(num * multiplier)
	return result, nil
}

func main() {
	// Ouvrir et lire le fichier GlobalReport.json
	globalReportFile, err := os.Open("GlobalReport.json")
	if err != nil {
		fmt.Println("Erreur lors de l'ouverture du fichier GlobalReport.json:", err)
		return
	}
	defer globalReportFile.Close()

	var globalReport GlobalReport
	err = json.NewDecoder(globalReportFile).Decode(&globalReport)
	if err != nil {
		fmt.Println("Erreur lors de la lecture du fichier GlobalReport.json:", err)
		return
	}

	// Convertir TotalLinesOfCode en nombre entier
	totalLines, err := convertToInteger(globalReport.TotalLinesOfCode)
	if err != nil {
		fmt.Println("Erreur lors de la conversion du nombre total de lignes de code:", err)
		return
	}

	// Ouvrir et lire le fichier resultat.json
	resultatFile, err := os.Open("resultat.json")
	if err != nil {
		fmt.Println("Erreur lors de l'ouverture du fichier resultat.json:", err)
		return
	}
	defer resultatFile.Close()

	var languages []Language
	err = json.NewDecoder(resultatFile).Decode(&languages)
	if err != nil {
		fmt.Println("Erreur lors de la lecture du fichier resultat.json:", err)
		return
	}

	// Calculer les pourcentages de lignes de code pour chaque langage
	for i := range languages {
		languages[i].Percentage = float64(languages[i].CodeLines) / float64(totalLines) * 100
	}

	// Trier les langages par pourcentage décroissant
	sort.Slice(languages, func(i, j int) bool {
		return languages[i].Percentage > languages[j].Percentage
	})

	// Charger le template HTML
	tmpl := template.Must(template.New("index").Parse(htmlTemplate))

	// Créer un gestionnaire de requêtes HTTP pour servir la page HTML
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Exécuter le template avec les données et écrire la réponse dans la réponse HTTP
		err := tmpl.Execute(w, languages)
		if err != nil {
			http.Error(w, "Erreur lors de l'exécution du template HTML", http.StatusInternalServerError)
			return
		}
	})

	// Démarrer le serveur HTTP sur le port 8080
	fmt.Println("Serveur démarré sur le port :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Erreur lors du démarrage du serveur:", err)
	}
}

// Modèle HTML pour la page de résultat avec un diagramme circulaire
const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Diagramme circulaire des langages utilisés</title>
<script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
</head>
<body>
<h1>Diagramme circulaire des langages utilisés</h1>
<div style="width: 50%; float: left;">
    <canvas id="leftChart"></canvas>
</div>
<div style="width: 50%; float: right;">
    <canvas id="rightChart"></canvas>
</div>
<script>
    // Données pour le diagramme circulaire des langages utilisés
    var labels = {{printf "%q" . |html}};
    var percentages = {{printf "%v" . |html}};
    
    // Configuration du diagramme circulaire
    var leftConfig = {
        type: 'doughnut',
        data: {
            labels: labels,
            datasets: [{
                label: 'Pourcentage des langages utilisés',
                data: percentages,
                backgroundColor: [
                    'rgba(255, 99, 132, 0.5)',
                    'rgba(54, 162, 235, 0.5)',
                    'rgba(255, 206, 86, 0.5)',
                    'rgba(75, 192, 192, 0.5)',
                    'rgba(153, 102, 255, 0.5)',
                    'rgba(255, 159, 64, 0.5)',
                    'rgba(255, 99, 132, 0.5)',
                    'rgba(54, 162, 235, 0.5)',
                    'rgba(255, 206, 86, 0.5)',
                    'rgba(75, 192, 192, 0.5)',
                    'rgba(153, 102, 255, 0.5)',
                    'rgba(255, 159, 64, 0.5)'
                ],
                borderColor: [
                    'rgba(255, 99, 132, 1)',
                    'rgba(54, 162, 235, 1)',
                    'rgba(255, 206, 86, 1)',
                    'rgba(75, 192, 192, 1)',
                    'rgba(153, 102, 255, 1)',
                    'rgba(255, 159, 64, 1)',
                    'rgba(255, 99, 132, 1)',
                    'rgba(54, 162, 235, 1)',
                    'rgba(255, 206, 86, 1)',
                    'rgba(75, 192, 192, 1)',
                    'rgba(153, 102, 255, 1)',
                    'rgba(255, 159, 64, 1)'
                ],
                borderWidth: 1
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            title: {
                display: true,
                text: 'Pourcentage des langages utilisés'
            }
        }
    };

    var rightConfig = {
        type: 'doughnut',
        data: {
            labels: labels,
            datasets: [{
                label: 'Pourcentage des langages utilisés',
                data: percentages,
                backgroundColor: [
                    'rgba(255, 99, 132, 0.5)',
                    'rgba(54, 162, 235, 0.5)',
                    'rgba(255, 206, 86, 0.5)',
                    'rgba(75, 192, 192, 0.5)',
                    'rgba(153, 102, 255, 0.5)',
                    'rgba(255, 159, 64, 0.5)',
                    'rgba(255, 99, 132, 0.5)',
                    'rgba(54, 162, 235, 0.5)',
                    'rgba(255, 206, 86, 0.5)',
                    'rgba(75, 192, 192, 0.5)',
                    'rgba(153, 102, 255, 0.5)',
                    'rgba(255, 159, 64, 0.5)'
                ],
                borderColor: [
                    'rgba(255, 99, 132, 1)',
                    'rgba(54, 162, 235, 1)',
                    'rgba(255, 206, 86, 1)',
                    'rgba(75, 192, 192, 1)',
                    'rgba(153, 102, 255, 1)',
                    'rgba(255, 159, 64, 1)',
                    'rgba(255, 99, 132, 1)',
                    'rgba(54, 162, 235, 1)',
                    'rgba(255, 206, 86, 1)',
                    'rgba(75, 192, 192, 1)',
                    'rgba(153, 102, 255, 1)',
                    'rgba(255, 159, 64, 1)'
                ],
                borderWidth: 1
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            title: {
                display: true,
                text: 'Pourcentage des langages utilisés'
            }
        }
    };

    // Création des diagrammes
    var leftChart = new Chart(document.getElementById('leftChart'), leftConfig);
    var rightChart = new Chart(document.getElementById('rightChart'), rightConfig);
</script>
</body>
</html>
`
