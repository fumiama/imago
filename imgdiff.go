package imago

import (
	"encoding/binary"
	"image"

	"github.com/corona10/goimagehash"
	base14 "github.com/fumiama/go-base16384"
)

var lastchar = "㴁"

func decodeDHash(imgname string) *goimagehash.ImageHash {
	b, err := base14.UTF82utf16be(StringToBytes(imgname + lastchar))
	if err == nil {
		dhb := base14.Decode(b)
		dh := binary.BigEndian.Uint64(dhb)
		return goimagehash.NewImageHash(dh, goimagehash.DHash)
	}
	return nil
}

// HammDistance Get hamming distance between two dhash strings
func HammDistance(img1 string, img2 string) (int, error) {
	b1 := decodeDHash(img1)
	b2 := decodeDHash(img2)
	return b1.Distance(b2)
}

// GetDHashStr Get image dhash encoded by go-base16384
func GetDHashStr(img image.Image) (string, error) {
	dh, err := goimagehash.DifferenceHash(img)
	if err == nil {
		var data [8]byte
		binary.BigEndian.PutUint64(data[:], dh.GetHash())
		e := base14.Encode(data[:])
		b, _ := base14.UTF16be2utf8(e)
		return BytesToString(b)[:15], nil
	}
	return "", err
}
