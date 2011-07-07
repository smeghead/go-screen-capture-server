package main

import (
	"os"
	"time"
	"http"
	"log"
	"exec"
	"io/ioutil"
	"fmt"
	"template"
	"strings"
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
const MAX_VD = 3
var workingBoxes = map[int] bool {}

/**
  Get enable display number of virtual screen.
  */
func GetDisplay() (int) {
//	for display, working := range workingBoxes {
//		fmt.Printf("[%d:%s]", display, working)
//	}
//	fmt.Println("=")

	for display, working := range workingBoxes {
		if !working {
			workingBoxes[display] = true
			return display
		}
	}

	time.Sleep(1000000000)
	return GetDisplay()
}

var sem = make(chan int, MAX_VD)

/**
  Get byte array content of given filename.
  */
func GetFileByteArray(name string, retry int) ([]byte) {
	if retry > 30 {
		fmt.Printf("ERROR: failed to read file(%s)\n", name)
		return []byte{}
	}
	content, err := ioutil.ReadFile(name)
	if err != nil {
		fmt.Printf("ERROR: failed to read file(%s) %s\n", name, err)
		time.Sleep(1000000000)
		return GetFileByteArray(name, retry + 1)
	}
	return content
}

/**
  Get image byte array of given url web site.
  */
func CaptureUrl(url string) []byte {
	log.Print("CaptureUrl begin")
	sem <- 1    // アクティブキューの空き待ち
	log.Print("found active queue.")
	display := GetDisplay()
	filename := fmt.Sprintf("/home/smeghead/work/go-screen-capture-server/images/tmp_%d.png", display)
	dem := "?";
	if strings.Index(url, dem) > -1 {
		dem = "&"
	}
	//URLにディスプレイ番号として隠しパラメータを付加する。
	url += fmt.Sprintf("%s___n=%d", dem, display)
	log.Print(url)

	os.Remove(filename)
	environ := os.Environ()
	environ = append(environ, fmt.Sprintf("DISPLAY=:%d.0", display))
	command := "/home/smeghead/work/go-screen-capture-server/capture.sh"
	args := []string {"capture.sh", fmt.Sprintf("%d", display), url}
	RunCommand(command, args, environ)
	//ファイルが生成されるまで待つ
	bytes := GetFileByteArray(filename, 1)
	workingBoxes[display] = false // displayの利用が完了したので、返却する。
	<-sem       // 完了。次のリクエストを処理可能にする
	log.Print("CaptureUrl end")
	return bytes
}

func InitVirtualScreen() {
	for i := 0; i < MAX_VD; i++ {
		fmt.Printf("%d\n", i)
		display := i + 1
		environ := os.Environ()
		environ = append(environ, fmt.Sprintf("DISPLAY=:%d.0", i + 1))
		// Xvfbの起動
		go func (d int, env []string) {
			command := "/usr/bin/Xvfb"
			args := []string {"Xvfb", fmt.Sprintf(":%d", d), "-screen", "0", "1024x768x24"}
			RunCommand(command, args, env)
		}(display, environ)
		time.Sleep(3000000000)
		// Firefoxの起動
		go func (d int, env []string) {
			command := "/usr/bin/firefox"
			args := []string {"firefox", "-display", fmt.Sprintf(":%d", display), "-width", "1024", "-height", "800", "-P", fmt.Sprintf("P%d", display)}
			RunCommand(command, args, env)
		}(display, environ)
		time.Sleep(3000000000)
		workingBoxes[display] = false
	}
}

func RunCommand(command string, args []string, environ []string) {
	cmd, err := exec.Run(command, args, environ, ".", exec.DevNull, exec.Pipe, exec.MergeWithStdout)
	log.Printf("Ran [%s]", command)
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
func Capture(w http.ResponseWriter, req *http.Request) {
	image := CaptureUrl(req.FormValue("url"))
	header := w.Header()
	header.Set("Content-Type", "image/png")
	w.Write(image)
}

func Index(w http.ResponseWriter, req *http.Request) {
	p := map[string] string {}
	t, _ := template.ParseFile("index.html", nil)
	t.Execute(w, p)
}

func main() {
	InitVirtualScreen()
	CaptureUrl("http://blog.starbug1.com/")

	http.HandleFunc("/", Index)
	http.HandleFunc("/capture", Capture)
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.String())
	}
}
