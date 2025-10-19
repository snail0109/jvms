package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/tucnak/store"
	"github.com/ystyle/jvms/internal/cmdCli"
	"github.com/ystyle/jvms/internal/entity"
	"github.com/ystyle/jvms/utils/file"
	"github.com/ystyle/jvms/utils/web"
)

var version = "2.1.0"

const (
	defaultOriginalpath = "https://raw.githubusercontent.com/ystyle/jvms/new/jdkdlindex.json"
)

var cfx entity.Config

func main() {
	app := cli.NewApp()
	app.Name = "jvms"
	app.Usage = `JDK Version Manager (JVMS) for Windows`
	app.Version = version
	app.CommandNotFound = commandNotFound
	app.Commands = commands()

	app.Before = startup
	app.After = shutdown
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err.Error())
		os.Exit(1)
	}
}

func commands() []cli.Command {
	cmds := cmdCli.Commands(&cmdCli.CommandParams{
		DefaultOriginalPath: defaultOriginalpath,
		Config:              &cfx,
	})
	return cmds
}

func commandNotFound(c *cli.Context, command string) {
	log.Fatal("Command Not Found")
}

func startup(c *cli.Context) error {
	store.Register(
		"json",
		func(v interface{}) ([]byte, error) {
			return json.MarshalIndent(v, "", "    ")
		},
		json.Unmarshal)

	store.Init("jvms")
	if err := store.Load("jvms.json", &cfx); err != nil {
		return errors.New("failed to load the config:" + err.Error())
	}
	s := file.GetCurrentPath()
	cfx.Store = filepath.Join(s, "store")

	cfx.Download = filepath.Join(s, "download")
	if cfx.Originalpath == "" {
		cfx.Originalpath = defaultOriginalpath
	}
	if cfx.Proxy != "" {
		web.SetProxy(cfx.Proxy)
	}
	return nil
}

func shutdown(c *cli.Context) error {
	if err := store.Save("jvms.json", &cfx); err != nil {
		return errors.New("failed to save the config:" + err.Error())
	}
	return nil
}
