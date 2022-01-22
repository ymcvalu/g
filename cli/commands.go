package cli

import "github.com/urfave/cli"

var (
	commands = []cli.Command{
		{
			Name:      "ls",
			Usage:     "List installed versions",
			UsageText: "gvm ls",
			Action:    list,
		},
		{
			Name:      "lsall",
			Usage:     "List all versions available for install",
			UsageText: "gvm lsall [stable|archived|unstable]",
			Action:    listRemote,
		},
		{
			Name:      "use",
			Usage:     "Switch to specified version",
			UsageText: "gvm use <version>",
			Action:    use,
		},
		{
			Name:      "install",
			Usage:     "Download and install a version",
			UsageText: "gvm install <version>",
			Action:    install,
		},
		{
			Name:      "uninstall",
			Usage:     "Uninstall a version",
			UsageText: "gvm uninstall <version>",
			Action:    uninstall,
		},
		{
			Name:      "clean",
			Usage:     "Remove files from the package download directory",
			UsageText: "gvm clean",
			Action:    clean,
		},
	}
)
