package daemon

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	pb "github.com/conseweb/common/protos"
	"github.com/hyperledger/fabric/farmer/account"
	// fpb "github.com/hyperledger/fabric/protos"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
	// "golang.org/x/net/context"
	"google.golang.org/grpc"
	// "google/protobuf"
)

const (
	DefaultListenAddr = ":9375"
	DefaultFarmerPath = "/var/run/farmer"
	// DefaultSocketFile    = DefaultFarmerPath + "/farmer.sock"
	DefaultPidFile = DefaultFarmerPath + "/farmer.pid"

	// for grpc
	DefaultSupervisorAddr = ":9376"
	DefaultIDProviderAddr = ":9377"
	defaultTimeout        = 3 * time.Second
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

	pid    int
	exitCh chan error
	sync.Mutex

	supervisorConn *grpc.ClientConn
	idproviderConn *grpc.ClientConn
	svCli          pb.FarmerPublicClient
	idppCli        pb.IDPPClient
}

func NewDaemon() *Daemon {
	d := &Daemon{
		SupervisorAddr: DefaultSupervisorAddr,
		IDProviderAddr: DefaultIDProviderAddr,
		ListenAddr:     DefaultListenAddr,

		pid:    os.Getpid(),
		exitCh: make(chan error),
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
	return nil
}

func (d *Daemon) WaitExit() {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		logger.Info("exit by user.")
		d.Exit(nil)
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
		return
	default:
		close(d.exitCh)
	}

	os.RemoveAll(DefaultPidFile)
	d.CloseConn()

	time.Sleep(3 * time.Second)
	if err != nil {
		logger.Errorf("farmer daemon(PID: %d) exit by error: %s", d.pid, err)
	} else {
		logger.Infof("farmer daemon(PID: %d) exit...", d.pid)
	}
	os.Exit(0)
}

func (d *Daemon) GetLogger() *logging.Logger {
	if logger != nil {
		return logger
	}
	return logging.MustGetLogger("daemon")
}

func (d *Daemon) ResetAccount(a *account.Account) {
	d.Lock()
	defer d.Unlock()
	if a != nil {
		d.farmerAccount = a
	}
}

// func (d *Daemon) proxyFarmerPublic() {
// 	opts := []grpc.DialOption{
// 		grpc.WithInsecure(),
// 		grpc.WithBlock(),
// 	}

// 	conn, err := grpc.Dial(d.SupervisorAddr, opts...)
// 	if err != nil {
// 		logger.Error(err)
// 	}

// 	client := pb.NewFarmerPublicClient()
// }
