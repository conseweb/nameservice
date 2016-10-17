package farmer

import (
	"github.com/hyperledger/fabric/farmer"
	"github.com/spf13/cobra"
)

func startCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "start farmer daemon.",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Debugf("start farmer daemon.")
			farmer.StartFarmer()
			return nil
		},
	}

	return cmd
}
