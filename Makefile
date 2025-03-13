notebox: main.go
	go build -o box .

clean: box
	rm -rf ./data/*
	rm .metadata.json
	rm ./box
