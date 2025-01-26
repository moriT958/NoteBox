notebox: main.go
	go build -o notebox main.go

clean: notebox
	rm ./notebox
