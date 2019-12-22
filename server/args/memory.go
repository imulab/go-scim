package args

import "github.com/urfave/cli/v2"

type Memory struct {
	MemoryDB	bool
}

func (arg *Memory) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:        "memory",
			Usage:       "Use in memory database",
			Value:       false,
			Destination: &arg.MemoryDB,
		},
	}
}