build:
	go build -o apitm cmd/app/main.go

run:
	swag init -g cmd/app/main.go
	go run cmd/app/main.go