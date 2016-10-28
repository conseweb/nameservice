package daemon

import (
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	pb "github.com/conseweb/common/protos"
	"github.com/hyperledger/fabric/farmer/account"
	"github.com/hyperledger/fabric/peer/node"
	_ "github.com/mattn/go-sqlite3"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

const (
	DefaultListenAddr = ":9375"

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
	RESTURL        string
	RootFS         string

	Account *account.Account

	pid    int
	exitCh chan error
	sync.Mutex

	supervisorConn *grpc.ClientConn
	idproviderConn *grpc.ClientConn
	svCli          pb.FarmerPublicClient
	idppCli        pb.IDPPClient

	// used save account info.
	localDB *sql.DB
}

func NewDaemon() *Daemon {
	d := &Daemon{
		SupervisorAddr: DefaultSupervisorAddr,
		IDProviderAddr: DefaultIDProviderAddr,
		ListenAddr:     DefaultListenAddr,
		RESTURL:        viper.GetString("rest.address"),
		RootFS:         viper.GetString("peer.fileSystemPath"),
		pid:            os.Getpid(),
		exitCh:         make(chan error),
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

	a, err := account.LoadFromFile()
	if err != nil {
		d.GetLogger().Errorf("load account failed", err)
	}
	if a != nil {
		d.Account = a
	}

	return d
}

//
type dbHandler interface {
	InitDB(db *sql.DB) error
}

func (d *Daemon) Init() error {
	if err := d.writePid(); err != nil {
		return err
	}

	a, err := account.LoadFromFile()
	if err == nil && a != nil {
		d.Account = a
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

func pidFilePath() string {
	return filepath.Join(viper.GetString("peer.fileSystemPath"), "farmer.pid")
}

func (d *Daemon) writePid() error {
	f, err := os.OpenFile(pidFilePath(), os.O_CREATE|os.O_EXCL|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("Write Pid File failed, error: %s", err.Error())
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "%d", os.Getpid())
	if err != nil {
		return err
	}

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

	os.RemoveAll(pidFilePath())
	d.CloseConn()

	time.Sleep(1 * time.Second)
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
		d.Account = a
	}
}

func (d *Daemon) GetRESTAddr() string {
	if d.RESTURL == "" {
		d.GetLogger().Warningf("rest_url is null")
		return ""
	}

	hostport := strings.Split(d.RESTURL, ":")
	if len(hostport) != 2 {
		d.GetLogger().Errorf("invalid rest addr %+v", hostport)
		return ""
	}

	host, port := hostport[0], hostport[1]
	if host == "" || host == "0.0.0.0" {
		host = "localhost"
	}

	return fmt.Sprintf("%s:%s", host, port)
}

func (d *Daemon) StartPeer() error {
	d.GetLogger().Debugf("try to start node")
	return node.Start()
}

func (d *Daemon) StopPeer() error {
	return node.Stop()
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
