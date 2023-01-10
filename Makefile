start:
	git pull
	go build -o rating.exe -tags=jsoniter cmd/app/*.go
	./rating.exe