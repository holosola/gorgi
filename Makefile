linux:
	@echo "Linux版本"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gorgi main.go

windows:
	@echo "Windows版本"
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o gorgi main.go