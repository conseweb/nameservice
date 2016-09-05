package farmer

import (
	"fmt"

	// fm "github.com/hyperledger/fabric/farmer"
	"github.com/op/go-logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	farmerFuncName = "farmer"
)

var (
	logger = logging.MustGetLogger("farmerCmd")

	farmerUserID         string
	farmerSupervisorAddr string
	isDaemon             bool

	farmerCmd = &cobra.Command{
		Use:   farmerFuncName,
		Short: fmt.Sprintf("%s specific commands.", farmerFuncName),
		Long:  fmt.Sprintf("%s specific commands.", farmerFuncName),
	}
)

// Cmd returns the cobra command for Node
func Cmd() *cobra.Command {
	// flags := farmerCmd.PersistentFlags()

	// // TODO: more falgs
	// flags.StringVarP(&farmerUserID, "userID", "u", "", "Your Account ID")
	// flags.StringVarP(&farmerSupervisorAddr, "supervisorAddr", "a", "", "Supervisor Address")

	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "User Logout.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return logout(args)
		},
	}

	farmerCmd.AddCommand(startCmd())
	farmerCmd.AddCommand(logoutCmd)

	return farmerCmd
}

func ParseArgs() error {
	if farmerUserID == "" {
		farmerUserID = viper.GetString("farmer.id")
	}
	if farmerSupervisorAddr == "" {
		farmerSupervisorAddr = viper.GetString("farmer.supervisorAddress")
	}

	// check
	if farmerUserID == "" {
		return fmt.Errorf("account id is required")
	}
	if farmerSupervisorAddr == "" {
		return fmt.Errorf("supervisor address is required")
	}
	return nil
}

func start(args []string) error {
	return nil
}

func logout(args []string) error {
	return nil
}
