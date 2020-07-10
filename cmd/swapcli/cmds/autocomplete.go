package cmds

import (
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	AutoCompleteCmd = &cobra.Command{
		Use:     "autocomplete",
		Short:   "generate shell autocomple",
		Aliases: []string{"auto", "a"},
	}
)

func init() {
	fishCommand := &cobra.Command{
		Use:     "fish",
		Short:   "generate fish autocomplete script",
		Aliases: []string{"f"},
		Run: func(cmd *cobra.Command, args []string) {
			cmdAutoComplete(cmd, args, func(w io.Writer) error {
				return cmd.Root().GenFishCompletion(w, flagBool(cmd.Flags(), "desc"))
			})
		},
	}
	fishCommand.Flags().BoolP("desc", "d", false, "include descriptions")
	for _, i := range []*cobra.Command{
		{
			Use:     "auto",
			Short:   "generate autocomplete script (try to guess the shell)",
			Aliases: []string{"a"},
			Run: func(cmd *cobra.Command, args []string) {
				cmdAutoComplete(cmd, args, nil)
			},
		},
		fishCommand,
		{
			Use:     "bash",
			Short:   "generate bash autocomplete script",
			Aliases: []string{"b"},
			Run: func(cmd *cobra.Command, args []string) {
				cmdAutoComplete(cmd, args, cmd.Root().GenBashCompletion)
			},
		},
		{
			Use:     "zsh",
			Short:   "generate zsh autocomplete script",
			Aliases: []string{"z"},
			Run: func(cmd *cobra.Command, args []string) {
				cmdAutoComplete(cmd, args, cmd.Root().GenZshCompletion)
			},
		},
		{
			Use:     "powershell",
			Short:   "generate powershell autocomplete script",
			Aliases: []string{"p"},
			Run: func(cmd *cobra.Command, args []string) {
				cmdAutoComplete(cmd, args, cmd.Root().GenPowerShellCompletion)
			},
		},
	} {
		AutoCompleteCmd.AddCommand(i)
	}
}

func cmdAutoComplete(cmd *cobra.Command, args []string, gen func(io.Writer) error) {
	out, closeOut := openOutput(cmd)
	defer closeOut()
	if gen == nil {
		// try to guess shell
		switch sh := filepath.Base(os.Getenv("SHELL")); sh {
		case "bash":
			gen = cmd.Root().GenBashCompletion
		case "fish":
			gen = func(w io.Writer) error { return cmd.Root().GenFishCompletion(w, true) }
		case "zsh":
			gen = cmd.Root().GenZshCompletion
		default:
			if os.Getenv("ComSpec") == "" {
				errorExit(-4, "can't identify shell")
			}
			gen = cmd.Root().GenPowerShellCompletion
		}
	}
	if err := gen(out); err != nil {
		errorExit(ECUnknownShell, "can't generate completion file: %#v\n", err)
	}
}
