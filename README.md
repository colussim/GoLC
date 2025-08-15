# 🚧 Redevelopment in Progress

[![Status](https://img.shields.io/badge/status-non--operational-red)]()
[![Branch](https://img.shields.io/badge/branch-redevelopment-orange)]()

> ⚠️ **Warning**  
> This branch is currently being **redeveloped** and is **not operational**.  
> Code in this branch may be incomplete, broken, or non-functional.  
> Use the `main` branch for a stable version.

## Introduction

![logo](imgs/Logob.png)

**GoLC** is a clever abbreviation for "Go Line Counter," drawing inspiration from [CLOC](https://github.com/AlDanial/cloc "AlDanial") and various other line-counting tools in Go like [GCloc](https://github.com/JoaoDanielRufino/gcloc "João Daniel Rufino").

**GoLC** counts physical lines of source code in numerous programming languages across your Bitbucket Cloud, Bitbucket Data Center, GitHub, GitLab, Azure DevOps and local repositories.
GoLC can be used to estimate LoC counts that would be produced by a SonarQube analysis of these projects, without having to implement this analysis.

GoLC The tool analyzes your repositories and identifies the largest branch of each repository, counting the total number of lines of code per language for that branch. At the end of the analysis, a text and PDF report is generated, along with a JSON results file for each repository.It starts an HTTP service to display an HTML page with the results.

