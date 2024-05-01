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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Reading data from the JSON file
		data, err := os.ReadFile("resultat.json")
		if err != nil {
			http.Error(w, "❌ Error reading data", http.StatusInternalServerError)
			return
		}

		// JSON data decoding
		var languages []LanguageData
		err = json.Unmarshal(data, &languages)
		if err != nil {
			http.Error(w, "❌ Error decoding JSON data", http.StatusInternalServerError)
			return
		}

		// Calculating percentages
		total := 0
		for _, lang := range languages {
			total += lang.CodeLines
		}
		for i := range languages {
			languages[i].Percentage = float64(languages[i].CodeLines) / float64(total) * 100
		}

		// Load HTML template
		tmpl := template.Must(template.New("index").Parse(htmlTemplate))

		// Run Template
		err = tmpl.Execute(w, languages)
		if err != nil {
			http.Error(w, "❌ Error executing HTML template", http.StatusInternalServerError)
			return
		}
	})

	// Démarrage du serveur HTTP sur le port 8080
	fmt.Println("✅ Server started on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

// Le template HTML
const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Result Go LOC</title>
    <link href="../dist/bootstrap/5.3.3/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="preconnect" href="https://fonts.gstatic.com">
    <link href="https://fonts.googleapis.com/css2?family=Manrope:wght@200;300;400;500;600;700&amp;display=swap" rel="stylesheet">
   
    <link href="../dist/css/theme.css" rel="stylesheet" />
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
<main class="main" id="top">
      <nav class="navbar navbar-expand-lg fixed-top navbar-dark" data-navbar-on-scroll="data-navbar-on-scroll">
        <div class="container"><a class="navbar-brand" href="index.html"><img src="../dist/img/Logo.png" alt="" /></a>
          <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarSupportedContent" aria-controls="navbarSupportedContent" aria-expanded="false" aria-label="Toggle navigation"><i class="fa-solid fa-bars text-white fs-3"></i></button>
          <div class="collapse navbar-collapse" id="navbarSupportedContent">
            <ul class="navbar-nav ms-auto mt-2 mt-lg-0">
            </ul>
          </div>
        </div>
      </nav>
      <div class="bg-dark"><img class="img-fluid position-absolute end-0" src="assets/img/hero/hero-bg.png" alt="" />

    <div class="container">
        <div class="chart-container">
            <canvas id="camembertChart" width="400" height="400"></canvas>
        </div>
        <div class="percentage-container">
            <ul>
                {{range .}}
                <li>{{.Language}}: {{printf "%.2f" .Percentage}}%</li>
                {{end}}
            </ul>
        </div>
    </div>

    </div>
    <script>

    function formatTooltipLabel(tooltipItem, data) {
        var label = data.labels[tooltipItem.index] || '';
        var value = data.datasets[tooltipItem.datasetIndex].data[tooltipItem.index];
        var unit = "";
        if (value >= 1000000) {
            unit = "M";
            value = value / 1000000;
        } else if (value >= 1000) {
            unit = "K";
            value = value / 1000;
        }
        return label + ': ' + value.toFixed(2) + unit;
    }
    

        var ctx = document.getElementById('camembertChart').getContext('2d');
        var camembertChart = new Chart(ctx, {
            type: 'doughnut',
            data: {
               labels: [{{range .}}"{{.Language}}",{{end}}],
            
                datasets: [{
                    label: 'LOC by language',
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
                },
                tooltips: {
                    callbacks: {
                        label: function(tooltipItem, data) {
                            return formatTooltipLabel(tooltipItem, data);
                        }
                    }
                }
            }

        });
    </script>
</body>
</html>

`
