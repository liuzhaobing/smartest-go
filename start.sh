#/bin/bash
kill -9 $(pidof smartest-go)
git pull
go mod download
go mod verify
go build smartest-go
nohup ./smartest-go -id 27997 &
