package cmdCli

import (
	"github.com/codegangsta/cli"
	"github.com/snail0109/jvms/internal/entity"
)

func use(cfx *entity.Config) *cli.Command {
	cmd := &cli.Command{
		Name:      "use",
		ShortName: "u",
		Usage:     "Switch to use the specified version or index number.",
		Action:    switchFunc(*cfx),
	}
	return cmd
}
