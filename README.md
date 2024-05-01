![Static Badge](https://img.shields.io/badge/Go-v1.22-blue:)

## Introduction

![architecture](imgs/Logo.png)

**GoLC** is a clever abbreviation for "Go Line Counter," drawing inspiration from [CLOC](https://github.com/AlDanial/cloc "AlDanial") and various other line-counting tools in Go like [GCloc](https://github.com/JoaoDanielRufino/gcloc "João Daniel Rufino").

**GoLC** counts blank lines, comment lines, and physical lines of source code in numerous programming languages supported by the Developer, Enterprise, and Data Center editions of [SonarQube](https://www.sonarsource.com/knowledge/languages/) across your Bitbucket Cloud, Bitbucket Data Center, GitHub, GitLab, and Azure DevOps repositories.These editions are billed per instance per year and are based on the number of lines of code (LOC).

GoLC scans your repositories and identifies the main branch of each repository, tallying the total lines of code per language for that branch.
At the end of the analysis, a text and PDF report is generated, along with a JSON result file for each repository. It starts an HTTP service to display an HTML page with the results.

> This initial version is available for Bitbucket Cloud and Bitbucket DC, and for GitHub, GitLab, Azure DevOps, and Files the next updates will be available soon, integrating these platforms.A Docker version will be planned.

---
## Installation

You can install from the stable release by clicking here

## Prerequisites 

* A personal access tokens for : Bitbucket Cloud,Bitbucket DC,GitHub, GitLab and Azure DevOps.The token must have repo scope.
* [Go language installed](https://go.dev/) : If you want to use the sources...

 ## Usage


✅ Using configuration for DevOps platform 'BitBucket'

❗️ Directory <'Results'> already exists. Do you want to delete it? (y/n): y

❗️ Do you want to create a backup of the directory before deleting? (y/n): n


🔎 Analysis of devops platform objects ...

✅ The number of project(s) to analyze is 8

        🟢  Analyse Projet: test2 
          ✅ The number of Repositories found is: 1

        🟢  Analyse Projet: tests 
          ✅ The number of Repository found is: 1
        ✅ Repo: testempty - Number of branches: 1

        🟢  Analyse Projet: LSA 
          ✅ The number of Repository found is: 0

        🟢  Analyse Projet: AdfsTestingTools 
          ✅ The number of Repository found is: 0

        🟢  Analyse Projet: cloc 
          ✅ The number of Repository found is: 1
        ✅ Repo: gcloc - Number of branches: 2

        🟢  Analyse Projet: sri 
          ✅ The number of Repository found is: 0

        🟢  Analyse Projet: Bitbucket Pipes 
          ✅ The number of Repository found is: 5
        ✅ Repo: sonarcloud-quality-gate - Number of branches: 9
        ✅ Repo: sonarcloud-scan - Number of branches: 8
        ✅ Repo: official-pipes - Number of branches: 14
        ✅ Repo: sonarqube-scan - Number of branches: 7
        ✅ Repo: sonarqube-quality-gate - Number of branches: 2

        🟢  Analyse Projet: SonarCloud Analysis Samples 
          ✅ The number of Repository found is: 4
        ✅ Repo: sample-maven-project - Number of branches: 6
        ✅ Repo: sample-gradle-project - Number of branches: 3
        ✅ Repo: sample-nodejs-project - Number of branches: 6
        ✅ Repo: sample-dotnet-project-azuredevops - Number of branches: 2

✅ The largest repo is <sample-nodejs-project> in the project <SAMPLES> with the branch <demo-app-week> and a size of 425.45 KB

✅ Total size of your organization's repositories: 877.65 KB

✅ Total repositories analyzed: 11 - Find empty : 1

🔎 Analysis of Repos ...

Extracting files from repo : testempty 
        ✅ json report exported to Results/Result_TES_testempty_main.json
        ✅ The repository <testempty> has been analyzed
                                                                                                    
        ✅ json report exported to Results/Result_CLOC_gcloc_DEV.json
        ✅ The repository <gcloc> has been analyzed
                                                                                              
        ✅ json report exported to Results/Result_BBPIPES_sonarcloud-quality-gate_master.json
        ✅ The repository <sonarcloud-quality-gate> has been analyzed
                                                                                              
        ✅ json report exported to Results/Result_BBPIPES_sonarcloud-scan_master.json
        ✅ The repository <sonarcloud-scan> has been analyzed
         ........

🔎 Analyse Report ...

✅ Number of Repository analyzed in Organization <sonar-demo> is 11 

✅ The repository with the largest line of code is in project <CLOC> the repo name is <gcloc> with <2.05M> lines of code

✅ The total sum of lines of code in Organization <sonar-demo> is : 2.06M Lines of Code


✅ Reports are located in the <'Results'> directory

✅ Time elapsed : 00:01:01

ℹ️  To generate and visualize results on a web interface, follow these steps: 

        ✅ run Analysis

        ✅ run Results