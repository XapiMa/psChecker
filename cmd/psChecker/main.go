package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/mitchellh/cli"
	"github.com/xapima/psmonitor"
)

// ShowCommand show the current process list
type ShowCommand struct{}

// Run showing the current process list
func (c *ShowCommand) Run(args []string) int {
	flags := flag.NewFlagSet("show", flag.ExitOnError)
	flags.Usage = func() { fmt.Fprintf(os.Stderr, "%s\n", c.Help()) }
	targetTypesString := flags.String("t", "exec|cmd|open|user|pid", "Display items separated by '|'")
	outputPath := flags.String("o", "", "path/to/output. default stdout")
	verbose := flags.Bool("v", false, "show Verbose")
	if err := flags.Parse(args); err != nil {
		fmt.Println(err)
		return 1
	}
	if !*verbose {
		log.SetOutput(ioutil.Discard)
	} else {
		log.SetOutput(os.Stderr)
	}

	shower, err := psmonitor.NewShower(*targetTypesString, *outputPath)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	if err := shower.Show(); err != nil {
		fmt.Println(err)
		return 1
	}

	return 0
}

// Synopsis show usagelist without a subcommand
func (c *ShowCommand) Synopsis() string {
	return "Display current process list"
}

// Help show usage with a subcommand
func (c *ShowCommand) Help() string {
	return "Usage: psMonitor show [-t typeOfDisplayItems] [-o path/to/output] [-v]"
}

// MonitorCommand monitors the process
type MonitorCommand struct{}

// Run monitoring the process
func (c *MonitorCommand) Run(args []string) int {
	flags := flag.NewFlagSet("monitor", flag.ExitOnError)
	flags.Usage = func() { fmt.Fprintf(os.Stderr, "%s\n", c.Help()) }
	whitelistPath := flags.String("w", "", "path/to/whitelist.yml")
	blacklistPath := flags.String("b", "", "path/to/blacklist.yml")
	outputPath := flags.String("o", "", "path/to/output. default stdout")
	monitoringSpan := flags.Int("i", 60, "Monitoring interval (second)")
	verbose := flags.Bool("v", false, "show Verbose")

	if !*verbose {
		log.SetOutput(ioutil.Discard)
	} else {
		log.SetOutput(os.Stderr)
	}

	if err := flags.Parse(args); err != nil {
		fmt.Println(err)
		return 1
	}

	if *whitelistPath == "" || *blacklistPath == "" {
		flags.Usage()
		return 1
	}
	monitor, err := psmonitor.NewMonitor(*whitelistPath, *blacklistPath, *outputPath, *monitoringSpan)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	if err := monitor.Monitor(); err != nil {
		fmt.Println(err)
		return 1
	}

	return 0
}

// Synopsis show usagelist without a subcommand
func (c *MonitorCommand) Synopsis() string {
	return "Monitor the process"
}

// Help show usage with a subcommand
func (c *MonitorCommand) Help() string {
	return "Usage: psChecker monitor -w path/to/whitelist.yml -b path/to/blacklist.yml [-i intervaltime of monitoring(second)] [-o outputPath] [-v]"
}

func main() {
	c := cli.NewCLI("psChecker", "0.0.1")

	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"show": func() (cli.Command, error) {
			return &ShowCommand{}, nil
		},
		"monitor": func() (cli.Command, error) {
			return &MonitorCommand{}, nil
		},
	}

	_, err := c.Run()
	if err != nil {
		fmt.Println(err)
	}
}
