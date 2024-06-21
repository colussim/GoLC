#!/bin/bash

LIST=`cat pkglist.txt` 

go test -coverprofile=coverage.out $LIST
gocover-cobertura < coverage.out > cobertura-coverage.xml
