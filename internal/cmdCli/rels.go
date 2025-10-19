package cmdCli

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/ystyle/jvms/internal/entity"
	"github.com/ystyle/jvms/utils/web"
)

func rls(cfx *entity.Config) *cli.Command {
	cmd := &cli.Command{
		Name:  "rls",
		Usage: "Show a list of versions available for download. ",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "a",
				Usage: "list all the version",
			},
		},
		Action: func(c *cli.Context) error {
			if cfx.Proxy != "" {
				web.SetProxy(cfx.Proxy)
			}
			versions, err := getJdkVersions(cfx)
			if err != nil {
				return err
			}
			for i, version := range versions {
				fmt.Printf("    %d) %s\n", i+1, version.Version)
				if !c.Bool("a") && i >= 9 {
					fmt.Println("\nuse \"jvm rls -a\" show all the versions ")
					break
				}
			}
			if len(versions) == 0 {
				fmt.Println("No availabled jdk veriosn for download.")
			}

			fmt.Printf("\nFor a complete list, visit %s\n", cfx.Originalpath)
			return nil
		},
	}
	return cmd
}
