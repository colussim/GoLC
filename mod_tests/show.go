package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
)

type Globalinfo struct {
	Organization           string `json:"Organization"`
	TotalLinesOfCode       string `json:"TotalLinesOfCode"`
	LargestRepository      string `json:"LargestRepository"`
	LinesOfCodeLargestRepo string `json:"LinesOfCodeLargestRepo"`
	DevOpsPlatform         string `json:"DevOpsPlatform"`
}
type LanguageData struct {
	Language   string  `json:"Language"`
	CodeLines  int     `json:"CodeLines"`
	Percentage float64 `json:"Percentage"`
}

type PageData struct {
	Languages    []LanguageData
	GlobalReport Globalinfo
}

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		// Reading data from the GlobalReport JSON file
		data0, err := os.ReadFile("Results/GlobalReport.json")
		if err != nil {
			http.Error(w, "❌ Error reading data0", http.StatusInternalServerError)
			return
		}

		// JSON data decoding
		var Ginfo Globalinfo

		err = json.Unmarshal(data0, &Ginfo)
		if err != nil {
			http.Error(w, "❌ Error decoding JSON data0", http.StatusInternalServerError)
			return
		}

		fmt.Println("P", Ginfo.DevOpsPlatform)

		// Reading data from the Resultanalyse JSON file
		data, err := os.ReadFile("Results/code_lines_by_language.json")
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

		pageData := PageData{
			Languages:    languages,
			GlobalReport: Ginfo,
		}

		// Run Template
		err = tmpl.Execute(w, pageData)
		if err != nil {
			http.Error(w, "❌ Error executing HTML template", http.StatusInternalServerError)
			return
		}
	})

	// Start HTTP server
	fmt.Println("✅ Server started on http://localhost:8080")
	http.Handle("/dist/", http.StripPrefix("/dist/", http.FileServer(http.Dir("dist"))))

	http.ListenAndServe(":8080", nil)
}

