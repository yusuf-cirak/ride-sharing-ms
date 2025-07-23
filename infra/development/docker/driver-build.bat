set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
@REM this service only has main package, so we have to pass the folder of the main package as entry point
go build -o build/driver-service ./services/driver-service
