package main

import (
	"fmt"
	"time"

	bash "github.com/tkdeng/gobash"
)

type appsInstaller struct {
	progressBar *bash.ProgressBar
	opts        *config
}

func installAppsConfig(opts *config) {
	//todo: add app config

	time.Sleep(1 * time.Second)
}

func installApps(opts *config) {
	progressBar := bash.NewProgressBar("Installing")
	defer progressBar.Stop()

	apps := &appsInstaller{progressBar: progressBar, opts: opts}

	progressBar.SetSize(1)

	fmt.Println("Installing Special Modifications...")

	//* update
	progressBar.Msg("Updating")
	update(true)
	progressBar.Step()

	//todo: detect desktop environment
	apps.gnome()
}

func (apps *appsInstaller) gnome() {
	
}
