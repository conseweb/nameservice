package api

import (
	"encoding/json"
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

// SetFileIndex /indexer/files/:device_id?clean=false clean old files in this deviceID
func SetFileIndex(ctx *RequestContext, orm *xorm.Engine, params martini.Params) {
	devID := params["device_id"]
	isClean, _ := strconv.ParseBool("clean")

	var files []*indexer.FileInfo
	err := json.NewDecoder(ctx.req.Body).Decode(&files)
	if err != nil {
		ctx.Error(400, err)
		return
	}

	if isClean {
		_, err := orm.Where("device_id = ?", devID).Delete(&indexer.FileInfo{})
		if err != nil {
			ctx.Error(500, err)
			return
		}
	}

	insrt := []interface{}{}
	for _, file := range files {
		file.DeviceID = devID
		insrt = append(insrt, file)
	}

	n, err := orm.Insert(insrt...)
	if err != nil {
		ctx.Error(500, err)
		return
	}

	ctx.Message(201, n)
}

func OnlineDevice(ctx *RequestContext, orm *xorm.Engine, params martini.Params) {
	devID := params["device_id"]

	var dev indexer.Device

	err := json.NewDecoder(ctx.req.Body).Decode(&dev)
	if err != nil {
		ctx.Error(400, "invalid address")
		return
	}

	_, err = orm.Where("device_id = ?", devID).Delete(&indexer.Device{})
	if err != nil {
		ctx.Error(500, err)
		return
	}

	dev.ID = devID

	_, err = orm.Insert(dev)
	if err != nil {
		ctx.Error(500, err)
		return
	}

	ctx.Message(201, "ok")
}

func OfflineDevice(ctx *RequestContext, orm *xorm.Engine, params martini.Params) {
	devID := params["device_id"]

	_, err := orm.Where("device_id = ?", devID).Delete(&indexer.Device{})
	if err != nil {
		ctx.Error(500, err)
		return
	}

	ctx.Message(200, "ok")
}
