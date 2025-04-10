.PHONY: build
build: main.go
	go build -o note .

.PHONY: clean
clean: note
	rm ./note
