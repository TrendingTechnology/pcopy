package main

import (
	"errors"
	"flag"
	"os"
	"pcopy"
	"regexp"
	"strings"
)

func execCopy(args []string) {
	config, fileId := parseClientArgs("copy", args)
	client := pcopy.NewClient(config)

	if err := client.Copy(os.Stdin, fileId); err != nil {
		fail(err)
	}
}

func execPaste(args []string)  {
	config, fileId := parseClientArgs("paste", args)
	client := pcopy.NewClient(config)

	if err := client.Paste(os.Stdout, fileId); err != nil {
		fail(err)
	}
}

func parseClientArgs(command string, args []string) (*pcopy.Config, string) {
	flags := flag.NewFlagSet(command, flag.ExitOnError)
	configFile := flags.String("config", "", "Alternate config file")
	serverAddr := flags.String("server", "", "Server address")
	if err := flags.Parse(args); err != nil {
		fail(err)
	}

	// Parse alias and file
	alias := "default"
	fileId := "default"
	if flags.NArg() > 0 {
		re := regexp.MustCompile(`^(?:([-_a-zA-Z0-9]+):)?([-_a-zA-Z0-9]*)$`)
		parts := re.FindStringSubmatch(flags.Arg(0))
		if len(parts) != 3 {
			fail(errors.New("invalid argument, must be in format [ALIAS:]FILEID"))
		}
		if parts[1] != "" {
			if *configFile != "" {
				fail(errors.New("invalid argument, -config cannot be set when alias is given"))
			}
			alias = parts[1]
		}
		if parts[2] != "" {
			fileId = parts[2]
		}
	}

	// Load config
	config, err := pcopy.LoadConfig(*configFile, alias)
	if err != nil {
		fail(err)
	}

	// Load defaults
	if config.CertFile == "" {
		certFile := strings.TrimSuffix(*configFile, ".conf") + ".crt"
		if _, err := os.Stat(certFile); err == nil {
			config.CertFile = certFile
		}
	}

	// Command line overrides
	if *serverAddr != "" {
		config.ServerAddr = *serverAddr
	}
	// FIXME add -key parsing

	// Validate
	if config.ServerAddr == "" {
		fail(errors.New("server address missing, specify -server flag or add 'ServerAddr' to config"))
	}
	if config.Key == nil {
		fail(errors.New("key missing, specify -key flag or add 'Key' to config"))
	}

	return config, fileId
}