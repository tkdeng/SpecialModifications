package main

import (
	"bytes"
	"fmt"
	"os"
	"time"

	bash "github.com/tkdeng/gobash"
	"github.com/tkdeng/goutil"
)

type appsInstaller struct {
	progressBar *bash.ProgressBar
	opts        *config
}

func installAppsConfig(opts *config) {
	opts.addBool("google", "Would you like to install Google Chrome?", true)
	opts.addBool("vscode", "Would you like to install VSCode?", true)
	opts.addBool("desktop-apps", "Would you like to install Desktop Apps?", true)
	if opts.bool("desktop-apps") {
		opts.addBool("steam", "Would you like to install Steam?", true)
	}

	fmt.Println("")
	time.Sleep(1 * time.Second)
}

func installApps(opts *config) {
	progressBar := bash.NewProgressBar("Installing")
	defer progressBar.Stop()

	apps := &appsInstaller{progressBar: progressBar, opts: opts}

	progressBar.SetSize(3)

	if opts.bool("desktop-apps") {
		progressBar.AddSize(5)
	}

	if PM == "dnf" {
		progressBar.AddSize(1)
	} else if PM == "apt" {
		progressBar.AddSize(1)
	}

	if opts.bool("google") {
		progressBar.AddSize(1)
	}

	if opts.bool("vscode") {
		progressBar.AddSize(1)
	}

	if opts.bool("steam") {
		progressBar.AddSize(2)
	}

	if goutil.Contains(DesktopENV, "gnome") {
		progressBar.AddSize(4)
	}

	fmt.Println("Installing Special Modifications...")

	//* update
	progressBar.Msg("Updating")
	update(true)
	progressBar.Step()

	//* install Essential apps
	if PM == "dnf" {
		progressBar.Msg("Installing Essential Apps")
		installPKG("gparted", "chromium", "firefox")
		progressBar.Step()
	}else if PM == "apt" {
		progressBar.Msg("Installing Essential Apps")
		installPKG("gparted", "chromium-browser", "firefox")
		progressBar.Step()
	}

	//* install apps
	if opts.bool("desktop-apps") {
		if PM == "dnf" {
			progressBar.Msg("Installing Apps")
			installPKG("blender", "gimp", "gnome-boxes", "audacity", "kdenlive")
			progressBar.Step()
		} else if PM == "apt" {
			progressBar.Msg("Installing Apps")
	
			bash.Run([]string{`add-apt-repository`, `-y`, `ppa:kdenlive/kdenlive-stable`}, "", nil)
			bash.Run([]string{`apt`, `-y`, `update`}, "", nil, true)
	
			installPKG("blender", "gimp", "gnome-boxes", "audacity", "kdenlive")
			progressBar.Step()
		}
	}

	//* install selected apps
	if PM == "dnf" {
		if apps.opts.bool("google") {
			progressBar.Msg("Installing Google Chrome")
			bash.Run([]string{`dnf`, `-y`, `config-manager`, `--set-enabled google-chrome`}, "", nil, true)
			installPKG("google-chrome-stable")
			progressBar.Step()
		}

		if apps.opts.bool("vscode") {
			progressBar.Msg("Installing VSCode")
			bash.Run([]string{`rpm`, `--import`, `https://packages.microsoft.com/keys/microsoft.asc`}, "", nil, true)
			bash.RunRaw(`if ! test -f "/etc/yum.repos.d/vscode.repo" ; then echo '[code]' | sudo tee -a "/etc/yum.repos.d/vscode.repo"; echo 'name=Visual Studio Code' | sudo tee -a "/etc/yum.repos.d/vscode.repo"; echo 'baseurl=https://packages.microsoft.com/yumrepos/vscode' | sudo tee -a "/etc/yum.repos.d/vscode.repo"; echo 'enabled=1' | sudo tee -a "/etc/yum.repos.d/vscode.repo"; echo 'gpgcheck=1' | sudo tee -a "/etc/yum.repos.d/vscode.repo"; echo 'gpgkey=https://packages.microsoft.com/keys/microsoft.asc' | sudo tee -a "/etc/yum.repos.d/vscode.repo"; fi`, "", nil, true)
			bash.Run([]string{`dnf`, `check-update`}, "", nil, true)
			installPKG("code")
			progressBar.Step()
		}

		if apps.opts.bool("steam") {
			progressBar.Msg("Installing Steam")
			bash.Run([]string{`dnf`, `-y`, `module`, `disable`, `nodejs`}, "", nil, true)
			installPKG("steam")
			bash.Run([]string{`dnf`, `-y`, `module`, `install`, `-y`, `--allowerasing`, `nodejs:16/development`}, "", nil, true)
			bash.RunRaw(`if ! [ ! -z $(grep "Steam" "$HOME/.hidden") ] ; then echo 'Steam' | sudo tee -a "$HOME/.hidden"; fi`, "", nil, true)
			bash.RunRaw(`if ! [ ! -z $(grep "Steam" "/etc/skel/.hidden") ] ; then echo 'Steam' | sudo tee -a "/etc/skel/.hidden"; fi`, "", nil, true)
			progressBar.Step()
		}
	} else if PM == "apt" {
		if apps.opts.bool("google") {
			progressBar.Msg("Installing Google Chrome")
			bash.Run([]string{`wget`, `https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb`}, "", nil, true)
			bash.Run([]string{`dpkg`, `-i`, `google-chrome-stable_current_amd64.deb`}, "", nil, true)
			bash.Run([]string{`rm`, `-f`, `google-chrome-stable_current_amd64.deb`}, "", nil, true)
			progressBar.Step()
		}

		if apps.opts.bool("vscode") {
			progressBar.Msg("Installing VSCode")
			bash.Run([]string{`snap`, `install`, `--classic`, `code`}, "", nil, true)
			progressBar.Step()
		}

		if apps.opts.bool("steam") {
			progressBar.Msg("Installing Steam")
			installPKG("steam")
			bash.RunRaw(`if ! [ ! -z $(grep "Steam" "$HOME/.hidden") ] ; then echo 'Steam' | sudo tee -a "$HOME/.hidden"; fi`, "", nil, true)
			bash.RunRaw(`if ! [ ! -z $(grep "Steam" "/etc/skel/.hidden") ] ; then echo 'Steam' | sudo tee -a "/etc/skel/.hidden"; fi`, "", nil, true)
			progressBar.Step()
		}
	}

	if apps.opts.bool("steam") {
		progressBar.Msg("Adding Games Dir")
		os.MkdirAll("/games", 0755)

		if out, err := bash.Run([]string{`ls`, `/home`}, "", nil); err == nil && len(out) != 0 {
			list := bytes.Split(out, []byte{'\n'})
			for _, item := range list {
				item = bytes.TrimSpace(item)
				if len(item) == 0 {
					continue
				}
				user := string(item)

				os.Mkdir("/games/"+user, 0755)
				bash.Run([]string{`chown`, user + `:` + user, `/games/` + user}, "", nil, true)
				bash.Run([]string{`chmod`, `-R`, `700`, `/games/` + user}, "", nil, true)
				bash.Run([]string{`ln`, `-s`, `/games/` + user, `/home/` + user + `/.games`}, "", nil, true)
			}
		}

		progressBar.Step()
	}

	
	if opts.bool("desktop-apps") {
		//* install flatpak apps
		if !hasPKG("liveusb-creator") {
			progressBar.Msg("Installing Media Writer")
			bash.Run([]string{`flatpak`, `install`, `-y`, `flathub`, `org.fedoraproject.MediaWriter`}, "", nil, true)
		}
		progressBar.Step()

		progressBar.Msg("Installing Video Downloader")
		bash.Run([]string{`flatpak`, `install`, `-y`, `flathub`, `com.github.unrud.VideoDownloader`}, "", nil, true)
		progressBar.Step()
	
		/* progressBar.Msg("Installing Kdenlive")
		bash.Run([]string{`flatpak`, `install`, `-y`, `flathub`, `org.kdenlive.kdenlive`}, "", nil, true)
		progressBar.Step() */
	
		progressBar.Msg("Installing OBS Studio")
		bash.Run([]string{`flatpak`, `install`, `-y`, `flathub`, `com.obsproject.Studio`}, "", nil, true)
		progressBar.Step()
	
		progressBar.Msg("Installing Spotify")
		bash.Run([]string{`flatpak`, `install`, `-y`, `flathub`, `com.spotify.Client`}, "", nil, true)
		progressBar.Step()
	}

	//* install desktop environment specific apps
	if goutil.Contains(DesktopENV, "gnome") {
		apps.gnome()
	}

	progressBar.Step()

	//* update
	progressBar.Msg("Updating")
	update(true)
	progressBar.Step()
}

