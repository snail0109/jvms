package cmdCli

import (
	"github.com/codegangsta/cli"
	"github.com/ystyle/jvms/internal/entity"
)

type CommandParams struct {
	DefaultOriginalPath string
	Config              *entity.Config
}

func Commands(cp *CommandParams) []cli.Command {
	cmds := []cli.Command{
		*init_(cp.DefaultOriginalPath, cp.Config),
		*list(cp.Config),
		*install(cp.Config),
		*switch_(cp.Config),
		*use(cp.Config),
		*remove(cp.Config),
		*rls(cp.Config),
		*proxy(cp.Config),
	}
	return cmds
}
