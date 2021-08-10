# imago
Image saving & comparing tool for go based on webp.

## Functions
### func Str2bytes(s string) []byte
Fast convert
### func Bytes2str(b []byte) string
Fast convert
### func GetDHashStr(img image.Image) (string, error)
Get image dhash encoded by [go-base16384](https://github.com/fumiama/go-base16384)
### func HammDistance(img1 string, img2 string) (int, error)
Get hamming distance between two dhash strings
### func Scanimgs(imgdir string) error
Scan all images like 编码后哈希.webp
### func Pick(exclude []string) string
Pick a random image
### func Saveimgbytes(b []byte, imgdir string, uid string, force bool) string
### func Saveimg(r io.Reader, imgdir string, uid string) string
Save image into imgdir with name like 编码后哈希.webp
### func Addimage(name string)
manually add an image name into map