package main

import (
	"github.com/openinx/huker"
	"github.com/qiniu/log"
	"github.com/urfave/cli"
	"os"
)

func main() {

	app := cli.NewApp()

	cfgRootDir := "./conf"
	agentRootDir := "/tmp/test"
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
			Usage: "Bootstrap a cluster of specific project, or jobs, or tasks",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "project",
					Usage: "project name, such as hdfs, yarn, zookeeper, hbase, etc",
				},
				cli.StringFlag{
					Name:  "cluster",
					Usage: "cluster name",
				},
				cli.StringFlag{
					Name:  "job",
					Usage: "job name of the project, for hbase, the job will be master, regionserver, canary etc.",
				},
				cli.StringFlag{
					Name:   "task",
					Hidden: true,
					Usage:  "task id of given service and job, type: integer",
				},
			},
			Action: hShell.Bootstrap,
		},
		{
			Name:  "show",
			Usage: "Show jobs of a given service",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "project",
					Usage: "project name, such as hdfs, yarn, zookeeper, hbase, etc",
				},
				cli.StringFlag{
					Name:  "cluster",
					Usage: "cluster name",
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
			Action: hShell.Show,
		},
		{
			Name:  "start",
			Usage: "Start a job",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "project",
					Usage: "project name, such as hdfs, yarn, zookeeper, hbase, etc",
				},
				cli.StringFlag{
					Name:  "cluster",
					Usage: "cluster name",
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
			Action: hShell.Start,
		},
		{
			Name:  "cleanup",
			Usage: "Cleanup a job",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "project",
					Usage: "project name, such as hdfs, yarn, zookeeper, hbase, etc",
				},
				cli.StringFlag{
					Name:  "cluster",
					Usage: "cluster name",
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
			Action: hShell.Cleanup,
		},
		{
			Name:  "rolling_update",
			Usage: "Rolling update a job",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "project",
					Usage: "project name, such as hdfs, yarn, zookeeper, hbase, etc",
				},
				cli.StringFlag{
					Name:  "cluster",
					Usage: "cluster name",
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
			Action: hShell.RollingUpdate,
		},
		{
			Name:  "restart",
			Usage: "Restart a job",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "project",
					Usage: "project name, such as hdfs, yarn, zookeeper, hbase, etc",
				},
				cli.StringFlag{
					Name:  "cluster",
					Usage: "cluster name",
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
			Action: hShell.Restart,
		},
		{
			Name:  "stop",
			Usage: "Stop a job",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "project",
					Usage: "project name, such as hdfs, yarn, zookeeper, hbase, etc",
				},
				cli.StringFlag{
					Name:  "cluster",
					Usage: "cluster name",
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
			Action: hShell.Stop,
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
