package main

import (
	"fmt"
	"time"

	bash "github.com/tkdeng/gobash"
)

type themeInstaller struct {
	progressBar *bash.ProgressBar
	opts        *config
}

func installThemeConfig(opts *config) {
	// opts.addBool("google", "Would you like to install Google Chrome?", true)

	fmt.Println("")
	time.Sleep(1 * time.Second)
}

func installTheme(opts *config) {
	progressBar := bash.NewProgressBar("Installing")
	defer progressBar.Stop()

	theme := &themeInstaller{progressBar: progressBar, opts: opts}
	_ = theme

	progressBar.SetSize(1)

	/* if opts.bool("desktop-apps") {
		progressBar.AddSize(5)
	} */

	fmt.Println("Installing Special Modifications...")

	//* update
	progressBar.Msg("Updating")
	update(true)
	progressBar.Step()

	/* if PM == "dnf" {
		progressBar.Msg("Installing Essential Apps")
		installPKG("gparted", "chromium", "firefox")
		progressBar.Step()
	}else if PM == "apt" {
		progressBar.Msg("Installing Essential Apps")
		installPKG("gparted", "chromium-browser", "firefox")
		progressBar.Step()
	} */
}
