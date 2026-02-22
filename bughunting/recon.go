package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

/* ================= COLORS ================= */

func green(s string) string  { return "\033[32m" + s + "\033[0m" }
func yellow(s string) string { return "\033[33m" + s + "\033[0m" }
func blue(s string) string   { return "\033[34m" + s + "\033[0m" }
func red(s string) string    { return "\033[31m" + s + "\033[0m" }

/* ================= PROGRESS BAR ================= */

func progress(task string, duration time.Duration) {
	fmt.Println(blue("[*] " + task))
	for i := 0; i <= 20; i++ {
		bar := strings.Repeat("█", i) + strings.Repeat(" ", 20-i)
		fmt.Printf("\r[%s] %d%%", bar, i*5)
		time.Sleep(duration / 20)
	}
	fmt.Println()
}

/* ================= COMMAND RUNNER ================= */

func runCmd(name string, args ...string) {
	fmt.Println(green("\n[+] Running: " + name + " " + strings.Join(args, " ")))
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}

/* ================= MAIN ================= */

func main() {

	reader := bufio.NewReader(os.Stdin)

	fmt.Println(yellow("===================================="))
	fmt.Println(yellow("  Automated Bug Bounty Recon Tool"))
	fmt.Println(yellow("===================================="))

	/* -------- DOMAIN INPUT -------- */

	fmt.Print(blue("Enter Target Domain: "))
	domain, _ := reader.ReadString('\n')
	domain = strings.TrimSpace(domain)

	/* -------- DIRECTORY SETUP -------- */

	baseDir := "/home/sahilcipher/hunt/"
	targetDir := baseDir + domain

	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		fmt.Println(green("[+] Creating directory: ") + targetDir)
		_ = os.MkdirAll(targetDir, 0755)
	} else {
		fmt.Println(yellow("[*] Using existing directory: ") + targetDir)
	}

	subsFile := targetDir + "/subs.txt"
	activeFile := targetDir + "/active.txt"

	/* -------- INTENSITY -------- */

	fmt.Print(blue("\nSelect Intensity Level (1-3): "))
	level, _ := reader.ReadString('\n')
	level = strings.TrimSpace(level)

	tools := []string{"subfinder", "chaos"}

	switch level {
	case "1":
		fmt.Println(yellow("Intensity 1 → subfinder + chaos-client"))
	case "2":
		fmt.Println(yellow("Intensity 2 → + assetfinder"))
		tools = append(tools, "assetfinder")
	case "3":
		fmt.Println(yellow("Intensity 3 → + assetfinder + sublist3r"))
		tools = append(tools, "assetfinder", "sublist3r")
	default:
		fmt.Println(red("Invalid choice → Defaulting to Level 1"))
	}

	/* -------- MANUAL TOOL SELECTION -------- */

	fmt.Print(blue("\nManually select tools? (y/N): "))
	manual, _ := reader.ReadString('\n')
	manual = strings.TrimSpace(strings.ToLower(manual))

	if manual == "y" {
		tools = []string{}
		fmt.Println(yellow("\n1. subfinder\n2. chaos-client\n3. assetfinder\n4. sublist3r"))
		fmt.Print(blue("Enter choices (comma separated): "))

		choice, _ := reader.ReadString('\n')
		for _, c := range strings.Split(choice, ",") {
			switch strings.TrimSpace(c) {
			case "1":
				tools = append(tools, "subfinder")
			case "2":
				tools = append(tools, "chaos")
			case "3":
				tools = append(tools, "assetfinder")
			case "4":
				tools = append(tools, "sublist3r")
			}
		}
	}

	fmt.Println(green("\nSelected Tools: "), tools)

	/* -------- CLEAN OLD FILE -------- */

	_ = os.WriteFile(subsFile, []byte(""), 0644)

	fmt.Println(yellow("\n===================================="))
	fmt.Println(yellow("        Starting Enumeration"))
	fmt.Println(yellow("===================================="))

	/* -------- TOOL EXECUTION -------- */

	for _, tool := range tools {

		switch tool {

		case "subfinder":
			progress("Running subfinder", 2*time.Second)
			runCmd("subfinder", "-d", domain, "-silent", "-o", subsFile)

		case "chaos":
			progress("Running chaos-client", 2*time.Second)
			api := "YOUR_CHAOS_API_KEY_HERE"
			runCmd("bash", "-c", "chaos -d "+domain+" -key "+api+" -silent >> "+subsFile)

		case "assetfinder":
			progress("Running assetfinder", 2*time.Second)
			runCmd("bash", "-c", "assetfinder --subs-only "+domain+" | tee -a "+subsFile)

		case "sublist3r":
			progress("Running sublist3r", 3*time.Second)
			tmp := targetDir + "/sublist3r.txt"
			runCmd("sublist3r", "-d", domain, "-o", tmp)
			runCmd("bash", "-c", "cat "+tmp+" >> "+subsFile)
		}
	}

	/* -------- CLEAN & DNSX -------- */

	progress("Removing duplicate subdomains", 2*time.Second)
	runCmd("bash", "-c", "sort -u "+subsFile+" -o "+subsFile)

	progress("Finding active subdomains (dnsx)", 3*time.Second)
	runCmd("bash", "-c", "cat "+subsFile+" | dnsx -silent > "+activeFile)

	/* -------- DONE -------- */

	fmt.Println(green("\n===================================="))
	fmt.Println(green(" Recon Completed Successfully "))
	fmt.Println(green(" subs.txt   → All subdomains"))
	fmt.Println(green(" active.txt → Live subdomains"))
	fmt.Println(green(" Output Dir → " + targetDir))
	fmt.Println(green("===================================="))
}
