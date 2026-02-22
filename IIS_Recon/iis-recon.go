package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
)

const (
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Cyan   = "\033[36m"
	Reset  = "\033[0m"
)

var wg sync.WaitGroup
var mutex sync.Mutex

func runCommandLive(command string, args []string, outfile *os.File) {

	cmd := exec.Command(command, args...)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	cmd.Start()

	multi := io.MultiWriter(os.Stdout, outfile)

	go io.Copy(multi, stdout)
	go io.Copy(multi, stderr)

	cmd.Wait()
}

func checkMethods(url string, outfile *os.File) {
	req, _ := http.NewRequest("OPTIONS", url, nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	mutex.Lock()
	outfile.WriteString("[+] Allowed Methods: " + resp.Header.Get("Allow") + "\n")
	mutex.Unlock()
}

func scanTarget(url string, outfile *os.File) {
	defer wg.Done()

	fmt.Println(Cyan + "\n===================================")
	fmt.Println(" Target: " + url)
	fmt.Println("===================================" + Reset)

	mutex.Lock()
	outfile.WriteString("\n===== TARGET: " + url + " =====\n")
	mutex.Unlock()

	checkMethods(url, outfile)

	fmt.Println(Yellow + "[MODULE] Running shortscan..." + Reset)
	runCommandLive("shortscan", []string{url}, outfile)

	fmt.Println(Yellow + "[MODULE] Running ffuf WebDAV scan..." + Reset)
	runCommandLive("ffuf", []string{
		"-u", url + "/FUZZ",
		"-w", "/usr/share/wordlists/hunter/webdav.txt",
	}, outfile)

	fmt.Println(Yellow + "[MODULE] Running cadaver test..." + Reset)
	runCommandLive("cadaver", []string{url}, outfile)

	fmt.Println(Green + "[✓] Completed: " + url + Reset)
}

func main() {

	reader := bufio.NewReader(os.Stdin)

	fmt.Println(Cyan + "===================================")
	fmt.Println("   IIS / WebDAV Professional Recon  ")
	fmt.Println("===================================" + Reset)

	fmt.Print("Enter URL(s) (space separated): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	urls := strings.Split(input, " ")

	file, _ := os.Create("IIS-Tester-Result.txt")
	defer file.Close()

	for _, url := range urls {
		wg.Add(1)
		go scanTarget(url, file)
	}

	wg.Wait()

	fmt.Println(Green + "\n[+] All scans completed. Results saved to IIS-Tester-Result.txt" + Reset)
}

