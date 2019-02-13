
release:
	CGO_ENABLED=0 go build -tags release -o main.exe main.go