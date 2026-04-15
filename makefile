BINARY_NAME=goc

build:
	go build -o ${BINARY_NAME} main.go
	rm -f ~/.local/bin/goc
	ln -s ./goc ~/.local/bin/goc
