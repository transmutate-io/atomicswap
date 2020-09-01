package cmds

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/transmutate-io/atomicswap/internal/uiutil"
	"github.com/transmutate-io/atomicswap/trade"
)

func inputTradeName(cmd *cobra.Command, pr string, mustExist bool) (string, error) {
	return uiutil.InputSandboxedFilename(pr, tradesDir(cmd), mustExist)
}

func openTradeFromInput(cmd *cobra.Command, pr string) (string, trade.Trade, error) {
	tn, err := inputTradeName(cmd, pr, true)
	if err != nil {
		return "", nil, err
	}
	if tn == "" {
		return "", nil, nil
	}
	tr, err := openTradeFile(tn)
	if err != nil {
		fmt.Printf("can't open trade: %s\n", err)
		return "", nil, err
	}
	return tn, tr, nil
}
