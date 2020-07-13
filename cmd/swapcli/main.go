package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/transmutate-io/atomicswap/cmd/swapcli/cmds"
)

var (
	rootCmd = &cobra.Command{
		Use:   "swapcli",
		Short: "atomic swaps cli tool",
		Long:  "swapcli is a command line tool to perform atomic swaps",
	}
)

func init() {
	hd, err := homedir.Dir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't get homedir: %v\n", err)
		os.Exit(-1)
	}
	for _, i := range []*cobra.Command{
		cmds.ListCryptosCmd,
		cmds.AutoCompleteCmd,
		cmds.TradeCmd,
		cmds.ProposalCmd,
	} {
		rootCmd.AddCommand(i)
	}
	pf := rootCmd.PersistentFlags()
	pf.StringP("data", "D", filepath.Join(hd, ".swapcli"), "set datadir")
	pf.StringP("output", "o", "-", "set output")
	pf.StringP("input", "i", "-", "set input")
}

func main() {
	rootCmd.Execute()
}
