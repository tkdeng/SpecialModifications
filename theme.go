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

	//todo: setup theme install

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
	theme.progressBar.Msg("Installing Gnome Theme Apps")
	if PM == "dnf" {
		installPKG("dconf-editor", "gnome-tweaks")
	} else if PM == "apt" {
		installPKG("dconf-editor", "gnome-tweak-tool")
	}
	bash.Run([]string{`flatpak`, `install`, `-y`, `flathub`, `org.gnome.Extensions`}, "", nil, true)
	bash.Run([]string{`flatpak`, `install`, `-y`, `flathub`, `com.mattjakeman.ExtensionManager`}, "", nil, true)
	theme.progressBar.Step()

	//* config gnome theme
	theme.progressBar.Msg("Configuring Theme Settings")
	/* bash.Run([]string{`gsettings`, `set`, `org.gnome.desktop.interface`, `clock-format`, `12h`}, "", nil)
	bash.Run([]string{`gsettings`, `set`, `org.gnome.mutter`, `center-new-windows`, `true`}, "", nil)
	bash.Run([]string{`gsettings`, `set`, `org.gnome.mutter`, `attach-modal-dialogs`, `false`}, "", nil)
	bash.Run([]string{`gsettings`, `set`, `org.gnome.desktop.wm.preferences`, `button-layout`, `appmenu:minimize,maximize,close`}, "", nil) */

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
		bash.Run([]string{`gsettings`, `set`, `org.gnome.desktop.interface`, `color-scheme`, `prefer-dark`}, "", nil)
	}

	bash.Run([]string{`gsettings`, `set`, `org.gnome.desktop.interface`, `gtk-theme`, `Fluent-round-Dark`}, "", nil)
	bash.Run([]string{`gsettings`, `set`, `org.gnome.desktop.interface`, `icon-theme`, `ZorinBlue-Dark`}, "", nil)
	bash.Run([]string{`gsettings`, `set`, `org.gnome.desktop.sound`, `theme-name`, `zorin`}, "", nil)
	bash.Run([]string{`gsettings`, `set`, `org.gnome.desktop.background`, `picture-uri`, `file:///usr/share/backgrounds/tkdeng/blue.webp`}, "", nil)
	bash.Run([]string{`gsettings`, `set`, `org.gnome.desktop.background`, `picture-uri-dark`, `file:///usr/share/backgrounds/tkdeng/black.webp`}, "", nil)

	theme.progressBar.Step()

	//* install gnome extensions
	theme.progressBar.Msg("Installing Gnome Extensions")
	// installPKG("gnome-shell-extension-vertical-workspaces", "gnome-shell-extension-arc-menu", "gnome-shell-extension-dash-to-panel", "gnome-shell-extension-dash-to-dock")
	installPKG("gnome-shell-extension-vertical-workspaces", "gnome-shell-extension-dash-to-panel", "gnome-shell-extension-dash-to-dock")

	//* install arcmenu
	installPKG("gnome-menus")
	bash.Run([]string{`git`, `clone`, `https://gitlab.com/arcmenu/ArcMenu.git`, `/usr/share/gnome-shell/extensions/arcmenu@arcmenu.com`}, "", nil)
	bash.Run([]string{`glib-compile-schemas`, `/usr/share/gnome-shell/extensions/arcmenu@arcmenu.com/schemas/`}, "", nil)
	bash.Run([]string{`chmod`, `-R`, `755`, `/usr/share/gnome-shell/extensions/arcmenu@arcmenu.com`}, "", nil)

	//* install app icons taskbar
	bash.Run([]string{`git`, `clone`, `https://gitlab.com/AndrewZaech/aztaskbar.git`, `/usr/share/gnome-shell/extensions/aztaskbar@aztaskbar.gitlab.com`}, "", nil)
	bash.Run([]string{`glib-compile-schemas`, `/usr/share/gnome-shell/extensions/aztaskbar@aztaskbar.gitlab.com/schemas/`}, "", nil)
	bash.Run([]string{`chmod`, `-R`, `755`, `/usr/share/gnome-shell/extensions/aztaskbar@aztaskbar.gitlab.com`}, "", nil)

	//* install open bar
	bash.Run([]string{`git`, `clone`, `https://github.com/neuromorph/openbar.git`, `/usr/share/gnome-shell/extensions/openbar@neuromorph`}, "", nil)
	bash.Run([]string{`glib-compile-schemas`, `/usr/share/gnome-shell/extensions/openbar@neuromorph/schemas/`}, "", nil)
	bash.Run([]string{`chmod`, `-R`, `755`, `/usr/share/gnome-shell/extensions/openbar@neuromorph`}, "", nil)

	theme.progressBar.Step()

	theme.progressBar.Msg("Configuring Gnome Extensions")
	bash.Run([]string{`dconf`, `update`}, "", nil)
	theme.progressBar.Step()

	return

	//* install gnome-extensions-cli
	theme.progressBar.Msg("Installing Extensions CLI")
	// bash.Run([]string{`pip3`, `install`, `--upgrade`, `gnome-extentions-cli`}, "", nil)
	bash.Run([]string{`pip3`, `install`, `--upgrade`, `gnome-extensions-cli`}, "", nil)
	theme.progressBar.Step()

	//todo: install via dnf
	// installPKG("gnome-shell-extension-vertical-workspaces")

	//* install arcmenu
	theme.progressBar.Msg("Installing arcmenu")
	// bash.Run([]string{`gext`, `-F`, `install`, `arcmenu@arcmenu.com`}, "", nil)
	installPKG("gnome-shell-extension-arc-menu")

	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/arcmenu@arcmenu.com/schemas/`, `set`, `org.gnome.shell.extensions.arcmenu`, `arcmenu-hotkey-overlay-key-enabled`, `false`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/arcmenu@arcmenu.com/schemas/`, `set`, `org.gnome.shell.extensions.arcmenu`, `hide-overview-on-startup`, `true`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/arcmenu@arcmenu.com/schemas/`, `set`, `org.gnome.shell.extensions.arcmenu`, `default-menu-view`, `Pinned_And_Frequent_Apps`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/arcmenu@arcmenu.com/schemas/`, `set`, `org.gnome.shell.extensions.arcmenu`, `arcmenu-layout-max-frequent-apps`, `16`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/arcmenu@arcmenu.com/schemas/`, `set`, `org.gnome.shell.extensions.arcmenu`, `vert-separator`, `true`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/arcmenu@arcmenu.com/schemas/`, `set`, `org.gnome.shell.extensions.arcmenu`, `show-external-devices`, `true`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/arcmenu@arcmenu.com/schemas/`, `set`, `org.gnome.shell.extensions.arcmenu`, `show-bookmarks`, `true`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/arcmenu@arcmenu.com/schemas/`, `set`, `org.gnome.shell.extensions.arcmenu`, `directory-shortcuts`, `[{'name': 'Home', 'icon': 'user-home-symbolic', 'id': 'ArcMenu_Home'}, {'name': 'Documents', 'icon': '. GThemedIcon folder-documents-symbolic folder-symbolic folder-documents folder', 'id': 'ArcMenu_Documents'}, {'name': 'Downloads', 'icon': '. GThemedIcon folder-download-symbolic folder-symbolic folder-download folder', 'id': 'ArcMenu_Downloads'}, {'name': 'Music', 'icon': '. GThemedIcon folder-music-symbolic folder-symbolic folder-music folder', 'id': 'ArcMenu_Music'}, {'name': 'Pictures', 'icon': '. GThemedIcon folder-pictures-symbolic folder-symbolic folder-pictures folder', 'id': 'ArcMenu_Pictures'}, {'name': 'Videos', 'icon': '. GThemedIcon folder-videos-symbolic folder-symbolic folder-videos folder', 'id': 'ArcMenu_Videos'}, {'name': 'Recent', 'icon': 'document-open-recent-symbolic', 'id': 'ArcMenu_Recent'}]`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/arcmenu@arcmenu.com/schemas/`, `set`, `org.gnome.shell.extensions.arcmenu`, `application-shortcuts`, `[{'name': 'Software', 'icon': 'org.gnome.Software', 'id': 'ArcMenu_Software'}, {'name': 'Settings', 'icon': 'org.gnome.Settings', 'id': 'org.gnome.Settings.desktop'}, {'name': 'Terminal', 'icon': 'org.gnome.Terminal', 'id': 'org.gnome.Terminal.desktop'}, {'name': 'System Monitor', 'icon': 'org.gnome.SystemMonitor', 'id': 'gnome-system-monitor'}]`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/arcmenu@arcmenu.com/schemas/`, `set`, `org.gnome.shell.extensions.arcmenu`, `search-provider-open-windows`, `true`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/arcmenu@arcmenu.com/schemas/`, `set`, `org.gnome.shell.extensions.arcmenu`, `search-provider-recent-files`, `true`}, "", nil)
	theme.progressBar.Step()

	//* install dash to panel
	theme.progressBar.Msg("Installing Dash To Panel")
	// bash.Run([]string{`gext`, `-F`, `install`, `dash-to-panel@jderose9.github.com`}, "", nil)
	installPKG("gnome-shell-extension-dash-to-panel")

	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/dash-to-panel@jderose9.github.com/schemas/`, `set`, `org.gnome.shell.extensions.dash-to-panel`, `panel-element-positions-monitors-sync`, `true`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/dash-to-panel@jderose9.github.com/schemas/`, `set`, `org.gnome.shell.extensions.dash-to-panel`, `show-showdesktop-hover`, `true`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/dash-to-panel@jderose9.github.com/schemas/`, `set`, `org.gnome.shell.extensions.dash-to-panel`, `panel-element-positions`, `{"1":[{"element":"showAppsButton","visible":false,"position":"stackedTL"},{"element":"activitiesButton","visible":false,"position":"stackedTL"},{"element":"leftBox","visible":true,"position":"stackedTL"},{"element":"taskbar","visible":true,"position":"stackedTL"},{"element":"centerBox","visible":true,"position":"stackedBR"},{"element":"rightBox","visible":true,"position":"stackedBR"},{"element":"systemMenu","visible":false,"position":"stackedBR"},{"element":"dateMenu","visible":true,"position":"stackedBR"},{"element":"desktopButton","visible":false,"position":"stackedBR"}],"AUO-0x00000000":[{"element":"showAppsButton","visible":false,"position":"stackedTL"},{"element":"activitiesButton","visible":false,"position":"stackedTL"},{"element":"leftBox","visible":true,"position":"stackedTL"},{"element":"taskbar","visible":true,"position":"stackedTL"},{"element":"centerBox","visible":true,"position":"stackedBR"},{"element":"rightBox","visible":true,"position":"stackedBR"},{"element":"systemMenu","visible":true,"position":"stackedBR"},{"element":"dateMenu","visible":true,"position":"stackedBR"},{"element":"desktopButton","visible":true,"position":"stackedBR"}]}`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/dash-to-panel@jderose9.github.com/schemas/`, `set`, `org.gnome.shell.extensions.dash-to-panel`, `isolate-workspaces`, `true`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/dash-to-panel@jderose9.github.com/schemas/`, `set`, `org.gnome.shell.extensions.dash-to-panel`, `isolate-monitors`, `true`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/dash-to-panel@jderose9.github.com/schemas/`, `set`, `org.gnome.shell.extensions.dash-to-panel`, `hide-overview-on-startup`, `true`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/dash-to-panel@jderose9.github.com/schemas/`, `set`, `org.gnome.shell.extensions.dash-to-panel`, `panel-size`, `32`}, "", nil)
	theme.progressBar.Step()

	//* install dash to dock
	theme.progressBar.Msg("Installing Dash To Dock")
	// bash.Run([]string{`gext`, `-F`, `install`, `dash-to-dock@micxgx.gmail.com`}, "", nil)
	installPKG("gnome-shell-extension-dash-to-dock")

	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/dash-to-dock@micxgx.gmail.com/schemas/`, `set`, `org.gnome.shell.extensions.dash-to-dock`, `multi-monitor`, `true`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/dash-to-dock@micxgx.gmail.com/schemas/`, `set`, `org.gnome.shell.extensions.dash-to-dock`, `dash-max-icon-size`, `32`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/dash-to-dock@micxgx.gmail.com/schemas/`, `set`, `org.gnome.shell.extensions.dash-to-dock`, `dock-fixed`, `false`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/dash-to-dock@micxgx.gmail.com/schemas/`, `set`, `org.gnome.shell.extensions.dash-to-dock`, `autohide`, `true`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/dash-to-dock@micxgx.gmail.com/schemas/`, `set`, `org.gnome.shell.extensions.dash-to-dock`, `intellihide-mode`, `FOCUS_APPLICATION_WINDOWS`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/dash-to-dock@micxgx.gmail.com/schemas/`, `set`, `org.gnome.shell.extensions.dash-to-dock`, `icon-size-fixed`, `true`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/dash-to-dock@micxgx.gmail.com/schemas/`, `set`, `org.gnome.shell.extensions.dash-to-dock`, `isolate-workspaces`, `true`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/dash-to-dock@micxgx.gmail.com/schemas/`, `set`, `org.gnome.shell.extensions.dash-to-dock`, `isolate-monitors`, `true`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/dash-to-dock@micxgx.gmail.com/schemas/`, `set`, `org.gnome.shell.extensions.dash-to-dock`, `show-show-apps-button`, `false`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/dash-to-dock@micxgx.gmail.com/schemas/`, `set`, `org.gnome.shell.extensions.dash-to-dock`, `show-trash`, `false`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/dash-to-dock@micxgx.gmail.com/schemas/`, `set`, `org.gnome.shell.extensions.dash-to-dock`, `show-mounts`, `false`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/dash-to-dock@micxgx.gmail.com/schemas/`, `set`, `org.gnome.shell.extensions.dash-to-dock`, `custom-theme-shrink`, `true`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/dash-to-dock@micxgx.gmail.com/schemas/`, `set`, `org.gnome.shell.extensions.dash-to-dock`, `disable-overview-on-startup`, `true`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/dash-to-dock@micxgx.gmail.com/schemas/`, `set`, `org.gnome.shell.extensions.dash-to-dock`, `apply-custom-theme`, `true`}, "", nil)
	theme.progressBar.Step()

	//* install app icons taskbar
	theme.progressBar.Msg("Installing App Icons Taskbar")
	// bash.Run([]string{`gext`, `-F`, `install`, `aztaskbar@aztaskbar.gitlab.com`}, "", nil)

	/* bash.Run([]string{`git`, `clone`, `https://gitlab.com/AndrewZaech/aztaskbar.git`, `~/.local/share/gnome-shell/extensions/aztaskbar@aztaskbar.gitlab.com`}, "", nil)
	bash.Run([]string{`glib-compile-schemas`, `~/.local/share/gnome-shell/extensions/aztaskbar@aztaskbar.gitlab.com/schemas/`}, "", nil) */

	bash.Run([]string{`git`, `clone`, `https://gitlab.com/AndrewZaech/aztaskbar.git`, `/usr/share/gnome-shell/extensions/aztaskbar@aztaskbar.gitlab.com`}, "", nil)
	bash.Run([]string{`glib-compile-schemas`, `/usr/share/gnome-shell/extensions/aztaskbar@aztaskbar.gitlab.com/schemas/`}, "", nil)
	bash.Run([]string{`chmod`, `-R`, `755`, `/usr/share/gnome-shell/extensions/aztaskbar@aztaskbar.gitlab.com`}, "", nil)

	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/aztaskbar@aztaskbar.gitlab.com/schemas/`, `set`, `org.gnome.shell.extensions.aztaskbar`, `favorites`, `false`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/aztaskbar@aztaskbar.gitlab.com/schemas/`, `set`, `org.gnome.shell.extensions.aztaskbar`, `show-running-apps`, `false`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/aztaskbar@aztaskbar.gitlab.com/schemas/`, `set`, `org.gnome.shell.extensions.aztaskbar`, `icon-size`, `32`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/aztaskbar@aztaskbar.gitlab.com/schemas/`, `set`, `org.gnome.shell.extensions.aztaskbar`, `panel-location`, `BOTTOM`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/aztaskbar@aztaskbar.gitlab.com/schemas/`, `set`, `org.gnome.shell.extensions.aztaskbar`, `main-panel-height`, `(true, 42)`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/aztaskbar@aztaskbar.gitlab.com/schemas/`, `set`, `org.gnome.shell.extensions.aztaskbar`, `show-panel-activities-button`, `false`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/aztaskbar@aztaskbar.gitlab.com/schemas/`, `set`, `org.gnome.shell.extensions.aztaskbar`, `clock-position-in-panel`, `RIGHT`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/aztaskbar@aztaskbar.gitlab.com/schemas/`, `set`, `org.gnome.shell.extensions.aztaskbar`, `override-panel-clock-format`, `(true, '%a, %b %d  %I:%M %p')`}, "", nil)
	theme.progressBar.Step()

	//* install open bar
	theme.progressBar.Msg("Installing Open Bar")
	// bash.Run([]string{`gext`, `-F`, `install`, `openbar@neuromorph`}, "", nil)

	/* bash.Run([]string{`git`, `clone`, `https://github.com/neuromorph/openbar.git`, `~/.local/share/gnome-shell/extensions/openbar@neuromorph`}, "", nil)
	bash.Run([]string{`glib-compile-schemas`, `~/.local/share/gnome-shell/extensions/openbar@neuromorph/schemas/`}, "", nil) */

	bash.Run([]string{`git`, `clone`, `https://github.com/neuromorph/openbar.git`, `/usr/share/gnome-shell/extensions/openbar@neuromorph`}, "", nil)
	bash.Run([]string{`glib-compile-schemas`, `/usr/share/gnome-shell/extensions/openbar@neuromorph/schemas/`}, "", nil)
	bash.Run([]string{`chmod`, `-R`, `755`, `/usr/share/gnome-shell/extensions/openbar@neuromorph`}, "", nil)

	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/openbar@neuromorph/schemas/`, `set`, `org.gnome.shell.extensions.openbar`, `bartype`, `Trilands`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/openbar@neuromorph/schemas/`, `set`, `org.gnome.shell.extensions.openbar`, `position`, `Bottom`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/openbar@neuromorph/schemas/`, `set`, `org.gnome.shell.extensions.openbar`, `height`, `32`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/openbar@neuromorph/schemas/`, `set`, `org.gnome.shell.extensions.openbar`, `margin`, `4`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/openbar@neuromorph/schemas/`, `set`, `org.gnome.shell.extensions.openbar`, `wmaxbar`, `true`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/openbar@neuromorph/schemas/`, `set`, `org.gnome.shell.extensions.openbar`, `cust-margin-wmax`, `true`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/openbar@neuromorph/schemas/`, `set`, `org.gnome.shell.extensions.openbar`, `buttonbg-wmax`, `false`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/openbar@neuromorph/schemas/`, `set`, `org.gnome.shell.extensions.openbar`, `autofg-bar`, `true`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/openbar@neuromorph/schemas/`, `set`, `org.gnome.shell.extensions.openbar`, `font`, `Sans Bold 12`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/openbar@neuromorph/schemas/`, `set`, `org.gnome.shell.extensions.openbar`, `balpha`, `0.2`}, "", nil)
	bash.Run([]string{`gsettings`, `--schemadir`, `~/.local/share/gnome-shell/extensions/openbar@neuromorph/schemas/`, `set`, `org.gnome.shell.extensions.openbar`, `neon`, `false`}, "", nil)
	theme.progressBar.Step()

	//* config vertical workspaces
	theme.progressBar.Msg("Installing Vertical Workspaces")
	// bash.Run([]string{`gext`, `-F`, `install`, `vertical-workspaces@G-dH.github.com`}, "", nil)
	installPKG("gnome-shell-extension-vertical-workspaces")

	bash.Run([]string{`gsettings`, `set`, `org.gnome.shell.extensions.vertical-workspaces`, `ws-thumbnails-position`, `0`}, "", nil)
	bash.Run([]string{`gsettings`, `set`, `org.gnome.shell.extensions.vertical-workspaces`, `show-search-entry`, `true`}, "", nil)
	bash.Run([]string{`gsettings`, `set`, `org.gnome.shell.extensions.vertical-workspaces`, `show-ws-preview-bg`, `false`}, "", nil)
	bash.Run([]string{`gsettings`, `set`, `org.gnome.shell.extensions.vertical-workspaces`, `overview-mode`, `0`}, "", nil)
	bash.Run([]string{`gsettings`, `set`, `org.gnome.shell.extensions.vertical-workspaces`, `startup-state`, `1`}, "", nil)
	bash.Run([]string{`gsettings`, `set`, `org.gnome.shell.extensions.vertical-workspaces`, `hot-corner-action`, `0`}, "", nil)
	bash.Run([]string{`gsettings`, `set`, `org.gnome.shell.extensions.vertical-workspaces`, `dash-isolate-workspaces`, `true`}, "", nil)
	bash.Run([]string{`gsettings`, `set`, `org.gnome.shell.extensions.vertical-workspaces`, `search-fuzzy`, `true`}, "", nil)
	bash.Run([]string{`gsettings`, `set`, `org.gnome.shell.extensions.vertical-workspaces`, `search-include-settings`, `true`}, "", nil)
	bash.Run([]string{`gsettings`, `set`, `org.gnome.shell.extensions.vertical-workspaces`, `notification-position`, `2`}, "", nil)
	bash.Run([]string{`gsettings`, `set`, `org.gnome.shell.extensions.vertical-workspaces`, `favorites-notify`, `0`}, "", nil)
	bash.Run([]string{`gsettings`, `set`, `org.gnome.shell.extensions.vertical-workspaces`, `ws-thumbnail-scale`, `13`}, "", nil)
	bash.Run([]string{`gsettings`, `set`, `org.gnome.shell.extensions.vertical-workspaces`, `ws-thumbnail-scale-appgrid`, `13`}, "", nil)
	bash.Run([]string{`gsettings`, `set`, `org.gnome.shell.extensions.vertical-workspaces`, `ws-preview-scale`, `95`}, "", nil)
	bash.Run([]string{`gsettings`, `set`, `org.gnome.shell.extensions.vertical-workspaces`, `ws-max-spacing`, `350`}, "", nil)
	bash.Run([]string{`gsettings`, `set`, `org.gnome.shell.extensions.vertical-workspaces`, `win-preview-height-compensation`, `50`}, "", nil)
	theme.progressBar.Step()

	// bash.Run([]string{`gext`, `-F`, `install`, `arcmenu@arcmenu.com`}, "", nil)
}