func (apps *appsInstaller) gnome() {
	//* install gnome apps
	if PM == "dnf" {
		apps.progressBar.Msg("Installing Gnome Apps")
		installPKG("dconf-editor", "gnome-tweaks", "nm-connection-editor")
		apps.progressBar.Step()
	} else if PM == "apt" {
		apps.progressBar.Msg("Installing Gnome Apps")
		installPKG("dconf-editor", "gnome-tweak-tool", "network-manager-gnome")
		apps.progressBar.Step()
	}

	//* install nemo file manager
	if PM == "dnf" {
		apps.progressBar.Msg("Finding Nemo")
		installPKG("nemo", "nemo-fileroller")
		bash.Run([]string{`xdg-mime`, `default`, `nemo.desktop`, `inode/directory`, `application/x-gnome-saved-search`}, "", nil, true)
		bash.Run([]string{`sed`, `-r`, `-i`, `s/^OnlyShowIn=/#OnlyShowIn=/m`, `/usr/share/applications/nemo.desktop`}, "", nil, true)
		apps.progressBar.Step()
	} else if PM == "apt" {
		apps.progressBar.Msg("Finding Nemo")
		installPKG("nemo")
		bash.Run([]string{`xdg-mime`, `default`, `nemo.desktop`, `inode/directory`, `application/x-gnome-saved-search`}, "", nil, true)
		apps.progressBar.Step()
	}

	//* install gnome flatpak apps
	apps.progressBar.Msg("Installing Gnome Apps")
	bash.Run([]string{`flatpak`, `install`, `-y`, `flathub`, `org.gnome.Extensions`}, "", nil, true)
	apps.progressBar.Step()
	bash.Run([]string{`flatpak`, `install`, `-y`, `flathub`, `com.mattjakeman.ExtensionManager`}, "", nil, true)
	apps.progressBar.Step()
}
