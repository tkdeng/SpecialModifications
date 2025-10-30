package main

import (
	"embed"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	bash "github.com/tkdeng/gobash"
	"github.com/tkdeng/goutil"
	"github.com/tkdeng/regex"
)

//go:embed assets/fs/*
var assetFS embed.FS

type coreInstaller struct {
	progressBar *bash.ProgressBar
	opts        *config
}

func installConfig() *config {
	opts := newConfig()

	if PM == "dnf" {
		if opts.addBool("ufw", "Would you like to install UFW (Uncomplicated Firewall)?", true) {
			fmt.Println("Using UFW...")
		} else {
			fmt.Println("Using Firewalld...")
		}
	} else {
		opts.setBool("ufw", true)
	}

	if opts.addBool("cloudflareDNS", "Would you like to use Cloudflare DNS?", true) {
		fmt.Println("Using Cloudflare DNS...")
		if opts.addBool("googleFallbackDNS", "Would you like to use Google DNS as a fallback?", true) {
			fmt.Println("Using Google Fallback DNS...")
		} else {
			fmt.Println("Using Cloudflare Fallback DNS...")
		}
	} else {
		fmt.Println("Using Google DNS...")
	}

	if !SSHClient {
		opts.addBool("disableSSH", "Would you like to disable SSH?", true)
	} else {
		opts.setBool("disableSSH", false)
	}

	time.Sleep(1 * time.Second)

	return opts
}

