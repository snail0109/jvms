package cmdCli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/codegangsta/cli"
	"github.com/snail0109/jvms/internal/entity"
	"github.com/snail0109/jvms/utils/file"
	"github.com/snail0109/jvms/utils/jdk"
)

func switch_(cfx *entity.Config) *cli.Command {
	cmd := &cli.Command{
		Name:      "switch",
		ShortName: "s",
		Usage:     "Switch to use the specified version or index number.",
		Action:    switchFunc(*cfx),
	}
	return cmd
}

func switchFunc(cfx entity.Config) func(*cli.Context) error {
	return func(c *cli.Context) error {
		v := c.Args().Get(0)
		if v == "" {
			return errors.New("you should input a version or index number, Type \"jvms list\" to see what is installed")
		}

		// Check if input is a number (index)
		index, err := strconv.Atoi(v)
		if err == nil && index > 0 {
			// Input is a valid number, get the list of installed JDKs
			installedJDKs := jdk.GetInstalled(cfx.Store)
			if len(installedJDKs) == 0 {
				return errors.New("no JDK installations found")
			}

			if index > len(installedJDKs) {
				return fmt.Errorf("invalid index: %d, should be between 1 and %d", index, len(installedJDKs))
			}

			v = installedJDKs[index-1]
			fmt.Printf("Using index %d to select JDK %s\n", index, v)
		}

		if !jdk.IsVersionInstalled(cfx.Store, v) {
			fmt.Printf("jdk %s is not installed. ", v)
			return nil
		}
		// Create or update the symlink
		if file.Exists(cfx.JavaHome) {
			err := os.Remove(cfx.JavaHome)
			if err != nil {
				return errors.New("Switch jdk failed, please manually remove " + cfx.JavaHome)
			}
		}
		cmd := exec.Command("cmd", "/C", "setx", "JAVA_HOME", cfx.JavaHome, "/M")
		err = cmd.Run()
		if err != nil {
			return errors.New("set Environment variable `JAVA_HOME` failure: Please run as admin user")
		}
		err = os.Symlink(filepath.Join(cfx.Store, v), cfx.JavaHome)
		if err != nil {
			return errors.New("Switch jdk failed, " + err.Error())
		}
		fmt.Println("Switch success.\nNow using JDK " + v)
		cfx.CurrentJDKVersion = v
		return nil
	}
}
