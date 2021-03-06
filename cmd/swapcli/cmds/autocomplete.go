package cmds

import (
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/transmutate-io/atomicswap/internal/cmdutil"
	"github.com/transmutate-io/atomicswap/internal/flagutil"
	"github.com/transmutate-io/atomicswap/internal/flagutil/exitcodes"
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
		Short:   "fish autocomplete script",
		Aliases: []string{"f"},
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cmdAutoComplete(cmd, args, func(w io.Writer) error {
				return cmd.Root().GenFishCompletion(w, flagutil.MustBool(cmd.Flags(), "desc"))
			})
		},
	}
	fishCommand.Flags().BoolP("desc", "d", false, "include descriptions")
	cmdutil.AddCommands(AutoCompleteCmd, []*cobra.Command{
		{
			Use:     "auto",
			Short:   "autocomplete script (try to guess the shell)",
			Aliases: []string{"a"},
			Args:    cobra.NoArgs,
			Run: func(cmd *cobra.Command, args []string) {
				cmdAutoComplete(cmd, args, nil)
			},
		},
		fishCommand,
		{
			Use:     "bash",
			Short:   "bash autocomplete script",
			Aliases: []string{"b"},
			Args:    cobra.NoArgs,
			Run: func(cmd *cobra.Command, args []string) {
				cmdAutoComplete(cmd, args, cmd.Root().GenBashCompletion)
			},
		},
		{
			Use:     "zsh",
			Short:   "zsh autocomplete script",
			Aliases: []string{"z"},
			Args:    cobra.NoArgs,
			Run: func(cmd *cobra.Command, args []string) {
				cmdAutoComplete(cmd, args, cmd.Root().GenZshCompletion)
			},
		},
		{
			Use:     "powershell",
			Short:   "powershell autocomplete script",
			Aliases: []string{"p"},
			Args:    cobra.NoArgs,
			Run: func(cmd *cobra.Command, args []string) {
				cmdAutoComplete(cmd, args, cmd.Root().GenPowerShellCompletion)
			},
		},
	})
}

func cmdAutoComplete(cmd *cobra.Command, args []string, gen func(io.Writer) error) {
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
				cmdutil.ErrorExit(exitcodes.UnknownShell, "")
			}
			gen = cmd.Root().GenPowerShellCompletion
		}
	}
	out, closeOut := flagutil.MustOpenOutput(cmd.Flags())
	defer closeOut()
	if err := gen(out); err != nil {
		cmdutil.ErrorExit(exitcodes.UnknownShell, err)
	}
}
