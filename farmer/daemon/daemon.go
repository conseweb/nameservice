package daemon

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	pb "github.com/conseweb/common/protos"
	"github.com/hyperledger/fabric/farmer/account"
	fpb "github.com/hyperledger/fabric/protos"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google/protobuf"
)

const (
	DefaultDaemonAddress = ":9375"
	DefaultFarmerPath    = "/var/run/farmer"
	// DefaultSocketFile    = DefaultFarmerPath + "/farmer.sock"
	DefaultPidFile = DefaultFarmerPath + "/farmer.pid"

	DefaultListenAddr     = ":9375"
	DefaultSupervisorAddr = ":9376"
	DefaultIDProviderAddr = ":9377"
)

var (
	std    *Daemon
	logger = logging.MustGetLogger("daemon")
)

type Daemon struct {
	SupervisorAddr string
	IDProviderAddr string
	ListenAddr     string

	farmerAccount *account.Account

	exitCh chan error
}

func NewDaemon() *Daemon {
	d := &Daemon{
		SupervisorAddr: DefaultSupervisorAddr,
		IDProviderAddr: DefaultIDProviderAddr,
		ListenAddr:     DefaultListenAddr,

		exitCh: make(chan int),
	}

	svAddr := viper.GetString("farmer.supervisorAddress")
	if svAddr == "" {
		logger.Warningf("not set %s, use default %s", "farmer.supervisorAddress", svAddr)
	} else {
		d.SupervisorAddr = svAddr
	}
	idpAddr := viper.GetString("farmer.idproviderAddress")
	if idpAddr == "" {
		logger.Warningf("not set %s, use default %s", "farmer.idproviderAddress", idpAddr)
	} else {
		d.IDProviderAddr = idpAddr
	}
	listenAddr := viper.GetString("daemon.address")
	if listenAddr != "" {
		d.ListenAddr = listenAddr
	}

	return d
}

func (d *Daemon) Init() error {
	if err := d.writePid(); err != nil {
		return err
	}
}

func (d *Daemon) waitExit() {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		logger.Info("exit by user.")
		d.Exit()
	case <-d.exitCh:
		return
	}
}

func (d *Daemon) writePid() error {
	f, err := os.OpenFile(DefaultPidFile, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("Write Pid File failed, error: %s", err.Error())
	}
	defer f.Close()

	fmt.Fprintf(f, "%d", os.Getpid())
	return nil
}

func (d *Daemon) Exit(err error) {
	select {
	case <-d.exitCh:
		logger.Debug("exited already.")
	default:
		close(d.exitCh)
	}
	time.Sleep(3 * time.Second)

	os.RemoveAll(DefaultPidFile)

	if err != nil {
		logger.Errorf("farmer daemon(PID: %d) exit by error: %s", d.Pid, err)
	} else {
		logger.Infof("farmer daemon(PID: %d) exit...", d.Pid)
	}
	os.Exit(0)
}

func (d *Daemon) GetLogger() *logging.Logger {
	if logger != nil {
		return logger
	}
	return logging.MustGetLogger("daemon")
}

// Return the serve status.
func (d *Daemon) GetFarmerStatus(ctx context.Context, in *google_protobuf.Empty, opts ...grpc.CallOption) (*fpb.FarmerStatus, error) {
	return nil, nil
}

func (d *Daemon) StartFarmer(ctx context.Context, in *fpb.FarmerUser, opts ...grpc.CallOption) (*fpb.Response, error) {
	return nil, nil
}

func (d *Daemon) StopFarmer(ctx context.Context, in *google_protobuf.Empty, opts ...grpc.CallOption) (*fpb.FarmerStatus, error) {
	return nil, nil
}

func (d *Daemon) proxyFarmerPublic() {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
	}

	conn, err := grpc.Dial(SupervisorAddr, opts...)
	if err != nil {
		logger.Error(err)
	}

	client := pb.NewFarmerPublicClient()
}