func installCore() {
	opts := installConfig()

	progressBar := bash.NewProgressBar("Installing")
	defer progressBar.Stop()

	core := &coreInstaller{progressBar: progressBar, opts: opts}

	progressBar.SetSize(18)

	core.countFiles("")

	if PM == "dnf" {
		progressBar.AddSize(7)
	} else if PM == "apt" {
		progressBar.AddSize(1)
	}

	if opts.bool("ufw") {
		progressBar.AddSize(1)
	}

	fmt.Println("Installing Special Modifications...")

	//* install files
	progressBar.Msg("Installing Files")
	core.files()
	core.progressBar.Step()

	//* update
	progressBar.Msg("Updating")
	update(true)
	core.progressBar.Step()

	//* install and remove apps
	progressBar.Msg("Installing Core Apps")
	installPKG("qemu-guest-agent", "tuned")
	core.progressBar.Step()

	progressBar.Msg("Removing Unneeded Apps")
	removePKG("cifs-utils", "samba-common-libs", "samba-client-libs", "libsmbclient", "libwbclient", "samba-common", "sssd-krb5-common", "sssd-ipa", "sssd-nfs-idmap", "sssd-ldap", "sssd-client", "sssd-ad", "sssd-common", "sssd-krb5", "sssd-common-pac")
	core.progressBar.Step()

	//* install ufw
	if core.opts.bool("ufw") {
		core.progressBar.Msg("Installing UFW")
		installPKG("ufw")
		bash.Run([]string{`systemctl`, `enable`, `--now`, `ufw`}, "", nil, true)

		if !SSHClient {
			bash.RunRaw(`for i in $(ufw status | wc -l); do ufw --force delete 1; done`, "", nil)
		}

		bash.Run([]string{`ufw`, `default`, `deny`, `incoming`}, "", nil)
		bash.Run([]string{`ufw`, `default`, `allow`, `outgoing`}, "", nil)
		bash.Run([]string{`ufw`, `enable`}, "", nil)

		bash.Run([]string{`systemctl`, `disable`, `--now`, `firewalld`}, "", nil)
		core.progressBar.Step()
	}

	//* secure dns
	core.progressBar.Msg("Securing DNS")
	if file, err := os.Open("/etc/systemd/resolved.conf"); err == nil {
		regex.Comp(`(?m)^#?DNSSEC=.*$`).RepFile(file, []byte(`DNSSEC=yes`), false)
		regex.Comp(`(?m)^#?DNSOverTLS=.*$`).RepFile(file, []byte(`DNSOverTLS=yes`), false)
		regex.Comp(`(?m)^#?Cache=.*$`).RepFile(file, []byte(`Cache=yes`), false)

		if core.opts.bool("cloudflareDNS") {
			regex.Comp(`(?m)^#?DNS=.*$`).RepFile(file, []byte(`DNS=1.1.1.2#security.cloudflare-dns.com 2606:4700:4700::1112#security.cloudflare-dns.com`), false)

			if core.opts.bool("googleFallbackDNS") {
				regex.Comp(`(?m)^#?FallbackDNS=.*$`).RepFile(file, []byte(`FallbackDNS=8.8.4.4#dns.google 2001:4860:4860::8844#dns.google`), false)
			} else {
				regex.Comp(`(?m)^#?FallbackDNS=.*$`).RepFile(file, []byte(`FallbackDNS=1.0.0.2#security.cloudflare-dns.com`), false)
			}

			regex.Comp(`(?m)^#?Domains=.*$`).RepFile(file, []byte(`Domains=security.cloudflare-dns.com?ip=1.1.1.2&name=Cloudflare&blockedif=zeroip dns.google`), false)
		} else {
			regex.Comp(`(?m)^#?DNS=.*$`).RepFile(file, []byte(`DNS=8.8.8.8#dns.google 2001:4860:4860::8888#dns.google`), false)
			regex.Comp(`(?m)^#?FallbackDNS=.*$`).RepFile(file, []byte(`FallbackDNS=8.8.4.4#dns.google 2001:4860:4860::8844#dns.google`), false)
			regex.Comp(`(?m)^#?Domains=.*$`).RepFile(file, []byte(`Domains=dns.google`), false)
		}

		file.Sync()
		file.Close()
	}

	bash.Run([]string{`systemctl`, `restart`, `systemd-resolved`}, "", nil)
	core.progressBar.Step()

	core.progressBar.Msg("Testing DNS")
	bash.RunRaw(`if [ "$(timeout 10 ping -c1 google.com 2>/dev/null)" = "" ]; then sed -r -i 's/^DNSSEC=.*$/DNSSEC=allow-downgrade/m' /etc/systemd/resolved.conf; systemctl restart systemd-resolved; fi`, "", nil)
	core.progressBar.Step()

	bash.RunRaw(`if [ "$(timeout 10 ping -c1 google.com 2>/dev/null)" = "" ]; then sed -r -i 's/^DNSSEC=/#DNSSEC=/m' /etc/systemd/resolved.conf; systemctl restart systemd-resolved; fi`, "", nil)
	core.progressBar.Step()

	//* install security tools
	core.progressBar.Msg("Installing Security Tools")

	//* install fail2ban
	installPKG(`fail2ban`)
	bash.RunRaw(`if ! [ -f "/etc/fail2ban/jail.local" ]; then touch "/etc/fail2ban/jail.local"; echo '[DEFAULT]' | tee -a "/etc/fail2ban/jail.local"; echo 'ignoreip = 127.0.0.1/8 ::1' | tee -a "/etc/fail2ban/jail.local"; echo 'bantime = 3600' | tee -a "/etc/fail2ban/jail.local"; echo 'findtime = 600' | tee -a "/etc/fail2ban/jail.local"; echo 'maxretry = 5' | tee -a "/etc/fail2ban/jail.local"; echo '' | tee -a "/etc/fail2ban/jail.local"; echo '[sshd]' | tee -a "/etc/fail2ban/jail.local"; echo 'enabled = true' | tee -a "/etc/fail2ban/jail.local"; fi`, "", nil)
	bash.Run([]string{`systemctl`, `enable`, `--now`, `fail2ban`}, "", nil)
	core.progressBar.Step()

	//* install clamav
	installPKG(`clamav`, `clamd`, `clamav-update`, `cronie`)
	bash.Run([]string{`systemctl`, `stop`, `clamav-freshclam`}, "", nil)
	bash.Run([]string{`freshclam`}, "", nil)
	bash.Run([]string{`systemctl`, `enable`, `--now`, `clamav-freshclam`}, "", nil)
	bash.Run([]string{`freshclam`}, "", nil)
	core.progressBar.Step()

	//* fix clamav permissions
	os.MkdirAll("/VirusScan/quarantine", 0664)
	bash.RunRaw(`if grep -R "^ScanOnAccess " "/etc/clamd.d/scan.conf"; then sed -r -i 's/^ScanOnAccess (.*)$/ScanOnAccess yes/m' /etc/clamd.d/scan.conf; else echo 'ScanOnAccess yes' | tee -a /etc/clamd.d/scan.conf; fi`, "", nil)
	bash.RunRaw(`if grep -R "^OnAccessMountPath " "/etc/clamd.d/scan.conf"; then sed -r -i 's#^OnAccessMountPath (.*)$#OnAccessMountPath /#m' /etc/clamd.d/scan.conf; else echo 'OnAccessMountPath /' | tee -a /etc/clamd.d/scan.conf; fi`, "", nil)
	bash.RunRaw(`if grep -R "^OnAccessPrevention " "/etc/clamd.d/scan.conf"; then sed -r -i 's/^OnAccessPrevention (.*)$/OnAccessPrevention no/m' /etc/clamd.d/scan.conf; else echo 'OnAccessPrevention no' | tee -a /etc/clamd.d/scan.conf; fi`, "", nil)
	bash.RunRaw(`if grep -R "^OnAccessExtraScanning " "/etc/clamd.d/scan.conf"; then sed -r -i 's/^OnAccessExtraScanning (.*)$/OnAccessExtraScanning yes/m' /etc/clamd.d/scan.conf; else echo 'OnAccessExtraScanning yes' | tee -a /etc/clamd.d/scan.conf; fi`, "", nil)
	bash.RunRaw(`if grep -R "^OnAccessExcludeUID " "/etc/clamd.d/scan.conf"; then sed -r -i 's/^OnAccessExcludeUID (.*)$/OnAccessExcludeUID 0/m' /etc/clamd.d/scan.conf; else echo 'OnAccessExcludeUID 0' | tee -a /etc/clamd.d/scan.conf; fi`, "", nil)
	bash.RunRaw(`if grep -R "^User " "/etc/clamd.d/scan.conf"; then sed -r -i 's/^User (.*)$/User root/m' /etc/clamd.d/scan.conf; else echo 'User root' | tee -a /etc/clamd.d/scan.conf; fi`, "", nil)
	core.progressBar.Step()

	//* install other security tools
	installPKG(`rkhunter`, `bleachbit`, `dnf-automatic`, `pwgen`)
	bash.RunRaw(`sed -r -i 's/^apply_updates(\s*)=(\s*)(.*)$/apply_updates\1=\2yes/m' "/etc/dnf/automatic.conf"`, "", nil)
	bash.Run([]string{`systemctl`, `enable`, `--now`, `dnf-automatic.timer`}, "", nil)
	core.progressBar.Step()

	bash.Run([]string{`rkhunter`, `--update`, `-q`}, "", nil)
	bash.Run([]string{`rkhunter`, `--propupd`, `-q`}, "", nil)

	//* schedule scans
	bash.RunRaw(`if ! [[ $(crontab -l) == *"# clamav-scan"* ]] ; then crontab -l | { cat; echo '0 2 * * * nice -n 15 clamscan && clamscan -r --bell --move="/VirusScan/quarantine" --exclude-dir="/VirusScan/quarantine" --exclude-dir="/home/$USER/.clamtk/viruses" --exclude-dir="smb4k" --exclude-dir="/run/user/$USER/gvfs" --exclude-dir="/home/$USER/.gvfs" --exclude-dir=".thunderbird" --exclude-dir=".mozilla-thunderbird" --exclude-dir=".evolution" --exclude-dir="Mail" --exclude-dir="kmail" --exclude-dir="^/sys" / # clamav-scan'; } | crontab -; fi`, "", nil)
	core.progressBar.Step()

	//todo: add scheduled scans to virus scanning app
	// also make new virus scanning app that uses clamav

	if PM == "dnf" {
		//* install rpm repos
		core.progressBar.Msg("Installing RPM repos")
		bash.RunRaw(`dnf -y install https://download1.rpmfusion.org/free/fedora/rpmfusion-free-release-$(rpm -E %fedora).noarch.rpm`, "", nil)
		bash.RunRaw(`dnf -y install https://download1.rpmfusion.org/nonfree/fedora/rpmfusion-nonfree-release-$(rpm -E %fedora).noarch.rpm`, "", nil)
		installPKG(`fedora-workstation-repositories`)
		bash.Run([]string{`fedora-third-party`, `enable`}, "", nil)
		bash.Run([]string{`fedora-third-party`, `refresh`}, "", nil)
		bash.Run([]string{`dnf`, `-y`, `groupupdate`, `core`}, "", nil)
		core.progressBar.Step()

		bash.Run([]string{`dnf`, `clean`, `all`}, "", nil)
		bash.Run([]string{`dnf`, `-y`, `autoremove`}, "", nil)
		bash.Run([]string{`dnf`, `-y`, `distro-sync`}, "", nil)
		core.progressBar.Step()

		//* install flatpak
		core.progressBar.Msg("Installing flatpak")
		installPKG(`flatpak`)
		bash.Run([]string{`flatpak`, `remote-add`, `--if-not-exists`, `flathub`, `https://flathub.org/repo/flathub.flatpakrepo`}, "", nil)
		// bash.Run([]string{`flatpak`, `update`, `-y`, `--noninteractive`}, "", nil)
		bash.Run([]string{`flatpak`, `install`, `-y`, `flathub`, `com.github.tchx84.Flatseal`}, "", nil)
		core.progressBar.Step()

		//* install snap
		core.progressBar.Msg("Installing snap")
		installPKG(`snap`)
		bash.Run([]string{`ln`, `-s`, `/var/lib/snapd/snap /snap`}, "", nil)
		bash.Run([]string{`systemctl`, `enable`, `snapd`, `--now`}, "", nil)
		bash.Run([]string{`snap`, `refresh`}, "", nil) // fix: not seeded yet will trigger and fix itself for the next command
		bash.Run([]string{`snap`, `install`, `core`}, "", nil)
		bash.Run([]string{`snap`, `refresh`, `core`}, "", nil)
		bash.Run([]string{`snap`, `refresh`}, "", nil)
		core.progressBar.Step()

		bash.Run([]string{`dnf`, `clean`, `all`}, "", nil)
		bash.Run([]string{`dnf`, `-y`, `autoremove`}, "", nil)
		bash.Run([]string{`dnf`, `-y`, `distro-sync`}, "", nil)
		core.progressBar.Step()

		core.progressBar.Msg("Updating multimedia codecs")
		bash.Run([]string{`dnf`, `-y`, `--skip-broken`, `install`, `@multimedia`}, "", nil)
		bash.Run([]string{`dnf`, `-y`, `groupupdate`, `multimedia`, `--setop=install_weak_deps=False`, `--exclude=PackageKit-gstreamer-plugin`, `--skip-broken`}, "", nil)
		bash.Run([]string{`dnf`, `-y`, `groupupdate`, `sound-and-video`}, "", nil)
		bash.Run([]string{`dnf`, `-y`, `--allowerasing`, `install`, `ffmpeg`}, "", nil)
		core.progressBar.Step()

		installPKG(`libwebp`, `libwebp-devel`)
		installPKG(`webp-pixbuf-loader`)
		core.progressBar.Step()
	} else if PM == "apt" {
		//todo: install apt repos (flatpak and snap)
	}

	//* disable startups
	core.progressBar.Msg("Disabling Time Wasting Programs")
	bash.Run([]string{`systemctl`, `disable`, `accounts-daemon.service`}, "", nil) // is a potential securite risk
	bash.Run([]string{`systemctl`, `disable`, `debug-shell.service`}, "", nil)     // opens a giant security hole
	removePKG(`dmraid`, `device-mapper-multipath`)
	core.progressBar.Step()

	//* install programming languages
	if PM == "apt" {
		//* install ubuntu extras
		core.progressBar.Msg("Installing Ubuntu Extras")
		installPKG(`ubuntu-restricted-extras`)
		core.progressBar.Step()
	}

	//* install python
	core.progressBar.Msg("Installing Python")
	installPKG(`python`, `python3`, `python-pip`, `python3-pip`)
	core.progressBar.Step()

	//* install c
	core.progressBar.Msg("Installing C")
	installPKG(`gcc-c++`, `make`, `gcc`)
	core.progressBar.Step()

	//* install java
	core.progressBar.Msg("Making Java")
	if PM == "apt" {
		// installPKG(`openjdk-8-jre`, `openjdk-8-jdk`, `openjdk-11-jre`, `openjdk-11-jdk`)
		installPKG(`openjdk-8-jre`, `openjdk-8-jdk`, `openjdk-25-jre`, `openjdk-25-jdk`)
	} else {
		// installPKG(`java-1.8.0-openjdk`, `java-11-openjdk`, `java-latest-openjdk`)
		installPKG(`java-1.8.0-openjdk`, `java-25-openjdk`, `java-latest-openjdk`)
	}
	core.progressBar.Step()

	//* install git and node
	core.progressBar.Msg("Installing Git and Node")
	installPKG(`git`, `nodejs`, `npm`)
	core.progressBar.Step()

	//* install golang
	core.progressBar.Msg("Installing Go")
	installPKG(`golang`, `pcre-devel`) //todo: lookup golang install for apt
	core.progressBar.Step()

	//* docker
	core.progressBar.Msg("Installing Docker")
	if PM == "dnf" {
		installPKG(`dnf-plugins-core`)
		bash.Run([]string{`dnf`, `config-manager`, `--add-repo`, `https://download.docker.com/linux/fedora/docker-ce.repo`}, "", nil)
		installPKG(`docker-ce`, `docker-ce-cli`, `containerd.io`, `docker-buildx-plugin`, `docker-compose-plugin`)
		installPKG(`docker`)
		bash.Run([]string{`systemctl`, `enable`, `docker`, `--now`}, "", nil)
	} else if PM == "apt" {
		//todo: lookup apt version of docker install
	}
	core.progressBar.Step()

	//* install common apps
	core.progressBar.Msg("Installing Common Packages")

	//todo: lookup apt equivalent for some packages
	installPKG(`nano`, `micro`, `neofetch`, `btrfs-progs`, `lvm2`, `xfsprogs`, `ntfs-3g`, `ntfsprogs`, `exfatprogs`, `udftools`, `p7zip`, `p7zip-plugins`, `hplip`, `hplip-gui`, `inotify-tools`, `guvcview`, `selinux-policy-devel`)
	bash.Run([]string{`systemctl`, `enable`, `fstrim.timer`, `--now`}, "", nil)
	bash.Run([]string{`systemctl`, `enable`, `systemd-oomd.service`, `--now`}, "", nil)
	bash.Run([]string{`systemctl`, `enable`, `sshd.socket`, `--now`}, "", nil)
	core.progressBar.Step()

	//* install fonts
	installPKG(`jetbrains-mono-fonts`)
	core.progressBar.Step()

	//* update
	core.progressBar.Msg("Updating")
	update(true)
	core.progressBar.Step()
}

