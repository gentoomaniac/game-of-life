package main

import (
	"github.com/alecthomas/kong"

	gocli "github.com/gentoomaniac/game-of-life/pkg/cli"
	"github.com/gentoomaniac/game-of-life/pkg/gameoflife"
	"github.com/gentoomaniac/game-of-life/pkg/logging"
)

var (
	version = "unknown"
	commit  = "unknown"
	binName = "unknown"
	builtBy = "unknown"
	date    = "unknown"
)

var cli struct {
	logging.LoggingConfig

	Width       int      `help:"World width" default:"320"`
	Height      int      `help:"World height" default:"240"`
	Scale       int      `help:"Graphic scale of the world" default:"4"`
	Density     float64  `help:"initial amount of lifeforms" default:"0.25"`
	HighLife    bool     `help:"enable the HighLife rule, born with 3 or 6" default:"false"`
	Seed        *int64   `help:"seed for randomnes"`
	Tps         int      `help:"simulation speed in ticks per seconf" default:"10"`
	RandomRaise *float64 `help:"randomly bring some alive"`

	Version gocli.VersionFlag `short:"V" help:"Display version."`
}

func main() {
	ctx := kong.Parse(&cli, kong.UsageOnError(), kong.Vars{
		"version": version,
		"commit":  commit,
		"binName": binName,
		"builtBy": builtBy,
		"date":    date,
	})
	logging.Setup(&cli.LoggingConfig)

	gameoflife.Show(cli.Width, cli.Height, cli.Scale, cli.Density, cli.HighLife, cli.Seed, cli.Tps, cli.RandomRaise)
	ctx.Exit(0)
}
