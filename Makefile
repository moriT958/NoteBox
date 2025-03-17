.PHONY: box
box: main.go
	go build -o box .

clean: box
	rm -rf ./data/*
	rm db.sqlite
	rm ./box
