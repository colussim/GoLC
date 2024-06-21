find . -type f -name '*.go' -not -name '*_test.go' | while read file; do
    gotests -w -all "$file"
done
