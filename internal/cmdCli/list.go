package cmdCli

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/snail0109/jvms/internal/entity"
	"github.com/snail0109/jvms/utils/jdk"
)

func list(cfx *entity.Config) *cli.Command {
	cmd := &cli.Command{
		Name:      "list",
		ShortName: "ls",
		Usage:     "List current JDK installations.",
		Action: func(c *cli.Context) error {
			fmt.Println("Installed jdk (* marks in use):")
			v := jdk.GetInstalled(cfx.Store)
			for i, version := range v {
				str := ""
				if cfx.CurrentJDKVersion == version {
					str = fmt.Sprintf("%s  * %d) %s", str, i+1, version)
				} else {
					str = fmt.Sprintf("%s    %d) %s", str, i+1, version)
				}
				fmt.Printf(str + "\n")
			}
			if len(v) == 0 {
				fmt.Println("No installations recognized.")
			}
			return nil
		},
	}
	return cmd
}
