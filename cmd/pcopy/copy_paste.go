package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"heckel.io/pcopy"
	"regexp"
	"syscall"
)

func execCopy(cmd string, args []string) {
	config, file := parseClientArgs(cmd, args)
	client, err := pcopy.NewClient(config)
	if err != nil {
		fail(err)
	}

	if err := client.Copy(os.Stdin, file); err != nil {
		fail(err)
	}
}

func execPaste(cmd string, args []string)  {
	config, fileId := parseClientArgs(cmd, args)
	client, err := pcopy.NewClient(config)
	if err != nil {
		fail(err)
	}

	if err := client.Paste(os.Stdout, fileId); err != nil {
		fail(err)
	}
}

func parseClientArgs(command string, args []string) (*pcopy.Config, string) {
	flags := flag.NewFlagSet(command, flag.ExitOnError)
	flags.Usage = func() { showCopyPasteUsage(flags) }
	configFileOverride := flags.String("config", "", "Alternate config file (default is based on clipboard name)")
	certFile := flags.String("cert", "", "Certificate file to use for cert pinning")
	serverAddr := flags.String("server", "", "Server address")
	if err := flags.Parse(args); err != nil {
		fail(err)
	}

	// Parse clipboard and file
	clipboard, file := parseClipboardAndFile(flags, *configFileOverride)

	// Load config
	configFile, config, err := pcopy.LoadConfig(*configFileOverride, clipboard)
	if err != nil {
		fail(err)
	}

	// Load defaults
	if config.CertFile == "" {
		config.CertFile = pcopy.DefaultCertFile(configFile, true)
	}

	// Command line overrides
	if *serverAddr != "" {
		config.ServerAddr = pcopy.ExpandServerAddr(*serverAddr)
	}
	if *certFile != "" {
		config.CertFile = *certFile
	}
	if os.Getenv("PCOPY_KEY") != "" {
		config.Key, err = pcopy.DecodeKey(os.Getenv("PCOPY_KEY"))
		if err != nil {
			fail(err)
		}
	}

	return config, file
}

func parseClipboardAndFile(flags *flag.FlagSet, configFileOverride string) (string, string) {
	clipboard := pcopy.DefaultClipboard
	file := pcopy.DefaultFile
	if flags.NArg() > 0 {
		re := regexp.MustCompile(`^(?:([-_a-zA-Z0-9]+):)?([-_a-zA-Z0-9]*)$`)
		parts := re.FindStringSubmatch(flags.Arg(0))
		if len(parts) != 3 {
			fail(errors.New("invalid argument, must be in format [CLIPBOARD:]FILE"))
		}
		if parts[1] != "" {
			if configFileOverride != "" {
				fail(errors.New("invalid argument, -config cannot be set when clipboard is given"))
			}
			clipboard = parts[1]
		}
		if parts[2] != "" {
			file = parts[2]
		}
	}
	return clipboard, file
}

func showCopyPasteUsage(flags *flag.FlagSet) {
	if flags.Name() == "pcopy copy" || flags.Name() == "pcp" {
		showCopyUsage(flags)
	} else {
		showPasteUsage(flags)
	}
}

func showCopyUsage(flags *flag.FlagSet) {
	fmt.Printf("Usage: %s [OPTIONS..] [[CLIPBOARD:]FILE]\n", flags.Name())
	fmt.Println()
	fmt.Println("Description:")
	fmt.Println("  Read from STDIN and copy to remote clipboard. FILE is the remote file name, and")
	fmt.Println("  CLIPBOARD is the name of the clipboard (both default to 'default').")
	fmt.Println()
	fmt.Println("  The command will load a the clipboard config from ~/.config/pcopy/$CLIPBOARD.conf or")
	fmt.Println("  /etc/pcopy/$CLIPBOARD.conf. Config options can be overridden using the command line options.")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Printf("  %s < foo.txt            # Copies foo.txt to the default clipboard\n", flags.Name())
	fmt.Printf("  %s bar < bar.txt        # Copies bar.txt to the default clipboard as 'bar'\n", flags.Name())
	fmt.Printf("  echo hi | %s work:      # Copies 'hi' to the 'work' clipboard\n", flags.Name())
	fmt.Printf("  echo ho | %s work:bla   # Copies 'ho' to the 'work' clipboard as 'bla'\n", flags.Name())
	fmt.Println()
	fmt.Println("Options:")
	flags.PrintDefaults()
	fmt.Println()
	fmt.Println("To override or specify the remote server key, you may pass the PCOPY_KEY variable.")
	syscall.Exit(1)
}

func showPasteUsage(flags *flag.FlagSet) {
	fmt.Printf("Usage: %s [OPTIONS..] [[CLIPBOARD:]FILE]\n", flags.Name())
	fmt.Println()
	fmt.Println("Description:")
	fmt.Println("  Write remote clipboard contents to STDOUT. FILE is the remote file name, and CLIPBOARD is")
	fmt.Println("  the name of the clipboard (both default to 'default').")
	fmt.Println()
	fmt.Println("  The command will load a the clipboard config from ~/.config/pcopy/$CLIPBOARD.conf or")
	fmt.Println("  /etc/pcopy/$CLIPBOARD.conf. Config options can be overridden using the command line options.")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Printf("  %s                   # Reads from the default clipboard and prints its contents\n", flags.Name())
	fmt.Printf("  %s bar > bar.txt     # Reads 'bar' from the default clipboard to file 'bar.txt'\n", flags.Name())
	fmt.Printf("  %s work:             # Reads from the 'work' clipboard and prints its contents\n", flags.Name())
	fmt.Printf("  %s work:ho > ho.txt  # Reads 'ho' from the 'work' clipboard to file 'ho.txt'\n", flags.Name())
	fmt.Println()
	fmt.Println("Options:")
	flags.PrintDefaults()
	fmt.Println()
	fmt.Println("To override or specify the remote server key, you may pass the PCOPY_KEY variable.")
	syscall.Exit(1)
}