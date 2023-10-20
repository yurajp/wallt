build:
	go build -o bin/wallt cmd/main.go    
run:
	@go run cmd/main.go
	
compile:
	@GOOS=darwin GOARCH=amd64 go build -o bin/wallt-darwin-amd64 cmd/main.go
	
	@GOOS=linux GOARCH=amd64 go build -o bin/wallt-linux-amd64 cmd/main.go
	
	@GOOS=windows GOARCH=amd64 go build -o bin/wallt-windows-amd64 cmd/main.go
