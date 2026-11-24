package texttools

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

// MD5Mode chooses one of the four hex forms people ask for.
type MD5Mode int

const (
	// MD5Hex32Lower is the full digest in lowercase.
	MD5Hex32Lower MD5Mode = iota
	// MD5Hex32Upper is the full digest in uppercase.
	MD5Hex32Upper
	// MD5Hex16Lower is characters 9 to 24 of the full digest, lowercase.
	MD5Hex16Lower
	// MD5Hex16Upper is characters 9 to 24 of the full digest, uppercase.
	MD5Hex16Upper
)

// MD5 returns the MD5 digest of s in the requested form. The 16-character
// forms are a slice of the 32-character digest, not another hash.
func MD5(s string, mode MD5Mode) string {
	sum := md5.Sum([]byte(s))
	h := hex.EncodeToString(sum[:])
	if mode == MD5Hex16Lower || mode == MD5Hex16Upper {
		h = h[8:24]
	}
	if mode == MD5Hex32Upper || mode == MD5Hex16Upper {
		h = strings.ToUpper(h)
	}
	return h
}
