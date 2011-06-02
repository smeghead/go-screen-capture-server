package main

import (
	"os"
	"time"
	"http"
	"io"
	"log"
	"exec"
	"io/ioutil"
	"fmt"
)

/**
 * Screen capture searver.
 * 
 * # Initialize
 *  - Display setting.
 *  - virtual display start.
 *  - web browser start.
 * # Main Process
 *  - open url.
 *  - take a picture.
 * # End Process
 *  - kill firefox.
 *  - kill virtual display.
 */
const MAX_VD = 1

func InitVirtualScreen() {
	for i := 0; i < MAX_VD; i++ {
		environ := os.Environ()
		environ = append(environ, fmt.Sprintf("DISPLAY=:%d.0", i + 1))
		// Xvfbの起動
		go func (n int) {
			log.Print(n);
			log.Print(environ);
			command := "/usr/bin/Xvfb"
			args := []string {"Xvfb", fmt.Sprintf(":%d", n + 1), "-screen", "0", "1024x768x24"}
			RunCommand(command, args, environ)
		}(i)
		time.Sleep(3000000000)
		// Firefoxの起動
		go func (n int) {
			log.Print(n);
			log.Print(environ);
			command := "/usr/bin/firefox"
			args := []string {"firefox", "-display", fmt.Sprintf(":%d", n + 1), "-width", "1024", "-height", "800"}
			RunCommand(command, args, environ)
		}(i)
	}
}

func RunCommand(command string, args []string, environ []string) {
	log.Print(command)
	log.Print(args)
	cmd, err := exec.Run(command, args, environ, ".", exec.DevNull, exec.Pipe, exec.MergeWithStdout)
	log.Print("Ran")
	if err != nil {
		log.Fatal(err)
		log.Fatal("failed to execute external command.")
		os.Exit(-1)
	}
	b, err := ioutil.ReadAll(cmd.Stdout)
	if err != nil {
		log.Fatal(err)
		log.Fatal("failed to execute external command.")
		os.Exit(-1)
	}
	log.Print("--stdout--")
	log.Print(string(b))
}

// hello world, the web server
func HelloServer(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "hello, world!\n")
}

func main() {
	InitVirtualScreen()
	//RunCommand("/bin/ls", 1);
	http.HandleFunc("/hello", HelloServer)
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.String())
	}
}
