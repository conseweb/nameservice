package api

import (
	"strconv"

	"github.com/go-martini/martini"
	"github.com/go-xorm/xorm"
	"github.com/hyperledger/fabric/farmer/indexer"
)

type FileWrapper struct {
	*indexer.FileInfo `json:",inline"`
	Addr              string `json:"address"`
}

// GET /indexer/address/:file_id
func GetFileAddr(ctx *RequestContext, orm *xorm.Engine, params martini.Params) {
	fileid, err := strconv.Atoi(params["file_id"])
	if err != nil {
		ctx.Error(400, "invalid params file_id.")
		return
	}

	files := make([]*indexer.FileInfo, 0)
	err = orm.Where("id = ?", fileid).Find(&files)
	if err != nil {
		ctx.Error(500, err)
		return
	}

	if len(files) == 0 {
		ctx.Error(404, "not found")
		return
	}

	file := files[0]
	devs := make([]*indexer.Device, 0)
	err = orm.Where("id = ?", file.DeviceID).Find(&devs)
	if err != nil {
		ctx.Error(500, err)
		return
	}

	if len(files) == 0 {
		ctx.Error(404, "not found running server.")
		return
	}

	ctx.rnd.JSON(200, FileWrapper{file, devs[0].Address})
}
