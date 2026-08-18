package main

import (
	"embed"
	"fmt"
	"os"
	"time"

	bash "github.com/tkdeng/gobash"
	"github.com/tkdeng/goutil"
)

//go:embed assets/theme/*
var assetTheme embed.FS

type themeInstaller struct {
	progressBar *bash.ProgressBar
	opts        *config
}

func installThemeConfig(opts *config) {
	if goutil.Contains(DesktopENV, "gnome") {
		opts.addBool("texteditor-session", "Restore Text Editor Sessions?", true)
		opts.addBool("darktheme", "Use Dark Theme?", true)
	}

	fmt.Println("")
	time.Sleep(1 * time.Second)
}

func installTheme(opts *config) {
	//todo: may skip theme setup on zorin
	// zorinos already has a good theme setup by default
	// or may just skip some parts
	// may also need to test for regular ubuntu distro

	progressBar := bash.NewProgressBar("Installing")
	defer progressBar.Stop()

	theme := &themeInstaller{progressBar: progressBar, opts: opts}

	progressBar.SetSize(4)

	if goutil.Contains(DesktopENV, "gnome") {
		progressBar.AddSize(4)
	}

	fmt.Println("Installing Special Modifications...")

	//* update
	progressBar.Msg("Updating")
	update(true)
	progressBar.Step()

	//todo: figure out ly for optional server minimal gui
	// may move ly option separate from theme (serverTheme or sshTheme or sshGUI)
	// note: also ask user if they want to install a GUI on SSHClient

	progressBar.Msg("Installing Theme Assets")
	extractEmbeddedTarGz("assets/themes.tar.gz", "/usr/share/themes")
	extractEmbeddedTarGz("assets/icons.tar.gz", "/usr/share/icons")
	extractEmbeddedTarGz("assets/sounds.tar.gz", "/usr/share/sounds")
	extractEmbeddedTarGz("assets/backgrounds.tar.gz", "/usr/share/backgrounds")
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

func (theme *themeInstaller) gnome() {
	//* install gnome theme apps
	theme.progressBar.Msg("Installing Gnome Apps")

	if PM == "dnf" {
		installPKG("dconf-editor", "gnome-tweaks")
	} else if PM == "apt" {
		installPKG("dconf-editor", "gnome-tweak-tool")
	}

	bash.Run([]string{`flatpak`, `install`, `-y`, `flathub`, `org.gnome.Extensions`}, "", nil, true)
	bash.Run([]string{`flatpak`, `install`, `-y`, `flathub`, `com.mattjakeman.ExtensionManager`}, "", nil, true)

	bash.Run([]string{`pip3`, `install`, `--upgrade`, `gnome-extensions-cli`}, "", nil)

	theme.progressBar.Step()

	//* config gnome theme
	theme.progressBar.Msg("Configuring Theme Settings")

	if files, err := assetTheme.ReadDir("assets/theme/dconf"); err == nil {
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			if out, err := goutil.JoinPath("/etc/dconf/db/local.d", file.Name()); err == nil {
				if buf, err := assetTheme.ReadFile("assets/theme/dconf/" + file.Name()); err == nil {
					os.WriteFile(out, buf, 0644)
				}
			}
		}
	}

	if PM == "dnf" {
		removePKG("gnome-shell-extension-background-logo")
	}

	if theme.opts.bool("texteditor-session") {
		bash.Run([]string{`gsettings`, `set`, `org.gnome.TextEditor`, `restore-session`, `true`}, "", nil)
	} else {
		bash.Run([]string{`gsettings`, `set`, `org.gnome.TextEditor`, `restore-session`, `false`}, "", nil)
	}

	theme.progressBar.Step()

	if theme.opts.bool("darktheme") {
		bash.RunUserSystemd([]string{`gsettings`, `set`, `org.gnome.desktop.interface`, `color-scheme`, `prefer-dark`}, sudouser, "", nil)
	}

	/* bash.RunUser(`dconf write /org/gnome/desktop/interface/gtk-theme "'Fluent-round-Dark'"`, sudouser, "", nil)
	bash.RunUser(`dconf write /org/gnome/desktop/interface/icon-theme "'ZorinBlue-Dark'"`, sudouser, "", nil)
	bash.RunUser(`dconf write /org/gnome/desktop/sound/theme-name "'zorin'"`, sudouser, "", nil)
	bash.RunUser(`dconf write /org/gnome/desktop/background/picture-uri "'file:///usr/share/backgrounds/tkdeng/blue.webp'"`, sudouser, "", nil)
	bash.RunUser(`dconf write /org/gnome/desktop/background/picture-uri-dark "'file:///usr/share/backgrounds/tkdeng/black.webp'"`, sudouser, "", nil) */

	theme.progressBar.Step()

	//* install gnome extensions
	theme.progressBar.Msg("Installing Gnome Extensions")
	installPKG(
		"gnome-shell-extension-vertical-workspaces",
		"gnome-shell-extension-dash-to-panel",
		"gnome-shell-extension-dash-to-dock",
		"gnome-shell-extension-appindicator",
		"gnome-shell-extension-drive-menu",
	)

	//* install core
	installPKG("gnome-menus")
	installExt("arcmenu@arcmenu.com")
	installExt("aztaskbar@aztaskbar.gitlab.com")
	installExt("openbar@neuromorph")
	installExt("gtk4-ding@smedius.gitlab.com")

	//* install extras
	installExt("clipboard-indicator@tudmotu.com")
	installExt("batterytime@typeof.pw")
	installExt("printers@linuxman.org")
	installExt("Vitals@CoreCoding.com")
	installExt("pop-shell@system76.com")
	installExt("burn-my-windows@schneegans.github.com")
	installExt("compiz-alike-magic-lamp-effect@hermes83.github.com")

	if PM == "dnf" {
		installPKG("gnome-shell-extension-pop-shell")
	} else if PM == "apt" {
		bash.Run([]string{`git`, `clone`, `--depth=1`, `https://github.com/pop-os/shell.git`, `/tmp/pop-shell`}, "", nil)
		bash.RunUser(`make local-install`, sudouser, "/tmp/pop-shell", nil)
		bash.Run([]string{`mv`, `/home/` + sudouser + `/.local/share/gnome-shell/extensions/pop-shell@system76.com`, `/usr/share/gnome-shell/extensions/`}, "", nil)
		bash.Run([]string{`chown`, `-R`, `root:root`, `/usr/share/gnome-shell/extensions/`}, "", nil)
		bash.Run([]string{`chmod`, `-R`, `755`, `/usr/share/gnome-shell/extensions/`}, "", nil)
	}

	bash.Run([]string{`mkdir`, `-p`, `/home/` + sudouser + `/.config/burn-my-windows/profiles`}, "", nil)
	bash.Run([]string{`cp`, `/root/skel/.config/burn-my-windows/profiles/effects.conf`, `/home/` + sudouser + `/.config/burn-my-windows/profiles/effects.conf`}, "", nil)
	bash.Run([]string{`chown`, `-R`, sudouser + `:` + sudouser, `/home/` + sudouser + `/.config`}, "", nil)
	bash.RunRaw(`rm -rf /home/`+sudouser+`/.config/burn-my-windows/profiles/*`, "", nil)
	bash.RunUserSystemd([]string{`dconf`, `write`, `/org/gnome/shell/extensions/burn-my-windows/active-profile`, `'$HOME/.config/burn-my-windows/profiles/effects.conf'`}, sudouser, "", nil)

	//* fix stubborn arcmenu keybinding
	bash.RunUserSystemd([]string{`gsettings`, `set`, `org.gnome.shell.arcmenu`, `arcmenu-hotkey-overlay-key-enabled`, `false`}, sudouser, "", nil)
	bash.RunUserSystemd([]string{`gsettings`, `set`, `org.gnome.shell.arcmenu`, `arcmenu-hotkey-binding`, `'None'`}, sudouser, "", nil)

	theme.progressBar.Step()

	theme.progressBar.Msg("Configuring Gnome Extensions")
	bash.Run([]string{`cp`, `/usr/share/gnome-shell/extensions/*/schemas/*.gschema.xml`, `/usr/share/glib-2.0/schemas/`}, "", nil)
	bash.Run([]string{`glib-compile-schemas`, `/usr/share/glib-2.0/schemas/`}, "", nil)
	bash.Run([]string{`dconf`, `update`}, "", nil)
	theme.progressBar.Step()
}

func installExt(name string) {
	bash.RunUser(`gext install `+name, sudouser, "", nil)
	bash.Run([]string{`mv`, `/home/` + sudouser + `/.local/share/gnome-shell/extensions/` + name, `/usr/share/gnome-shell/extensions/`}, "", nil)
	bash.Run([]string{"sudo", "chown", "-R", "root:root", `/usr/share/gnome-shell/extensions/` + name}, "", nil)
	bash.Run([]string{"sudo", "chmod", "-R", "755", `/usr/share/gnome-shell/extensions/` + name}, "", nil)
}