func (core *coreInstaller) files() {
	filePerms := map[string]os.FileMode{}

	if buf, err := os.ReadFile("assets/fs/.perms.json"); err == nil {
		if json, err := goutil.JSON.Parse(buf); err == nil {
			for key, val := range json {
				switch v := val.(type) {
				case string:
					if perm, err := strconv.ParseUint(v, 8, 32); err == nil {
						filePerms[key] = os.FileMode(perm)
					}
				case float64:
					filePerms[key] = os.FileMode(uint32(v))
				}
			}
		}
	}

	core.installFiles(&filePerms, "", 0755)
}

func (core *coreInstaller) countFiles(dir string) {
	if files, err := assetFS.ReadDir("assets/fs" + dir); err == nil {
		for _, file := range files {
			path := dir + "/" + file.Name()
			if strings.HasPrefix(path, "/.") {
				continue
			}

			if file.IsDir() {
				core.countFiles(path)
				continue
			}

			core.progressBar.AddSize(1)
		}
	}
}

func (core *coreInstaller) installFiles(filePerms *map[string]os.FileMode, dir string, dirPerm os.FileMode) {
	if files, err := assetFS.ReadDir("assets/fs" + dir); err == nil {
		for _, file := range files {
			path := dir + "/" + file.Name()
			if strings.HasPrefix(path, "/.") {
				continue
			}

			if file.IsDir() {
				if dir, err := os.Stat(path); err == nil {
					dirPerm = dir.Mode().Perm()
				}

				core.installFiles(filePerms, path, dirPerm)
				continue
			}

			if buf, err := assetFS.ReadFile("assets/fs" + path); err == nil {
				var perm os.FileMode = 0644
				if val, ok := (*filePerms)[path]; ok {
					perm = val
				}

				os.MkdirAll(dir, dirPerm)
				os.WriteFile(path, buf, perm)

				core.progressBar.Step()
			}
		}
	}
}
