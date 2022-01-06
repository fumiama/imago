package imago

import (
	"bytes"
	"image"
	"io"
	"net/url"
	"strings"

	"github.com/fumiama/simple-storage/client"
	"github.com/kolesa-team/go-webp/decoder"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	log "github.com/sirupsen/logrus"
)

type RemoteStorage struct {
	storage
	cli *client.Client
}

func NewRemoteStorage(apiurl string, key string) (r *RemoteStorage) {
	r = &RemoteStorage{cli: client.NewClient(apiurl, key)}
	r.images = make(map[string][]string)
	return
}

// GetImgBytes Get image from apiurl with name like 编码后哈希.webp Return value: imgbytes, error
func (remo *RemoteStorage) GetImgBytes(imgdir, name string) (b []byte, e error) {
	b, _, e = remo.cli.GetFile(imgdir, name)
	return
}

// SaveImgBytes Save image into apiurl with name like 编码后哈希.webp Return value: status, dhash
func (remo *RemoteStorage) SaveImgBytes(b []byte, imgdir string, force bool, samediff int) (string, string) {
	return remo.saveImg(bytes.NewReader(b), imgdir, force, samediff)
}

// SaveImgBytes Save image into apiurl with name like 编码后哈希.webp Return value: status, dhash
func (remo *RemoteStorage) saveImg(r io.ReadSeeker, imgdir string, force bool, samediff int) (string, string) {
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
		if remo.IsImgExsits(dh) {
			log.Debugf("[saveimg] force find similar image %s.", dh)
			return "\"stat\":\"exist\", \"img\": \"" + url.QueryEscape(dh) + "\"", dh
		}
	} else {
		remo.mutex.RLock()
		s := remo.images["sum"]
		remo.mutex.RUnlock()
		for _, name := range s {
			diff, err := HammDistance(dh, name)
			if err == nil && diff <= samediff { // 认为是一张图片
				log.Debugf("[saveimg] old %s.", name)
				return "\"stat\":\"exist\", \"img\": \"" + url.QueryEscape(name) + "\"", name
			}
		}
	}
	f := new(bytes.Buffer)
	if err != nil {
		log.Errorf("[saveimg] create webp file error: %v", err)
		return "\"stat\": \"ioerr\"", ""
	}
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
	remo.AddImage(dh)
	_ = remo.cli.SetFile(imgdir, dh+".webp", f.Bytes())
	return "\"stat\":\"success\", \"img\": \"" + url.QueryEscape(dh) + "\"", dh
}

// SaveImg Save image into apiurl with name like 编码后哈希.webp Return value: status, dhash
func (remo *RemoteStorage) SaveImg(r io.Reader, imgdir string, samediff int) (string, string) {
	imgbuff, err := io.ReadAll(r)
	if err != nil {
		return "\"stat\": \"ioerr\"", ""
	}
	return remo.saveImg(bytes.NewReader(imgbuff), imgdir, false, samediff)
}

// ScanImgs Scan all images like 编码后哈希.webp from apiurl
func (remo *RemoteStorage) ScanImgs(imgdir string) error {
	m, err := remo.cli.ListFiles(imgdir)
	if err != nil {
		return err
	}
	for name := range m {
		if strings.HasSuffix(name, ".webp") {
			name = name[:len(name)-5]
			if len([]rune(name)) == 5 {
				remo.AddImage(name)
			}
		}
	}
	return nil
}

// SaveConf Save config file into storage
func (remo *RemoteStorage) SaveConf(data []byte) error {
	return remo.cli.SetFile("cfg", "conf.pb", data)
}

// SaveConf Save config file into storage
func (remo *RemoteStorage) GetConf() (data []byte, err error) {
	data, _, err = remo.cli.GetFile("cfg", "conf.pb")
	return
}
