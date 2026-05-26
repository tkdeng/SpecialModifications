package main

import (
	"fmt"
	"time"

	bash "github.com/tkdeng/gobash"
	"github.com/tkdeng/goutil"
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

	progressBar.SetSize(1)

	/* if opts.bool("desktop-apps") {
		progressBar.AddSize(5)
	} */

	if goutil.Contains(DesktopENV, "gnome") {
		progressBar.AddSize(1)
	}

	fmt.Println("Installing Special Modifications...")

	//* update
	progressBar.Msg("Updating")
	update(true)
	progressBar.Step()

	//* install desktop environment specific theme
	if goutil.Contains(DesktopENV, "gnome") {
		theme.gnome()
	}

	progressBar.Step()

	//* update
	progressBar.Msg("Updating")
	update(true)
	progressBar.Step()
}

func (apps *themeInstaller) gnome() {
	//* install gnome theme
}
