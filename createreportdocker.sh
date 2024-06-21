export Release="v1.0.3"
export buildpath=" /Users/emmanuel.colussi/Documents/App/Dev/gcloc_m/Release/" 
CMD=`PWD`

export GOARCH=arm64
export GOOS=linux

mkdir -p ${buildpath}${Release}/Docker/${GOARCH}/${GOOS}/

go build -ldflags "-X main.version=${Release}" -o ${buildpath}${Release}/Docker/${GOARCH}/${GOOS}/ReslultsAllDocker ResultsAll/ResultsAllDocker.go

export GOARCH=amd64

mkdir -p ${buildpath}${Release}/Docker/${GOARCH}/${GOOS}/

go build -ldflags "-X main.version=${Release}" -o ${buildpath}${Release}/Docker/${GOARCH}/${GOOS}/ReslultsAllDocker ResultsAll/ResultsAllDocker.go
