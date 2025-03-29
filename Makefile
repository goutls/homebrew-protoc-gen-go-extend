prepare:
	rm -Rf ./Formula || true
	mkdir ./Formula
	go run main.go