package ipfs

import (
	"github.com/op/go-logging"
)

var (
	log = logging.MustGetLogger("filesystem")
)

type Driver struct {
	Name string
}
