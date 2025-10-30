package main

import bash "github.com/tkdeng/gobash"

type config struct {
	values map[string]string
}

func newConfig() *config {
	return &config{values: map[string]string{}}
}

func (c *config) addBool(key string, msg string, def bool) bool {
	if AssumeYes {
		if def {
			c.values[key] = "true"
		} else {
			c.values[key] = "false"
		}

		return def
	}

	if bash.InputYN(msg, def) {
		c.values[key] = "true"
		return true
	} else {
		c.values[key] = "false"
		return false
	}
}

func (c *config) setBool(key string, value bool) {
	if value {
		c.values[key] = "true"
	} else {
		c.values[key] = "false"
	}
}

func (c *config) bool(key string) bool {
	return c.values[key] == "true"
}

func (c *config) addValue(key string, msg string, def string) string {
	if AssumeYes {
		c.values[key] = def
		return def
	}

	val := bash.InputText(msg)
	if val != "" {
		c.values[key] = val
		return val
	}

	c.values[key] = def
	return def
}

func (c *config) setValue(key string, value string) {
	c.values[key] = value
}

func (c *config) value(key string) string {
	return c.values[key]
}

func update(cleanup ...bool) {
	switch PM {
	case "apt":
		bash.Run([]string{`apt`, `-y`, `update`}, "", nil, true)
		bash.Run([]string{`apt`, `-y`, `upgrade`}, "", nil, true)

		if len(cleanup) != 0 && cleanup[0] {
			bash.Run([]string{`dpkg`, `--configure`, `-a`}, "", nil, true)
			bash.Run([]string{`apt`, `-y`, `-f`, `install`}, "", nil, true)
			bash.Run([]string{`apt`, `-y`, `autoremove`, `--purge`}, "", nil, true)
			bash.Run([]string{`apt`, `-y`, `autoclean`}, "", nil, true)
			bash.Run([]string{`apt`, `-y`, `clean`}, "", nil, true)

			if hasNalaPM {
				bash.Run([]string{`nala`, `update`}, "", nil, true)
				bash.Run([]string{`nala`, `upgrade`, `-y`}, "", nil, true)
				bash.Run([]string{`nala`, `install`, `-y`}, "", nil, true)
				bash.Run([]string{`nala`, `autoremove`, `-y`}, "", nil, true)
				bash.Run([]string{`nala`, `clean`}, "", nil, true)
			}
		}
	case "dnf":
		bash.Run([]string{`dnf`, `-y`, `update`}, "", nil, true)

		if len(cleanup) != 0 && cleanup[0] {
			bash.Run([]string{`dnf`, `clean`, `all`}, "", nil, true)
			bash.Run([]string{`dnf`, `-y`, `autoremove`}, "", nil, true)
			bash.Run([]string{`dnf`, `-y`, `distro-sync`}, "", nil, true)
		}
	}
}

func installPKG(pkg ...string) {
	switch PM {
	case "apt":
		if hasNalaPM {
			bash.Run(append([]string{`nala`, `install`, `-y`}, pkg...), "", nil, true)
		} else {
			bash.Run(append([]string{`apt`, `-y`, `install`}, pkg...), "", nil, true)
		}
	case "dnf":
		bash.Run(append([]string{`dnf`, `-y`, `install`}, pkg...), "", nil, true)
	}
}

func removePKG(pkg ...string) {
	//todo: uninstall packages
	switch PM {
	case "apt":
		if hasNalaPM {
			bash.Run(append([]string{`nala`, `remove`, `-y`}, pkg...), "", nil, true)
		} else {
			bash.Run(append([]string{`apt`, `-y`, `remove`}, pkg...), "", nil, true)
		}
	case "dnf":
		bash.Run(append([]string{`dnf`, `-y`, `remove`}, pkg...), "", nil, true)
	}
}

func hasPKG(pkg ...string) bool {
	for _, name := range pkg {
		switch PM {
		case "apt":
			out, err := bash.RunRaw(`dpkg-query -W --showformat='${Status}\n' "`+name+`" 2>/dev/null|grep "install ok installed"`, "", nil)
			if err != nil || len(out) == 0 {
				return false
			}
		case "dnf":
			out, err := bash.Run([]string{`rpm`, `-q`, name}, "", nil)
			if err != nil || len(out) == 0 {
				return false
			}
		}
	}

	return true
}
