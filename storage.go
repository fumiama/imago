// Package imago Image saving & comparing tool for go based on webp.
package imago

import (
	"io"
	"math/rand"
	"sync"

	log "github.com/sirupsen/logrus"
)

type StorageInstance interface {
	GetImgBytes(imgdir, name string) ([]byte, error)
	SaveImgBytes(b []byte, imgdir string, force bool, samediff int) (string, string)
	SaveImg(r io.Reader, imgdir string, samediff int) (string, string)
	ScanImgs(imgdir string)
}

type storage struct {
	images map[string][]string
	mutex  sync.RWMutex
	StorageInstance
}

// IsImgExsits Return whether the name is in map
func (sr *storage) IsImgExsits(name string) bool {
	index := name[:3]
	tail := name[3:]
	sr.mutex.RLock()
	tails, ok := sr.images[index]
	sr.mutex.RUnlock()
	if ok {
		found := false
		for _, t := range tails {
			if tail == t {
				found = true
				break
			}
		}
		return found
	}
	return false
}

// AddImage manually add an image name into map
func (sr *storage) AddImage(name string) {
	index := name[:3]
	tail := name[3:]
	sr.mutex.Lock()
	defer sr.mutex.Unlock()
	if sr.images[index] == nil {
		sr.images[index] = make([]string, 0)
		log.Debugf("[addimage] create index %v.", index)
	}
	sr.images[index] = append(sr.images[index], tail)
	log.Debugf("[addimage] index %v append file %v.", index, tail)
	sr.images["sum"] = append(sr.images["sum"], name)
}

func namein(name string, list []string) bool {
	in := false
	for _, item := range list {
		if name == item {
			in = true
			break
		}
	}
	return in
}

// Pick Pick a random image
func (sr *storage) Pick(exclude []string) string {
	sr.mutex.RLock()
	sum := sr.images["sum"]
	sr.mutex.RUnlock()
	le := len(exclude)
	ls := len(sum)
	if le >= ls {
		return ""
	} else if le == 0 {
		return sum[rand.Intn(len(sum))]
	} else if (ls >> 2) > le {
		name := sum[rand.Intn(len(sum))]
		for namein(name, exclude) {
			name = sum[rand.Intn(len(sum))]
		}
		return name
	} else {
		for _, n := range sum {
			if !namein(n, exclude) {
				return n
			}
		}
		return ""
	}
}
