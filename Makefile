
default: main.go
	6g appconfig.go
	6g main.go
	6l -o gscs main.6

run: default
	./gscs
