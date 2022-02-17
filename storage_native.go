package imago

import (
	"bytes"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/kolesa-team/go-webp/decoder"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	log "github.com/sirupsen/logrus"
)

type NativeStorage struct {
	storage
}

func NewNativeStorage() (ns *NativeStorage) {
	ns = &NativeStorage{}
	ns.images = make(map[string][]string)
	return
}

// GetImgBytes Get image from apiurl with name like 编码后哈希.webp Return value: imgbytes, error
func (*NativeStorage) GetImgBytes(imgdir, name string) ([]byte, error) {
	return os.ReadFile(imgdir + "/" + name)
}

// SaveImgBytes Save image into imgdir with name like 编码后哈希.webp Return value: status, dhash
func (ns *NativeStorage) SaveImgBytes(b []byte, imgdir string, force bool, samediff int) (string, string) {
	r := bytes.NewReader(b)
	img, _, err := image.Decode(r)
	iswebp := false
	if err != nil {
		r.Seek(0, io.SeekStart)
		img, err = webp.Decode(r, &decoder.Options{})
		if err == nil {
			iswebp = true
		} else {
			log.Errorf("[saveimg] decode image error: %v", err)
			return "\"stat\": \"notanimg\"", ""
		}
	}
	dh, err := GetDHashStr(img)
	if err != nil {
		log.Errorf("[saveimg] get dhash error: %v", err)
		return "\"stat\": \"dherr\"", ""
	}
	if force {
		if ns.IsImgExsits(dh) {
			log.Debugf("[saveimg] force find similar image %s.", dh)
			return "\"stat\":\"exist\", \"img\": \"" + url.QueryEscape(dh) + "\"", dh
		}
	} else {
		ns.mutex.RLock()
		s := ns.images["sum"]
		ns.mutex.RUnlock()
		for _, name := range s {
			diff, err := HammDistance(dh, name)
			if err == nil && diff <= samediff { // 认为是一张图片
				log.Debugf("[saveimg] old %s.", name)
				return "\"stat\":\"exist\", \"img\": \"" + url.QueryEscape(name) + "\"", name
			}
		}
	}
	f, err := os.Create(imgdir + "/" + dh + ".webp")
	if err != nil {
		log.Errorf("[saveimg] create webp file error: %v", err)
		return "\"stat\": \"ioerr\"", ""
	}
	defer f.Close()
	if !iswebp {
		options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, 75)
		if err != nil || webp.Encode(f, img, options) != nil {
			log.Errorf("[saveimg] encode webp error: %v", err)
			return "\"stat\": \"encerr\"", ""
		}
	} else {
		r.Seek(0, io.SeekStart)
		c, err := io.Copy(f, r)
		if err != nil {
			log.Errorf("[saveimg] copy file error: %v", err)
			return "\"stat\": \"ioerr\"", ""
		}
		log.Debugf("[saveimg] save %d bytes.", c)
	}
	log.Debugf("[saveimg] new %s.", dh)
	ns.AddImage(dh)
	return "\"stat\":\"success\", \"img\": \"" + url.QueryEscape(dh) + "\"", dh
}

// SaveImg Save image into imgdir with name like 编码后哈希.webp Return value: status, dhash
func (n *NativeStorage) SaveImg(r io.Reader, imgdir string, samediff int) (string, string) {
	imgbuff, err := io.ReadAll(r)
	if err != nil {
		return "\"stat\": \"ioerr\"", ""
	}
	return n.SaveImgBytes(imgbuff, imgdir, false, samediff)
}

// ScanImgs Scan all images like 编码后哈希.webp
func (ns *NativeStorage) ScanImgs(imgdir string) error {
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
					ns.AddImage(name)
				}
			}
		}
	}
	return nil
}

// SaveConf Save config file into storage
func (ns *NativeStorage) SaveConf(data []byte) error {
	return os.WriteFile("conf.pb", data, 0644)
}

// SaveConf Save config file into storage
func (ns *NativeStorage) GetConf() ([]byte, error) {
	return os.ReadFile("conf.pb")
}
