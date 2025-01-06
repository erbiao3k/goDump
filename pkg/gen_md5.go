package pkg

import (
	"crypto/md5"
	"encoding/hex"
)

func GenMd5(str string) string {
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}
