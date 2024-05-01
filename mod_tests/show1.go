package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
)

type GlobalReport struct {
	Organization           string `json:"Organization"`
	TotalLinesOfCode       string `json:"TotalLinesOfCode"`
	LargestRepository      string `json:"LargestRepository"`
	LinesOfCodeLargestRepo string `json:"LinesOfCodeLargestRepo"`
}

type LanguagePercentage struct {
	Language   string
	Percentage float64
}

func main() {
	// Chargement des données du fichier GlobalReport.json
	globalReportFile, err := os.Open("GlobalReport.json")
	if err != nil {
		panic(err)
	}
	defer globalReportFile.Close()

	var globalReport GlobalReport
	err = json.NewDecoder(globalReportFile).Decode(&globalReport)
	if err != nil {
		panic(err)
	}

	// Chargement des données du fichier code_lines_by_language.json
	codeLinesByLanguageFile, err := os.Open("code_lines_by_language.json")
	if err != nil {
		panic(err)
	}
	defer codeLinesByLanguageFile.Close()

	var codeLinesByLanguage map[string]int
	err = json.NewDecoder(codeLinesByLanguageFile).Decode(&codeLinesByLanguage)
	if err != nil {
		panic(err)
	}

	// Calcul des pourcentages des langages
	var languagePercentages []LanguagePercentage
	totalLinesOfCode, _ := strconv.Atoi(globalReport.TotalLinesOfCode)
	if totalLinesOfCode > 0 {
		for language, lines := range codeLinesByLanguage {
			percentage := float64(lines) / float64(totalLinesOfCode) * 100
			languagePercentages = append(languagePercentages, LanguagePercentage{Language: language, Percentage: percentage})
		}
	} else {
		// Si totalLinesOfCode est égal à zéro, définir tous les pourcentages à zéro
		for language := range codeLinesByLanguage {
			languagePercentages = append(languagePercentages, LanguagePercentage{Language: language, Percentage: 0})
		}
	}

	// Création de la page HTML en utilisant le modèle
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			GlobalReport        GlobalReport
			LanguagePercentages []LanguagePercentage
		}{
			GlobalReport:        globalReport,
			LanguagePercentages: languagePercentages,
		}

		tmpl := template.Must(template.New("dashboard").Funcs(template.FuncMap{
			"extractLanguages": func(lp []LanguagePercentage) []string {
				var languages []string
				for _, l := range lp {
					languages = append(languages, l.Language)
				}
				return languages
			},
			"extractPercentages": func(lp []LanguagePercentage) []float64 {
				var percentages []float64
				for _, l := range lp {
					percentages = append(percentages, l.Percentage)
				}
				return percentages
			},
			"jsonify": func(data interface{}) (template.JS, error) {
				jsonData, err := json.Marshal(data)
				if err != nil {
					return "", err
				}
				return template.JS(jsonData), nil
			},
		}).Parse(dashboardTemplate))
		err := tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Démarrage du serveur HTTP
	fmt.Println("Server running on port 8080")
	http.ListenAndServe(":8080", nil)
}

const dashboardTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Dashboard</title>
  <link href="https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css" rel="stylesheet">
</head>
<body>
  <div class="container">
    <header class="mt-5 mb-4">
      <h1>Dashboard</h1>
      <div class="row">
        <div class="col">
          <h3>Organization: {{ .GlobalReport.Organization }}</h3>
          <h3>Total Lines of Code: {{ .GlobalReport.TotalLinesOfCode }}</h3>
          <h3>Largest Repository: {{ .GlobalReport.LargestRepository }}</h3>
          <h3>Lines of Code in Largest Repo: {{ .GlobalReport.LinesOfCodeLargestRepo }}</h3>
        </div>
      </div>
    </header>
    <div class="row">
      <div class="col">
        <h2>Languages Percentage</h2>
        <div class="row">
          <div class="col-md-6">
            <canvas id="languageChart"></canvas>
          </div>
          <div class="col-md-6">
            <ul class="list-group">
              {{ range extractLanguages .LanguagePercentages }}
              <li class="list-group-item d-flex justify-content-between align-items-center">
                {{ . }}
              </li>
              {{ end }}
            </ul>
          </div>
        </div>
      </div>
    </div>
  </div>

  <script src="https://cdnjs.cloudflare.com/ajax/libs/Chart.js/2.9.3/Chart.min.js"></script>
  <script>
    var ctx = document.getElementById('languageChart').getContext('2d');
    var data = {
      labels: {{ jsonify (extractLanguages .LanguagePercentages) }},
      datasets: [{
        data: {{ jsonify (extractPercentages .LanguagePercentages) }},
        backgroundColor: [
          'rgba(255, 99, 132, 0.6)',
          'rgba(54, 162, 235, 0.6)',
          'rgba(255, 206, 86, 0.6)',
          'rgba(75, 192, 192, 0.6)',
          'rgba(153, 102, 255, 0.6)',
          'rgba(255, 159, 64, 0.6)'
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
    };
    var options = {
      responsive: true,
      maintainAspectRatio: false
    };
    var myPieChart = new Chart(ctx, {
      type: 'pie',
      data: data,
      options: options
    });
  </script>
</body>
</html>


`
