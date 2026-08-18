# TODO

consider installing thonny ide (or something else) for a lighter alternative to vscode

```shell
sudo dnf install python3-tkinter thonny
```

or vscodium

```shell
# Add the repository key
sudo rpm --import https://gitlab.com/paulcarroty/vscodium-deb-rpm-repo/raw/master/pub.gpg

# Add the DNF repository configuration
printf "[vscodium]\nname=vscodium\nbaseurl=https://download.vscodium.com/rpms/\nenabled=1\ngpgcheck=1\ngpgkey=https://gitlab.com/paulcarroty/vscodium-deb-rpm-repo/raw/master/pub.gpg\n" | sudo tee /etc/yum.repos.d/vscodium.repo

# Install the editor
sudo dnf install codium
```

or zed

```shell
sudo dnf copr enable pgdev/zed
sudo dnf install zed
```

