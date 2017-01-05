package api

import (
	"fmt"
	"io"
	"strings"

	"github.com/go-martini/martini"
	"github.com/hyperledger/fabric/storage"
	"golang.org/x/net/context"
)

// GetFile GET /fs/cat/**
func GetFile(ctx *RequestContext, params martini.Params, fs storage.StorageDriver) {
	f, err := fs.Reader(context.TODO(), getFilePath(params))
	if err != nil {
		ctx.Error(400, err)
		return
	}

	defer f.Close()
	io.Copy(ctx.res, f)
}

// GetFileList GET /fs/ls/**
func GetFileList(ctx *RequestContext, params martini.Params, fs storage.StorageDriver) {
	fis, err := fs.List(context.TODO(), getFilePath(params))
	if err != nil {
		ctx.Error(500, err)
		return
	}

	ctx.rnd.JSON(200, fis)
}

// UploadFile PUT /fs/new/**
func UploadFile(ctx *RequestContext, params martini.Params, fs storage.StorageDriver) {
	mf, _, err := ctx.req.FormFile("file")
	if err != nil {
		ctx.Error(400, err)
		return
	}
	defer mf.Close()

	fw, err := fs.Writer(context.TODO(), getFilePath(params), false)
	if err != nil {
		ctx.Error(400, err)
		return
	}
	defer fw.Close()

	_, err = io.Copy(fw, mf)
	if err != nil {
		ctx.Error(500, err)
		return
	}

	ctx.Message(201, "ok")
}

// NewDir POST /fs/mkdir/**
func NewDir(ctx *RequestContext, params martini.Params, fs storage.StorageDriver) {
	err := fs.Mkdir(context.TODO(), getFilePath(params))
	if err != nil {
		ctx.Error(400, err)
		return
	}
	ctx.Message(200, getFilePath(params))
}

// RenameFile PATCH /fs/rename/**
func RenameFile(ctx *RequestContext, params martini.Params, fs storage.StorageDriver) {
	oldPath := getFilePath(params)
	newPath := ctx.params["newpath"]
	if newPath == "" {
		ctx.Error(400, fmt.Errorf("required newpath"))
		return
	}

	err := fs.Move(context.TODO(), oldPath, newPath)
	if err != nil {
		ctx.Error(400, err)
		return
	}

	ctx.Message(200, newPath)
}

// RemoveFile DELETE /fs/rm/**
func RemoveFile(ctx *RequestContext, params martini.Params, fs storage.StorageDriver) {
	err := fs.Delete(context.TODO(), getFilePath(params))
	if err != nil {
		ctx.Error(400, err)
		return
	}

	ctx.Message(200, "ok")
}

func paramsToSlice(params martini.Params) []string {
	ret := []string{}
	for i := 1; ; i++ {
		if val, ok := params[fmt.Sprintf("_%v", i)]; ok {
			ret = append(ret, val)
		} else {
			break
		}
	}
	return ret
}

func getFilePath(params martini.Params) string {
	sli := paramsToSlice(params)
	if len(sli) == 0 {
		return "/"
	}

	return strings.Join(sli, "/")
}
