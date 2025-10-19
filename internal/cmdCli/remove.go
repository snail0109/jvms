package cmdCli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/ystyle/jvms/internal/entity"
	"github.com/ystyle/jvms/utils/jdk"
)

func remove(cfx *entity.Config) *cli.Command {
	cmd := &cli.Command{
		Name:      "remove",
		ShortName: "rm",
		Usage:     "Remove a specific version.",
		Action: func(c *cli.Context) error {
			v := c.Args().Get(0)
			if v == "" {
				return errors.New("you should input a version, Type \"jvms list\" to see what is installed")
			}
			if jdk.IsVersionInstalled(cfx.Store, v) {
				fmt.Printf("Remove JDK %s ...\n", v)
				if cfx.CurrentJDKVersion == v {
					os.Remove(cfx.JavaHome)
				}
				dir := filepath.Join(cfx.Store, v)
				e := os.RemoveAll(dir)
				if e != nil {
					fmt.Println("Error removing jdk " + v)
					fmt.Println("Manually remove " + dir + ".")
				} else {
					fmt.Printf(" done")
				}
			} else {
				fmt.Println("jdk " + v + " is not installed. Type \"jvms list\" to see what is installed.")
			}
			return nil
		},
	}
	return cmd
}
