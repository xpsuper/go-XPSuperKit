package XPSuperKit

import (
	"fmt"
	"strings"
	"strconv"
	"unicode"
	"unicode/utf8"
	"crypto/sha1"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/base64"
	"hash/crc32"
	"net/url"
)

type XPStringImpl struct {

}

func (s *XPStringImpl) SHA1(str string) string {
	hasher := sha1.New()
	hasher.Write([]byte(str))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (s *XPStringImpl) CRC32(str string) uint32 {
	return crc32.ChecksumIEEE([]byte(str))
}

func (s *XPStringImpl) MD5(str string) string {
	hasher := md5.New()
	hasher.Write([]byte(str))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (s *XPStringImpl) Base64Encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func (s *XPStringImpl) Base64Decode(str string) (result string, err error) {
	r, e := base64.StdEncoding.DecodeString(str)
	if e != nil {
		return "", e
	} else {
		return string(r), nil
	}
}

func (s *XPStringImpl) UrlEncode(str string) string {
	return url.QueryEscape(str)
}

func (s *XPStringImpl) UrlDecode(str string) (result string, err error) {
	r, e := url.QueryUnescape(str)
	if e != nil {
		return "", e
	} else {
		return string(r), nil
	}
}

func (s *XPStringImpl) Split(str, sep string) []string {
	return strings.Split(str, sep)
}

func (s *XPStringImpl) IndexOfSubString(str string, substr string) int {
	// 子串在字符串的字节位置
	result := strings.Index(str,substr)

	if result >= 0 {
		// 获得子串之前的字符串并转换成[]byte
		prefix := []byte(str)[0:result]
		// 将子串之前的字符串转换成[]rune
		rs := []rune(string(prefix))
		// 获得子串之前的字符串的长度，便是子串在字符串的字符位置
		result = len(rs)
	}

	return result
}

func (s *XPStringImpl) SubString(str string, start, length int) (substr string) {
	// 将字符串的转换成[]rune
	rs  := []rune(str)
	lth := len(rs)

	// 简单的越界判断
	if start < 0 {
		start = 0
	}

	if start >= lth {
		start = lth
	}

	end := start + length
	if end > lth {
		end = lth
	}

	// 返回子串
	return string(rs[start:end])
}

func (s *XPStringImpl) Format(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a ...)
}

func (s *XPStringImpl) Trim(str string) string {
	return strings.Trim(str, " ")
}

func (s *XPStringImpl) TrimLeft(str string) string {
	return strings.TrimLeft(str, " ")
}

func (s *XPStringImpl) TrimRight(str string) string {
	return strings.TrimRight(str, " ")
}

func (s *XPStringImpl) Equal(source, target string) bool {
	return strings.EqualFold(source, target)
}

func (s *XPStringImpl) StartWith(str, prefix string) bool {
	return strings.HasPrefix(str, prefix)
}

func (s *XPStringImpl) EndWith(str, suffix string) bool {
	return strings.HasSuffix(str, suffix)
}

func (s *XPStringImpl) IsEmpty(str string) bool {
	return s.Trim(str) == ""
}

func (s *XPStringImpl) Random(length int) string {
	if length == 0 {
		return ""
	}

	var seed = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")

	clen := len(seed)
	if clen < 2 || clen > 256 {
		panic("Wrong charset length for NewLenChars()")
	}

	maxrb := 255 - (256 % clen)
	b := make([]byte, length)
	r := make([]byte, length+(length/4))
	i := 0

	for {
		if _, e := rand.Read(r); e != nil {
			return ""
		}

		for _, rb := range r {
			c := int(rb)
			if c > maxrb {
				continue
			}
			b[i] = seed[c%clen]
			i++
			if i == length {
				return string(b)
			}
		}
	}
}

func (s *XPStringImpl) RandomNumber(length int) string {
	if length == 0 {
		return ""
	}

	var seed = []byte("0123456789")

	clen := len(seed)
	if clen < 2 || clen > 256 {
		panic("Wrong charset length for NewLenChars()")
	}

	maxrb := 255 - (256 % clen)
	b := make([]byte, length)
	r := make([]byte, length+(length/4))
	i := 0

	for {
		if _, e := rand.Read(r); e != nil {
			return ""
		}

		for _, rb := range r {
			c := int(rb)
			if c > maxrb {
				continue
			}
			b[i] = seed[c%clen]
			i++
			if i == length {
				return string(b)
			}
		}
	}
}

func (s *XPStringImpl) RandomWithSeed(length int, seed string) string {
	if length == 0 {
		return ""
	}

	var seedByte = []byte(seed)

	clen := len(seedByte)
	if clen < 2 || clen > 256 {
		panic("Wrong charset length for NewLenChars()")
	}

	maxrb := 255 - (256 % clen)
	b := make([]byte, length)
	r := make([]byte, length+(length/4))
	i := 0

	for {
		if _, e := rand.Read(r); e != nil {
			return ""
		}

		for _, rb := range r {
			c := int(rb)
			if c > maxrb {
				continue
			}
			b[i] = seedByte[c%clen]
			i++
			if i == length {
				return string(b)
			}
		}
	}
}

// 分词
// 将字符串分隔为一个单词列表，支持数字、驼峰风格等
func (s *XPStringImpl) Segment(str string) (entries []string) {
	if !utf8.ValidString(str) {
		return []string{str}
	}
	entries = []string{}
	var runes [][]rune
	lastClass := 0
	class := 0

	for _, r := range str {
		switch true {
		case unicode.IsLower(r):
			class = 1
		case unicode.IsUpper(r):
			class = 2
		case unicode.IsDigit(r):
			class = 3
		default:
			class = 4
		}
		if class == lastClass {
			runes[len(runes)-1] = append(runes[len(runes)-1], r)
		} else {
			runes = append(runes, []rune{r})
		}
		lastClass = class
	}

	for i := 0; i < len(runes)-1; i++ {
		if unicode.IsUpper(runes[i][0]) && unicode.IsLower(runes[i+1][0]) {
			runes[i+1] = append([]rune{runes[i][len(runes[i])-1]}, runes[i+1]...)
			runes[i] = runes[i][:len(runes[i])-1]
		}
	}

	for _, s := range runes {
		if len(s) > 0 {
			entries = append(entries, string(s))
		}
	}

	return
}


func (s *XPStringImpl) ToInt64(str string) int64 {
	out, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	}

	return out
}

func (s *XPStringImpl) ToInt(str string) int {
	out, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	}
	return int(out)
}

func (s *XPStringImpl) ToInt32(str string) int32 {
	out, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return 0
	}
	return int32(out)
}

