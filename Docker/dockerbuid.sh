docker build -t resultsall:arm64 -f Dockerfile.ResultsAll.arm64 .
docker build -t golc:arm64 -f Dockerfile.golc.arm64 .  
docker build -t resultsall:amd64 -f Dockerfile.ResultsAll.amd64 .
docker build -t golc:amd64 -f Dockerfile.golc.amd64 . 

