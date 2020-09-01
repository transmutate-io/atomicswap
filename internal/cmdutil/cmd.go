package cmdutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/transmutate-io/atomicswap/internal/flagutil/exitcodes"
)

func ErrorExit(code int, a ...interface{}) {
	f, ok := exitcodes.Messages[code]
	if !ok {
		t := make([]string, 0, len(a))
		for i := 0; i < len(a); i++ {
			t = append(t, "%#v")
		}
		f = "args: " + strings.Join(t, " ") + "\n"
	}
	fmt.Fprintf(os.Stderr, f, a...)
	os.Exit(code)
}

func AddCommands(cmd *cobra.Command, sub []*cobra.Command) {
	for _, i := range sub {
		cmd.AddCommand(i)
	}
}

var PathSep = string([]rune{filepath.Separator})

func TrimPath(s, p string) string {
	return strings.TrimPrefix(strings.TrimPrefix(s, p), PathSep)
}
