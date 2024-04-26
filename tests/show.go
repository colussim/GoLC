package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
)

type LanguageData struct {
	Language   string  `json:"Language"`
	CodeLines  int     `json:"CodeLines"`
	Percentage float64 `json:"Percentage"`
}

func main() {
	// Définition du gestionnaire pour la racine "/"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Lecture des données depuis le fichier JSON
		data, err := os.ReadFile("resultat.json")
		if err != nil {
			http.Error(w, "Erreur lors de la lecture des données", http.StatusInternalServerError)
			return
		}

		// Décodage des données JSON
		var languages []LanguageData
		err = json.Unmarshal(data, &languages)
		if err != nil {
			http.Error(w, "Erreur lors du décodage des données JSON", http.StatusInternalServerError)
			return
		}

		// Calcul des pourcentages
		total := 0
		for _, lang := range languages {
			total += lang.CodeLines
		}
		for i := range languages {
			languages[i].Percentage = float64(languages[i].CodeLines) / float64(total) * 100
		}

		// Chargement du template HTML
		tmpl := template.Must(template.New("index").Parse(htmlTemplate))

		// Exécution du template avec les données
		err = tmpl.Execute(w, languages)
		if err != nil {
			http.Error(w, "Erreur lors de l'exécution du template HTML", http.StatusInternalServerError)
			return
		}
	})

	// Démarrage du serveur HTTP sur le port 8080
	fmt.Println("Serveur démarré sur http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

// Le template HTML
const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Diagramme Circulaire des lignes de code par langage</title>
    <style>
        .container {
            display: flex;
            align-items: center;
        }
        .chart-container {
            flex: 1;
        }
        .percentage-container {
            flex: 1;
            padding-left: 20px;
        }
    </style>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
</head>
<body>
    <div class="container">
        <div class="chart-container">
            <canvas id="camembertChart" width="400" height="400"></canvas>
        </div>
        <div class="percentage-container">
            <ul>
                {{range .}}
                <li>{{.Language}}: {{printf "%.1f" .Percentage}}%</li>
                {{end}}
            </ul>
        </div>
    </div>
    <script>
        var ctx = document.getElementById('camembertChart').getContext('2d');
        var camembertChart = new Chart(ctx, {
            type: 'doughnut',
            data: {
                labels: [{{range .}}"{{.Language}}",{{end}}],
                datasets: [{
                    label: 'Lignes de code par langage',
                    data: [{{range .}}{{.CodeLines}},{{end}}],
                    backgroundColor: [
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
                        'rgba(255, 159, 64, 1)'
                    ],
                    borderWidth: 1
                }]
            },
            options: {
                responsive: false,
                legend: {
                    display: false
                }
            }
        });
    </script>
</body>
</html>

`
