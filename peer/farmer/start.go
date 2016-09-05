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
			farmer.StartFarmer()
			return nil
		},
	}

	// flags := cmd.PersistentFlags()
	// flags.StringVarP(&farmerSupervisorAddr, "supervisorAddr", "a", "", "Supervisor Address")

	return cmd
}
