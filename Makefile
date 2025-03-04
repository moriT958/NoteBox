notebox: main.go
	go build -o note .

clean: notebox
	rm -rf ./data/*
	rm .metadata.json
	rm ./note
