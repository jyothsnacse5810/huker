package main

import (
	"fmt"
	"github.com/openinx/huker"
	"github.com/qiniu/log"
	"github.com/urfave/cli"
	"os"
	"strings"
)

func main() {

	app := cli.NewApp()

	cfgRootDir := "/Users/openinx/gopath/src/github.com/openinx/huker/conf"
	agentRootDir := "/Users/openinx/test"
	pkgServerAddress := "http://127.0.0.1:4000"
	supervisorPort := 9001

	hShell, err := huker.NewHukerShell(cfgRootDir, agentRootDir, pkgServerAddress, supervisorPort)

	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	app.Commands = []cli.Command{
		{
			Name:  "bootstrap",
			Usage: "Bootstrap a service",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "cluster",
					Usage: "cluster name",
				},
				cli.StringFlag{
					Name:  "service",
					Usage: "service name",
				},
				cli.StringFlag{
					Name:  "job",
					Usage: "job name",
				},
				cli.StringFlag{
					Name:   "task",
					Hidden: true,
					Usage:  "task id of given service and job",
				},
			},
			Action: hShell.Bootstrap,
		},
		{
			Name:  "show",
			Usage: "Show jobs of a given service",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "service",
					Usage: "service name",
				},
				cli.StringFlag{
					Name:  "job",
					Usage: "job name",
				},
				cli.StringFlag{
					Name:   "task",
					Hidden: true,
					Usage:  "task id of given service and job",
				},
			},
			Action: func(c *cli.Context) error {
				fmt.Println("Show jobs of a given service: ", strings.Join(c.Args(), " "))
				return nil
			},
		},
		{
			Name:  "Start",
			Usage: "Start a job",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "service",
					Usage: "service name",
				},
				cli.StringFlag{
					Name:  "job",
					Usage: "job name",
				},
				cli.StringFlag{
					Name:   "task",
					Hidden: true,
					Usage:  "task id of given service and job",
				},
			},
			Action: func(c *cli.Context) error {
				fmt.Println("Start a job: ", strings.Join(c.Args(), " "))
				return nil
			},
		},
		{
			Name:  "Start",
			Usage: "Start a job",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "service",
					Usage: "service name",
				},
				cli.StringFlag{
					Name:  "job",
					Usage: "job name",
				},
				cli.StringFlag{
					Name:   "task",
					Hidden: true,
					Usage:  "task id of given service and job",
				},
			},
			Action: func(c *cli.Context) error {
				fmt.Println("Start a job: ", strings.Join(c.Args(), " "))
				return nil
			},
		},
		{
			Name:  "cleanup",
			Usage: "Cleanup a job",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "service",
					Usage: "service name",
				},
				cli.StringFlag{
					Name:  "job",
					Usage: "job name",
				},
				cli.StringFlag{
					Name:   "task",
					Hidden: true,
					Usage:  "task id of given service and job",
				},
			},
			Subcommands: []cli.Command{
				{
					Name:  "add",
					Usage: "add a new template",
					Action: func(c *cli.Context) error {
						fmt.Println("new task template: ", c.Args().First())
						return nil
					},
				},
				{
					Name:  "remove",
					Usage: "remove an existing template",
					Action: func(c *cli.Context) error {
						fmt.Println("removed task template: ", c.Args().First())
						return nil
					},
				},
			},
		},
		{
			Name:  "rolling_update",
			Usage: "Rolling update a job",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "service",
					Usage: "service name",
				},
				cli.StringFlag{
					Name:  "job",
					Usage: "job name",
				},
				cli.StringFlag{
					Name:   "task",
					Hidden: true,
					Usage:  "task id of given service and job",
				},
			},
			Action: func(c *cli.Context) error {
				fmt.Println("Rolling update a job: ", strings.Join(c.Args(), " "))
				return nil
			},
		},
		{
			Name:  "restart",
			Usage: "Restart a job",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "service",
					Usage: "service name",
				},
				cli.StringFlag{
					Name:  "job",
					Usage: "job name",
				},
				cli.StringFlag{
					Name:   "task",
					Hidden: true,
					Usage:  "task id of given service and job",
				},
			},
			Action: func(c *cli.Context) error {
				fmt.Println("restart a job: ", strings.Join(c.Args(), " "))
				return nil
			},
		},
		{
			Name:  "stop",
			Usage: "Stop a job",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "service",
					Usage: "service name",
				},
				cli.StringFlag{
					Name:  "job",
					Usage: "job name",
				},
				cli.StringFlag{
					Name:   "task",
					Hidden: true,
					Usage:  "task id of given service and job",
				},
			},
			Action: func(c *cli.Context) error {
				fmt.Println("stop a job: ", strings.Join(c.Args(), " "))
				return nil
			},
		},
		{
			Name:  "start-agent",
			Usage: "Start Huker Agent",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "dir, d",
					Value: ".",
					Usage: "Root directory of huker agent.",
				},
				cli.IntFlag{
					Name:  "port, p",
					Value: 9001,
					Usage: "Port to listen for huker agent.",
				},
				cli.StringFlag{
					Name:  "file, f",
					Value: "./supervisor.db",
					Usage: "file to store process meta.",
				},
			},
			Action: func(c *cli.Context) error {
				dir := c.String("dir")
				port := c.Int("port")
				file := c.String("file")
				s, err := huker.NewSupervisor(dir, port, file)
				if err != nil {
					log.Error(err)
					return err
				}
				if err = s.Start(); err != nil {
					log.Error(err)
					return err
				}
				return err
			},
		},
		{
			Name:  "start-pkg-manager",
			Usage: "Start Huker Package Manager",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "port, p",
					Value: 4000,
					Usage: "Port to listen for huker package manager",
				},
				cli.StringFlag{
					Name:  "dir, d",
					Value: "./lib",
					Usage: "Libaray directory of huker package manager for downloading package",
				},
				cli.StringFlag{
					Name:  "conf, c",
					Value: "./conf/pkg.yaml",
					Usage: "Configuration file of huker package manager",
				},
			},
			Action: func(c *cli.Context) error {
				port := c.Int("port")
				dir := c.String("dir")
				conf := c.String("conf")
				p, err := huker.NewPackageServer(port, dir, conf)
				if err != nil {
					log.Error(err)
					return err
				}
				if err = p.Start(); err != nil {
					log.Error(err)
					return err
				}
				return err
			},
		},
	}

	app.Run(os.Args)
}
