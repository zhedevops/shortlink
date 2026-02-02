gmod:
	go mod tidy

bld:
	go build ./...

build:
	go build -o cmd/shortener/shortener ./cmd/shortener

test:
	go test -count 1 ./...

testcov:
	go test -count 1 ./... -cover

vet:
	go vet ./...

vetstat:
	go vet -vettool=$(which statictest) ./...

runfl:
	go run cmd/shortener/main.go -a=localhost:8888 -b=ftp://localhost:8383 -l=fatal -f=data/files/flagpath/store.json

run:
	go run cmd/shortener/main.go -a=localhost:8888 -b=ftp://localhost:8383 -l=fatal

