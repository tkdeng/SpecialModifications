package main

import (
	_ "embed"
	"fmt"
	"os"

	bash "github.com/tkdeng/gobash"
	"github.com/tkdeng/goutil"
)

//go:embed assets/falcon.txt
var falconTXT []byte

var PM = ""
var SSHClient = true
var AssumeYes = false

var cliArgs = goutil.MapArgs()

func main() {
	fmt.Println("Special Modifacations by TKD Engineer")

	fmt.Println(string(falconTXT))

	if out, err := bash.Run([]string{`which`, `apt`}, "", nil); err == nil && len(out) != 0 {
		PM = "apt"
	} else if out, err := bash.Run([]string{`which`, `dnf`}, "", nil); err == nil && len(out) != 0 {
		PM = "dnf"
	} else {
		fmt.Println("Unsupported Linux Distribution")
		return
	}

	SSHClient = !bash.If(`"$SSH_CLIENT" == "" && "$SSH_TTY" == ""`, "", nil)

	if cliArgs["help"] == "true" || cliArgs["h"] == "true" {
		//todo: add help message
		return
	}

	if cliArgs["assume-yes"] == "true" || cliArgs["y"] == "true" {
		AssumeYes = true
	}

	if os.Geteuid() != 0 {
		fmt.Println("This program must be run as root (use sudo)")
		return
	}

	if cliArgs["install"] == "true" || cliArgs["core"] == "true" || cliArgs["i"] == "true" {
		fmt.Println("")
		installCore()
		return
	} else if cliArgs["apps"] == "true" || cliArgs["a"] == "true" {
		fmt.Println("")
		fmt.Println("Not yet implemented")
		return
	} else if cliArgs["theme"] == "true" || cliArgs["t"] == "true" {
		fmt.Println("")
		fmt.Println("Not yet implemented")
		return
	} else if cliArgs["update-kernel"] == "true" || cliArgs["kernel"] == "true" || cliArgs["k"] == "true" {
		fmt.Println("")
		fmt.Println("Not yet implemented")
		return
	}

	initPrompt()
}

func initPrompt() {
	sel := bash.InputSelect("What would you like to do?", "Exit", "Install Core", "Install Apps", "Install Theme", "Update Linux Kernel")

	switch sel {
	case 1:
		installCore()
		initPrompt()
	case 2:
		fmt.Println("Not yet implemented!")
		initPrompt()
	case 3:
		fmt.Println("Not yet implemented!")
		initPrompt()
	case 4:
		fmt.Println("Not yet implemented!")
		initPrompt()
	default:
		fmt.Println("Exiting...")
	}
}
