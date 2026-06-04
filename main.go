package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"

	bash "github.com/tkdeng/gobash"
	"github.com/tkdeng/goutil"
	"github.com/tkdeng/regex"
)

//go:embed assets/falcon.txt
var falconTXT []byte

var PM = ""
var hasNalaPM = false
var SSHClient = true
var DesktopENV = []string{}
var AssumeYes = false

var cliArgs = goutil.MapArgs()

var supportedPM = []string{"apt", "dnf"}

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

	if out, _ := bash.RunRaw(`ls -1d /usr/share/{xsessions,wayland-sessions}/*.desktop 2>/dev/null`, "", nil); len(out) != 0 {
		regex.Comp(`(?m)\/([\w_\-]+)\.desktop$`).RepFunc(out, func(b func(int) []byte) []byte {
			dt := bytes.ToLower(b(1))
			DesktopENV = append(DesktopENV, string(dt))

			for _, d := range bytes.Split(dt, []byte{'-'}) {
				DesktopENV = append(DesktopENV, string(d))
			}

			return nil
		})
	}

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
		//* install core
		fmt.Println("")

		lock := bash.SleepLock()

		opts := newConfig()
		installConfig(opts)
		installCore(opts)

		lock.Release()

		return
	} else if cliArgs["apps"] == "true" || cliArgs["a"] == "true" {
		//* install apps
		fmt.Println("")

		lock := bash.SleepLock()

		opts := newConfig()
		installAppsConfig(opts)
		installApps(opts)

		lock.Release()
		return
	} else if cliArgs["theme"] == "true" || cliArgs["t"] == "true" {
		//* install theme
		fmt.Println("")

		lock := bash.SleepLock()

		opts := newConfig()
		installThemeConfig(opts)
		installTheme(opts)

		lock.Release()
		return
	} else if cliArgs["update-kernel"] == "true" || cliArgs["kernel"] == "true" || cliArgs["k"] == "true" {
		//* update linux kernel
		fmt.Println("")

		lock := bash.SleepLock()

		fmt.Println("Not yet implemented")

		lock.Release()
		return
	} else if cliArgs["all"] == "true" || cliArgs["install"] == "true" || cliArgs["i"] == "true" {
		//todo: automatically run all install methods and kernel updates
		// may also include system reboot
		// also remember to include getting all config options before running anything

		lock := bash.SleepLock()

		opts := newConfig()
		installConfig(opts)

		if !SSHClient {
			installAppsConfig(opts)
			installThemeConfig(opts)
		}

		installCore(opts)

		if !SSHClient {
			installApps(opts)
			installTheme(opts)
		}

		//todo: allow SSHClient to optionally install minimal ly gui

		lock.Release()

		//todo: auto reboot
		return
	}

	initPrompt()
}

func initPrompt() {
	sel := bash.InputSelect("What would you like to do?", "Exit", "Install Core", "Install Apps", "Install Theme", "Update Linux Kernel", "Run All")

	switch sel {
	case 1:
		//* install core
		lock := bash.SleepLock()

		opts := newConfig()
		installConfig(opts)
		installCore(opts)

		lock.Release()
		initPrompt()
	case 2:
		//* install apps
		lock := bash.SleepLock()

		if SSHClient {
			fmt.Println("App installation is not supported over SSH connections.")
			initPrompt()
			return
		}

		opts := newConfig()
		installAppsConfig(opts)
		installApps(opts)

		lock.Release()
		initPrompt()
	case 3:
		//* install theme
		lock := bash.SleepLock()

		//todo: install theme (also detect desktop environment for different themes)

		opts := newConfig()
		installThemeConfig(opts)
		installTheme(opts)

		lock.Release()
		initPrompt()
	case 4:
		//* update linux kernel
		lock := bash.SleepLock()

		//todo: update linux kernel

		fmt.Println("Not yet implemented!")

		lock.Release()
		initPrompt()
	case 5:
		lock := bash.SleepLock()

		//todo: automatically run all install methods and kernel updates
		// may also include system reboot
		// also remember to include getting all config options before running anything

		opts := newConfig()
		installConfig(opts)

		if !SSHClient {
			installAppsConfig(opts)
			installThemeConfig(opts)
		}

		installCore(opts)

		if !SSHClient {
			installApps(opts)
			installTheme(opts)
		}

		lock.Release()

		//todo: prompt for reboot
	default:
		fmt.Println("Exiting...")
	}
}
