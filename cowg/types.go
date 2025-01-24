package cowg

import (
	"github.com/utkarsh867/cowg/db"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Cowg struct {
  Db *db.DB
  WgClient *wgctrl.Client
  WgDevice *wgtypes.Device
  Config string
}
