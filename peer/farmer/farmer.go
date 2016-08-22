package farmer

import (
	"fmt"

	fm "github.com/hyperledger/fabric/farmer"
	"github.com/op/go-logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	farmerFuncName = "farmer"
)

var logger = logging.MustGetLogger("farmerCmd")

var (
	farmerUserID         string
	farmerSupervisorAddr string
)

// Cmd returns the cobra command for Node
func Cmd() *cobra.Command {
	flags := chaincodeCmd.PersistentFlags()

	// TODO: more falgs
	flags.StringVarP(&farmerUserID, "userID", "u", "", "Your Account ID")
	flags.StringVarP(&farmerSupervisorAddr, "supervisorAddr", "a", "", "Supervisor Address")

	nodeCmd := &cobra.Command{
		Use:   farmerFuncName,
		Short: fmt.Sprintf("%s specific commands.", farmerFuncName),
		Long:  fmt.Sprintf("%s specific commands.", farmerFuncName),
	}

	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "User Login.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ParseArgs()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return login(args)
		},
	}
	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "User Logout.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return logout(args)
		},
	}

	nodeCmd.AddCommand(loginCmd)
	nodeCmd.AddCommand(logoutCmd)

	return nodeCmd
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

func login(args []string) error {
	fm.Login()
	return nil
}

func logout(args []string) error {
	return nil
}
