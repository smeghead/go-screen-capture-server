package main

import (
	"os"
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

func RunCommand(command string, display int) {
	environ := os.Environ()
	environ = append(environ, fmt.Sprintf("DISPLAY=%d:0", display))
	var args []string
	args = append(args, "")
	args = append(args, "-l")
	args = append(args, "/home")

	cmd, err := exec.Run(command, args, environ, ".", exec.DevNull, exec.Pipe, exec.Pipe)
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
	log.Print(string(b))
}

// hello world, the web server
func HelloServer(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "hello, world!\n")
}

func main() {
	RunCommand("/bin/ls", 1);
//	http.HandleFunc("/hello", HelloServer)
//	err := http.ListenAndServe(":80", nil)
//	if err != nil {
//		log.Fatal("ListenAndServe: ", err.String())
//	}
}
