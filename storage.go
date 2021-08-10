// Package imago Image saving & comparing tool for go based on webp.
package imago

import (
	"bytes"
	"image"
	"io"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/kolesa-team/go-webp/decoder"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	log "github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

var (
	images = make(map[string][]string)
	mutex  sync.Mutex
)

func init() {
	log.SetFormatter(&easy.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		LogFormat:       "[imago][%time%][%lvl%]: %msg% \n",
	})
	log.SetLevel(log.DebugLevel)
}

// Setloglevel
func Setloglevel(level log.Level) {
	log.SetLevel(level)
}

// Imgexsits Return whether the name is in map
func Imgexsits(name string) bool {
	index := name[:3]
	tail := name[3:]
	tails, ok := images[index]
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

// Addimage manually add an image name into map
func Addimage(name string) {
	index := name[:3]
	tail := name[3:]
	mutex.Lock()
	defer mutex.Unlock()
	if images[index] == nil {
		images[index] = make([]string, 0)
		log.Debugln("[addimage] create index", index, ".")
	}
	images[index] = append(images[index], tail)
	log.Debugln("[addimage] index", index, "append file", tail, ".")
	images["sum"] = append(images["sum"], name)
}

// Saveimgbytes Save image into imgdir with name like 编码后哈希.webp Return value: status, dhash
func Saveimgbytes(b []byte, imgdir string, uid string, force bool, samediff int) (string, string) {
	r := bytes.NewReader(b)
	img, _, err := image.Decode(r)
	iswebp := false
	if err != nil {
		r.Seek(0, io.SeekStart)
		img, err = webp.Decode(r, &decoder.Options{})
		if err == nil {
			iswebp = true
		} else {
			log.Errorf("[saveimg] decode image error: %v\n", err)
			return "\"stat\": \"notanimg\"", ""
		}
	}
	dh, err := GetDHashStr(img)
	if err != nil {
		log.Errorf("[saveimg] get dhash error: %v\n", err)
		return "\"stat\": \"dherr\"", ""
	}
	if force {
		if Imgexsits(dh) {
			log.Debugf("[saveimg] force find similar image %s.\n", dh)
			return "\"stat\":\"exist\", \"img\": \"" + url.QueryEscape(dh) + "\"", dh
		}
	} else {
		for _, name := range images["sum"] {
			diff, err := HammDistance(dh, name)
			if err == nil && diff <= samediff { // 认为是一张图片
				log.Debugf("[saveimg] old %s.\n", name)
				return "\"stat\":\"exist\", \"img\": \"" + url.QueryEscape(name) + "\"", name
			}
		}
	}
	f, err := os.Create(imgdir + dh + ".webp")
	if err != nil {
		log.Errorf("[saveimg] create webp file error: %v\n", err)
		return "\"stat\": \"ioerr\"", ""
	}
	defer f.Close()
	if !iswebp {
		options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, 75)
		if err != nil || webp.Encode(f, img, options) != nil {
			log.Errorf("[saveimg] encode webp error: %v\n", err)
			return "\"stat\": \"encerr\"", ""
		}
	} else {
		r.Seek(0, io.SeekStart)
		c, err := io.Copy(f, r)
		if err != nil {
			log.Errorf("[saveimg] copy file error: %v\n", err)
			return "\"stat\": \"ioerr\"", ""
		}
		log.Debugf("[saveimg] save %d bytes.\n", c)
	}
	log.Debugf("[saveimg] new %s.\n", dh)
	return "\"stat\":\"success\", \"img\": \"" + url.QueryEscape(dh) + "\"", dh
}

// Saveimg Save image into imgdir with name like 编码后哈希.webp Return value: status, dhash
func Saveimg(r io.Reader, imgdir string, uid string, samediff int) (string, string) {
	imgbuff := make([]byte, 1024*1024) // 1m
	r.Read(imgbuff)
	return Saveimgbytes(imgbuff, imgdir, uid, false, samediff)
}

// Scanimgs Scan all images like 编码后哈希.webp
func Scanimgs(imgdir string) error {
	entry, err := os.ReadDir(imgdir)
	if err != nil {
		return err
	}
	for _, i := range entry {
		if !i.IsDir() {
			name := i.Name()
			if strings.HasSuffix(name, ".webp") {
				name = name[:len(name)-5]
				if len([]rune(name)) == 5 {
					Addimage(name)
				}
			}
		}
	}
	return nil
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
func Pick(exclude []string) string {
	sum := images["sum"]
	le := len(exclude)
	ls := len(sum)
	if le >= ls {
		return ""
	} else if le == 0 {
		return sum[rand.Intn(len(sum))]
	} else if ls/le > 10 {
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
