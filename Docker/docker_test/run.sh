docker run --rm -v $(pwd)/Results:/app/Results -v $(pwd)/config.json:/app/config.json golc:arm64 -devops Github

docker run --rm -p 8080:8080 -v $(pwd)/Results:/app/Results resultsall:arm64
