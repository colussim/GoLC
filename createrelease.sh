
export Release="v1.0.0"
export buildpath="/Users/emmanuel.colussi/Documents/App/Dev/gcloc_m/Release/"
CMD=`PWD`

export GOARCH=arm64
export GOOS=darwin

mkdir -p ${buildpath}${Release}/${GOARCH}/${GOOS}/


go build -o ${buildpath}${Release}/${GOARCH}/${GOOS}/golc golc.go
go build -o ${buildpath}${Release}/${GOARCH}/${GOOS}/ResultsAll ResultsAll.go
cp -r imgs ${buildpath}${Release}/${GOARCH}/${GOOS}/
cp -r dist ${buildpath}${Release}/${GOARCH}/${GOOS}/
cp config_sample.json ${buildpath}${Release}/${GOARCH}/${GOOS}/config.json
cd ${buildpath}${Release}/${GOARCH}/${GOOS}/
zip golc_1.0.0_${GOOS}_${GOARCH}.zip ResultsAll config.json golc -r imgs -r dist
cd $CMD

export GOOS=linux

mkdir -p ${buildpath}${Release}/${GOARCH}/${GOOS}/

go build -o ${buildpath}${Release}/${GOARCH}/${GOOS}/golc golc.go
go build -o ${buildpath}${Release}/${GOARCH}/${GOOS}/ResultsAll ResultsAll.go
cp -r imgs ${buildpath}${Release}/${GOARCH}/${GOOS}/
cp -r dist ${buildpath}${Release}/${GOARCH}/${GOOS}/
cp config_sample.json ${buildpath}${Release}/${GOARCH}/${GOOS}/config.json
cd ${buildpath}${Release}/${GOARCH}/${GOOS}/
zip golc_1.0.0_${GOOS}_${GOARCH}.zip ResultsAll config.json golc -r imgs -r dist
cd $CMD

export GOOS=windows

mkdir -p ${buildpath}${Release}/${GOARCH}/${GOOS}/

go build -o ${buildpath}${Release}/${GOARCH}/${GOOS}/golc golc.go
go build -o ${buildpath}${Release}/${GOARCH}/${GOOS}/ResultsAll ResultsAll.go
cp -r imgs ${buildpath}${Release}/${GOARCH}/${GOOS}/
cp -r dist ${buildpath}${Release}/${GOARCH}/${GOOS}/
cp config_sample.json ${buildpath}${Release}/${GOARCH}/${GOOS}/config.json
cd ${buildpath}${Release}/${GOARCH}/${GOOS}/
zip golc_1.0.0_${GOOS}_${GOARCH}.zip ResultsAll config.json golc -r imgs -r dist
cd $CMD

export GOARCH=amd64
export GOOS=darwin

mkdir -p ${buildpath}${Release}/${GOARCH}/${GOOS}/

go build -o ${buildpath}${Release}/${GOARCH}/${GOOS}/golc golc.go
go build -o ${buildpath}${Release}/${GOARCH}/${GOOS}/ResultsAll ResultsAll.go
cp -r imgs ${buildpath}${Release}/${GOARCH}/${GOOS}/
cp -r dist ${buildpath}${Release}/${GOARCH}/${GOOS}/
cp config_sample.json ${buildpath}${Release}/${GOARCH}/${GOOS}/config.json
cd ${buildpath}${Release}/${GOARCH}/${GOOS}/
zip golc_1.0.0_${GOOS}_${GOARCH}.zip ResultsAll config.json golc -r imgs -r dist
cd $CMD

export GOOS=linux

mkdir -p ${buildpath}${Release}/${GOARCH}/${GOOS}/

go build -o ${buildpath}${Release}/${GOARCH}/${GOOS}/golc golc.go
go build -o ${buildpath}${Release}/${GOARCH}/${GOOS}/ResultsAll ResultsAll.go
cp -r imgs ${buildpath}${Release}/${GOARCH}/${GOOS}/
cp -r dist ${buildpath}${Release}/${GOARCH}/${GOOS}/
cp config_sample.json ${buildpath}${Release}/${GOARCH}/${GOOS}/config.json
cd ${buildpath}${Release}/${GOARCH}/${GOOS}/
zip golc_1.0.0_${GOOS}_${GOARCH}.zip ResultsAll config.json golc -r imgs -r dist
cd $CMD

export GOOS=windows

mkdir -p ${buildpath}${Release}/${GOARCH}/${GOOS}/

go build -o ${buildpath}${Release}/${GOARCH}/${GOOS}/golc golc.go
go build -o ${buildpath}${Release}/${GOARCH}/${GOOS}/ResultsAll ResultsAll.go
cp -r imgs ${buildpath}${Release}/${GOARCH}/${GOOS}/
cp -r dist ${buildpath}${Release}/${GOARCH}/${GOOS}/
cp config_sample.json ${buildpath}${Release}/${GOARCH}/${GOOS}/config.json
cd ${buildpath}${Release}/${GOARCH}/${GOOS}/
zip golc_1.0.0_${GOOS}_${GOARCH}.zip ResultsAll config.json golc -r imgs -r dist
cd $CMD
