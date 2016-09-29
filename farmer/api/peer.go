package api

import (
	"fmt"
	"net/http"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/hyperledger/fabric/core/peer"
	pb "github.com/hyperledger/fabric/protos"
	"golang.org/x/net/context"
)

func StartPeer(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
	if err := daemon.StartPeer(); err != nil {
		ctx.Error(400, err)
		return
	}
	ctx.rnd.JSON(200, map[string]string{"msg": "ok"})
}

func StopPeer(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
	if err := daemon.StopPeer(); err != nil {
		ctx.Error(400, err)
		return
	}
	ctx.rnd.JSON(200, map[string]string{"msg": "ok"})
}

func RestartPeer(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
	ctx.rnd.JSON(200, map[string]string{"msg": "ok"})
}

func GetPeerState(rw http.ResponseWriter, req *http.Request, ctx *RequestContext) {
	clientConn, err := peer.NewPeerClientConnection()
	if err != nil {
		log.Infof("Error trying to connect to local peer: %s", err)
		ctx.Error(500, fmt.Errorf("Error trying to connect to local peer: %s", err))
		return
	}

	serverClient := pb.NewAdminClient(clientConn)

	status, err := serverClient.GetStatus(context.Background(), &empty.Empty{})
	if err != nil {
		log.Infof("Error trying to get status from local peer: %s", err)
		ctx.Error(500, fmt.Errorf("Error trying to connect to local peer: %s", err))
		return
	}
	ctx.rnd.JSON(200, status)
}
