
default: main.go
	6g main.go
	6l main.6

run: 6.out
	./6.out
