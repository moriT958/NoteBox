.PHONY: build
build: main.go
	go build -o note .

.PHONY: clean
clean: note
	rm ./note
	rm ~/.config/notebox/.metadata.sqlite
