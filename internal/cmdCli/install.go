package cmdCli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/snail0109/jvms/internal/entity"
	"github.com/snail0109/jvms/utils/file"
	"github.com/snail0109/jvms/utils/jdk"
	"github.com/snail0109/jvms/utils/web"
)

func install(cfx *entity.Config) *cli.Command {
	cmd := &cli.Command{
		Name:      "install",
		ShortName: "i",
		Usage:     "Install available remote jdk",
		Action: func(c *cli.Context) error {
			if cfx.Proxy != "" {
				web.SetProxy(cfx.Proxy)
			}
			v := c.Args().Get(0)
			if v == "" {
				return errors.New("invalid version., Type \"jvms rls\" to see what is available for install")
			}

			if jdk.IsVersionInstalled(cfx.Store, v) {
				fmt.Println("Version " + v + " is already installed.")
				return nil
			}
			versions, err := getJdkVersions(cfx)
			if err != nil {
				return err
			}

			if !file.Exists(cfx.Download) {
				os.MkdirAll(cfx.Download, 0777)
			}
			if !file.Exists(cfx.Store) {
				os.MkdirAll(cfx.Store, 0777)
			}

			for _, version := range versions {
				if version.Version == v {
					dlzipfile, success := web.GetJDK(cfx.Download, v, version.Url)
					if success {
						fmt.Printf("Installing JDK %s ...\n", v)

						// Extract jdk to the temp directory
						jdktempfile := filepath.Join(cfx.Download, fmt.Sprintf("%s_temp", v))
						if file.Exists(jdktempfile) {
							err := os.RemoveAll(jdktempfile)
							if err != nil {
								panic(err)
							}
						}
						err := file.Unzip(dlzipfile, jdktempfile)
						if err != nil {
							return fmt.Errorf("unzip failed: %w", err)
						}

						// Copy the jdk files to the installation directory
						temJavaHome := getJavaHome(jdktempfile)
						err = os.Rename(temJavaHome, filepath.Join(cfx.Store, v))
						if err != nil {
							return fmt.Errorf("unzip failed: %w", err)
						}

						// Remove the temp directory
						// may consider keep the temp files here
						os.RemoveAll(jdktempfile)
						fmt.Printf("Installation completedly succesfully. Use: jvms switch %v, if you'd like to use this version", v)
					} else {
						fmt.Println("Could not download JDK " + v + " executable.")
					}
					return nil
				}
			}
			return errors.New("invalid version., Type \"jvms rls\" to see what is available for install")
		},
	}
	return cmd
}
