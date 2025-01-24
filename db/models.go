package db

import "gorm.io/gorm"

type Peer struct {
  gorm.Model
  Name string
  PublicKey string `gorm:"unique"`
  PrivateKey string
  PresharedKey string
}
