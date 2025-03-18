.PHONY: notebox
notebox: main.go
	go build -o notebox .

.PHONY: clean
clean: notebox
	rm ./notebox
