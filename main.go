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
var hasNalaPM = false
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

	if out, err := bash.Run([]string{`which`, `nala`}, "", nil); err == nil && len(out) != 0 {
		hasNalaPM = true
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

	if cliArgs["core"] == "true" || cliArgs["c"] == "true" {
		fmt.Println("")
		opts := newConfig()
		installConfig(opts)
		installCore(opts)
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
	} else if cliArgs["all"] == "true" || cliArgs["install"] == "true" || cliArgs["i"] == "true" {
		//todo: automatically run all install methods and kernel updates
		// may also include system reboot
		// also remember to include getting all config options before running anything

		opts := newConfig()
		installConfig(opts)

		installCore(opts)
		return
	}

	initPrompt()
}

func initPrompt() {
	sel := bash.InputSelect("What would you like to do?", "Exit", "Install Core", "Install Apps", "Install Theme", "Update Linux Kernel", "Run All")

	switch sel {
	case 1:
		opts := newConfig()
		installConfig(opts)
		installCore(opts)
		initPrompt()
	case 2:
		//todo: install apps
		fmt.Println("Not yet implemented!")
		initPrompt()
	case 3:
		//todo: install theme (also detect desktop environment for different themes)
		fmt.Println("Not yet implemented!")
		initPrompt()
	case 4:
		//todo: update linux kernel
		fmt.Println("Not yet implemented!")
		initPrompt()
	case 5:
		//todo: automatically run all install methods and kernel updates
		// may also include system reboot
		// also remember to include getting all config options before running anything

		opts := newConfig()
		installConfig(opts)

		installCore(opts)
	default:
		fmt.Println("Exiting...")
	}
}