// <div class="col-lg-6 text-center text-lg-end mt-3 mt-lg-0"><img class="img-fluid" src="dist/img/devops1.png" alt="" /></div>
// <div class="bg-dark"><img class="img-fluid position-absolute end-0" src="dist/img/devops1.png" alt="" />
// https://cdn.jsdelivr.net/npm/chart.js
// HTML template
const htmlTemplate = `
<!DOCTYPE html>
<html lang="en-US" dir="ltr">

  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Result Go LOC</title>

    <link href="https://fonts.googleapis.com/css2?family=Manrope:wght@200;300;400;500;600;700&amp;display=swap" rel="stylesheet">
    <link href="/dist/css/theme.min.css" rel="stylesheet" type="text/css" />
    <link href="/dist/vendors/fontawesome/css/all.min.css" rel="stylesheet" type="text/css" />
    
  </head>
    <style>
       
        .chart-container {
            flex: 1;
        }
        .percentage-container {
            flex: 1;
            padding-left: 20px;
        }
      
    </style>
    <script src="/dist/vendors/chartjs/chart.js"></script>
</head>
<body>
<main class="main" id="top">
      <nav class="navbar navbar-expand-lg fixed-top navbar-dark" data-navbar-on-scroll="data-navbar-on-scroll">
        <div class="container"><a class="navbar-brand" href="index.html"><img src="dist/img/Logo.png" alt="" /></a>
         <div class="collapse navbar-collapse" id="navbarSupportedContent">
            <ul class="navbar-nav ms-auto mt-2 mt-lg-0">
            </ul>
          </div>
        </div>
      </nav>
      <div class="bg-dark"><img class="img-fluid position-absolute end-0" src="dist/img/bg.png" alt="" />
  
     

    <section>

      <div class="container">
        <div class="row align-items-center py-lg-8 py-6" style="margin-top: -5%">
          <div class="col-lg-6 text-center text-lg-start">
            <h1 class="text-white fs-5 fs-xl-6">Results</h1>     
              <div class="card text-white bg-primary mb-4" style="max-width: 23rem;">
                <h5 class="card-header text-white" style="padding: 1rem 1rem;"> <i class="fas fa-chart-line"></i> Organization: {{.GlobalReport.Organization}}

                {{if eq .GlobalReport.DevOpsPlatform "bitbucket_dc"}}
                    <i class="fab fa-bitbucket"></i>
                {{else if eq .GlobalReport.DevOpsPlatform "bitbucket"}}
                    <i class="fab fa-bitbucket"></i>
                {{else if eq .GlobalReport.DevOpsPlatform "github"}}
                     <i class="fab fa-github"></i>
                {{else if eq .GlobalReport.DevOpsPlatform "gitlab"}}
                    <i class="fab fa-gitlab"></i>
                {{else if eq .GlobalReport.DevOpsPlatform "azure"}}
                    <i class="fab fa-microsoft"></i>
                {{else}}
                    <i class="fas fa-folder"></i>
                {{end}}

                </h5>

                 <div class="card-body" style="padding: 1rem 1rem;">
                   <p class="card-text"><i class="fas fa-code-branch"></i> Total lines Of code : {{.GlobalReport.TotalLinesOfCode}}</p>
                   <p class="card-text"><i class="fas fa-folder"></i> Largest Repository : {{.GlobalReport.LargestRepository}}</p>
                   <p class="card-text"><i class="fas fa-code-branch"></i> Lines of code largest Repository : {{.GlobalReport.LinesOfCodeLargestRepo}}</p>
                 </div>
               </div>
               <div class="chart-container">
                <canvas id="camembertChart" width="400" height="400" ></canvas>
               </div>
          </div>
          <div class="col-lg-6  mt-3 mt-lg-0">
            <div class="card text-white bg-primary mb-4" style="max-width: 18rem;">
                <h5 class="card-header text-white" style="padding: 1rem 1rem;"><i class="fas fa-code"></i> Languages</h5>
                <div class="card-body text-white" style="padding: 1rem 1rem;">
                    <ul>
                    {{range .Languages}}
                        <li>{{.Language}}: {{printf "%.2f" .Percentage}}%</li>
                    {{end}}
                    </ul>
                </div>    
            </div>
          </div>
          
         
        </div>
        <div class="swiper">
            
        </div>
     </div>
    </section>

 
</main>

    <script src="/dist/vendors/chartjs/chart.js"></script>
    <script> 

    function formatTooltipLabel(tooltipItem, data) {
        var label =tooltipItem || '';
        var value = data;
        
        var unit = "";
    
        if (value >= 1000000) {
            unit = "M";
            value = (value / 1000000).toFixed(2) + unit;
        } else if (value >= 1000) {
            unit = "K";
            value = (value / 1000).toFixed(2) + unit;
        }
    
        return label + ': ' + value;
    }
    
    function commarize(min) {
        min = min || 1e3;
        // Alter numbers larger than 1k
        if (this >= min) {
          var units = ["k", "M", "B", "T"];
      
          var order = Math.floor(Math.log(this) / Math.log(1000));
      
          var unitname = units[(order - 1)];
          var num = Math.floor(this / 1000 ** order);
      
          // output number remainder + unitname
          return num + unitname
        }
      
        // return formatted original number
        return this.toLocaleString()
      }
      
    
    

        var ctx = document.getElementById('camembertChart').getContext('2d');
        var camembertChart = new Chart(ctx, {
            type: 'doughnut',
            data: {
               labels: [{{range .Languages}}"{{.Language}}",{{end}}],
            
                datasets: [{
                    label: 'LOC ',
                    data: [{{range .Languages}}{{.CodeLines}},{{end}}],
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
                plugins: {
                    legend: {
                        labels: {
                            color: 'white' 
                        }
                    }, 
                    tooltip: {
                        callbacks: {
                            label: function(context) {
                              // let value1:=context.dataset.data[context.dataIndex] ;
                            //  alert(context.dataset.data[context.dataIndex]);
                              //  alert(context.dataset.data);
                                return formatTooltipLabel(context.dataset.label, context.dataset.data[context.dataIndex]);
                            
                            }
                             
                        }
                    }
                }
                
            }
        });
    </script>
</body>
</html>

`
