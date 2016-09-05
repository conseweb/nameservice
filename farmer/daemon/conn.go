package daemon

import (
	pb "github.com/conseweb/common/protos"
	"google.golang.org/grpc"
)

func (d *Daemon) ConnSupervisor(addr string) error {
	if d.supervisorConn != nil {
		d.supervisorConn.Close()
	}

	if err := d.connectSupervisor(addr); err != nil {
		return err
	}

	d.svCli = pb.NewFarmerPublicClient(d.supervisorConn)

	return nil
}

func (d *Daemon) connectSupervisor(addr string) error {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(defaultTimeout),
		grpc.WithBlock(),
	}

	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return err
	}

	d.Lock()
	defer d.Unlock()
	d.supervisorConn = conn
	return nil
}

func (d *Daemon) ConnIdprovider(addr string) error {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(defaultTimeout),
		grpc.WithBlock(),
	}

	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return err
	}

	d.Lock()
	d.idppCli = pb.NewIDPPClient(conn)
	d.Unlock()
	return nil
}

func (d *Daemon) GetIDPClient() (pb.IDPPClient, error) {
	if d.idppCli != nil {
		return d.idppCli, nil
	}

	err := d.ConnIdprovider(d.IDProviderAddr)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return d.idppCli, nil
}

func (d *Daemon) GetSVClient() (pb.FarmerPublicClient, error) {
	if d.svCli != nil {
		return d.svCli, nil
	}

	err := d.ConnSupervisor(d.SupervisorAddr)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return d.svCli, nil
}

func (d *Daemon) CloseConn() {
	if d.idproviderConn != nil {
		d.idproviderConn.Close()
	}
	if d.supervisorConn != nil {
		d.supervisorConn.Close()
	}
}
