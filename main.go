package main

import (
	"os"
	"io"
	"time"
	"http"
	"log"
	"rand"
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
type WorkingBox struct {
	displayNo int
	working bool
	lastUrl string
}
var workingBoxes = map[int] *WorkingBox {}

/**
  Get enable display number of virtual screen.
  */
func GetDisplay(url string) (int) {
	for i := 0; i < MAX_VD * 2; i++ {
		display := rand.Intn(MAX_VD) + 1
		log.Printf("display is %d\n", display)
		workingBox := workingBoxes[display]
		//動作中でなくて、以前変換したURLと別であること。同じURLだとキャプチャできないため。これはfirefox addonの方の問題。
		if !workingBox.working && workingBox.lastUrl != url {
			workingBox.working = true
			workingBox.lastUrl = url
			return display
		}
	}
//	for display, workingBox := range workingBoxes {
//		//動作中でなくて、以前変換したURLと別であること。同じURLだとキャプチャできないため。これはfirefox addonの方の問題。
//		if !workingBox.working && workingBox.lastUrl != url {
//			workingBoxes[display].working = true
//			workingBoxes[display].lastUrl = url
//			return display
//		}
//	}

	time.Sleep(1000000000)
	return GetDisplay(url)
}

var sem = make(chan int, MAX_VD)

/**
  Get byte array content of given filename.
  */
func GetFileByteArray(name string, retry int) ([]byte) {
	if retry > 15 {
		fmt.Printf("ERROR: failed to read file(%s)\n", name)
		return []byte{}
	}
	content, err := ioutil.ReadFile(name)
	if err != nil {
		fmt.Printf("INFO: waiting... %s\n", err)
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
	display := GetDisplay(url)
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
	command := "/usr/bin/firefox"
	args := []string {command, "-display", fmt.Sprintf(":%d", display), "-remote", fmt.Sprintf("openUrl(%s)", url), fmt.Sprintf("P%d", display)}
	RunCommand(command, args, environ)
	//ファイルが生成されるまで待つ
	bytes := GetFileByteArray(filename, 1)
	workingBoxes[display].working = false // displayの利用が完了したので、返却する。
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
			args := []string {command, fmt.Sprintf(":%d", d), "-screen", "0", "1024x768x24"}
			RunCommand(command, args, env)
		}(display, environ)
		time.Sleep(3000000000)
		// Firefoxの起動
		go func (d int, env []string) {
			command := "/usr/bin/firefox"
			args := []string {command, "-display", fmt.Sprintf(":%d", display), "-width", "1024", "-height", "800", "-P", fmt.Sprintf("P%d", display)}
			RunCommand(command, args, env)
		}(display, environ)
		time.Sleep(3000000000)
		// WorkingBoxesの初期化
		workingBoxes[display] = &WorkingBox{display, false, ""}
	}
	rand.Seed(time.Nanoseconds() % 1e9)
}

func RunCommand(command string, args []string, environ []string) {
	cmd := exec.Command(command)
	cmd.Env = environ
	cmd.Args = args
	cmd.Dir = "."
	stdout, err1 := cmd.StdoutPipe()
	if err1 != nil {
		log.Fatal(err1)
		log.Fatal("failed to retrieve pipe.")
		os.Exit(-1)
	}

	log.Printf("Ran [%s]", command)
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
		log.Fatal("failed to execute external command.")
		os.Exit(-1)
	}
	
	WriteFileLines(stdout)
}

func WriteFileLines(reader io.Reader) {
	var (
		err os.Error
		n int
	)
	buf := make([]byte, 1024)

	log.Println("WriteFileLines");
	for {
		if n, err = reader.Read(buf); err != nil {
			break
		}
		log.Print(string(buf[0:n]))
	}
	if err == os.EOF {
		log.Println("stdout end");
		err = nil
	} else {
		log.Println("ERROR: " + err.String());
	}
}

// hello world, the web server
func Capture(w http.ResponseWriter, req *http.Request) {
	url := req.FormValue("url")
	header := w.Header()
	if url == "" {
		log.Printf("ERROR: url is required. (%s)\n", url)
		w.WriteHeader(http.StatusInternalServerError)
		header.Set("Content-Type", "text/plian;charset=UTF-8;")
		io.WriteString(w, "Internal Server Error: please input url.\n")
		return
	}
	if strings.Index(url, "http") != 0 {
		log.Printf("ERROR: url is invalid. (%s)\n", url)
		w.WriteHeader(http.StatusInternalServerError)
		header.Set("Content-Type", "text/plian;charset=UTF-8;")
		io.WriteString(w, "Internal Server Error: please input valid url.\n")
		return
	}
	image := CaptureUrl(url)
	if len(image) == 0 {
		log.Printf("ERROR: failed to capture. (%s)\n", url)
		w.WriteHeader(http.StatusInternalServerError)
		header.Set("Content-Type", "text/plian;charset=UTF-8;")
		io.WriteString(w, "Internal Server Error: Failed to capture.\n")
		return
	}
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
//	CaptureUrl("http://blog.starbug1.com/")

	portNo := 1975
	http.HandleFunc("/", Index)
	http.HandleFunc("/capture", Capture)
	err := http.ListenAndServe(fmt.Sprintf(":%d", portNo), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.String())
	}
	log.Printf("INFO: start server on %d\n", portNo)
}
