package account

import (
	"fmt"
	"time"

	pb "github.com/conseweb/common/protos"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	defaultTimeout = time.Second * 3
)

type Account struct {
	ID string

	logger         *logging.Logger
	supervisorAddr string
	supervisorConn *grpc.ClientConn
}

func NewAccount(id, svAddr string) *Account {
	return &Account{
		ID:             id,
		supervisorAddr: svAddr,
		logger:         logging.MustGetLogger("farmer"),
	}
}

func (a *Account) Login() error {
	if a.ID == "" {
		return fmt.Errorf("account id required")
	}

	client, err := a.GetSupervisorClient()
	if err != nil {
		return err
	}

	onlineReq := &pb.FarmerOnLineReq{FarmerID: a.ID}
	a.logger.Debugf("login with %+v", a)
	onlineRes, err := client.FarmerOnLine(context.Background(), onlineReq)
	if err != nil {
		return err
	}
	if onlineRes.Error != nil {
		a.logger.Errorf("login error: %#v", onlineRes.Error)
		return onlineRes.Error
	}

	return nil
}

func (a *Account) Logout() error {

	return nil
}

func (a *Account) GetStatus() error {
	return nil
}

func (a *Account) GetSupervisorClient() (*pb.FarmerPublicClient, error) {
	if a.supervisorConn == nil {
		if err := a.connectSupervisor(); err != nil {
			return nil, err
		}
	}
	return pb.NewFarmerPublicClient(a.supervisorConn), nil
}

func (a *Account) connectSupervisor() error {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(defaultTimeout),
		grpc.WithBlock(),
	}

	conn, err := grpc.Dial(SupervisorAddr, opts...)
	if err != nil {
		return err
	}

	a.supervisorConn = conn
	return nil
}

func (a *Account) ping() error {

}
