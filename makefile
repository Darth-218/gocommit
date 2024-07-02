BINARY_NAME=gcmt

build:
	go build -o ${BINARY_NAME} main.go
	rm -f ~/.local/bin/${BINARY_NAME}
	ln -s ~/the-duat/coding/go/gocommit/${BINARY_NAME} ~/.local/bin/${BINARY_NAME}
