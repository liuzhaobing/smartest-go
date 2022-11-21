#/bin/bash
go mod download
go mod verify
go build smartest-go
nohup ./smartest-go -id 27997 &
