package valkyrie

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

const (
	RecommendedRoundsSHA1   = 131000
	RecommendedRoundsSHA256 = 29000
	RecommendedRoundsSHA512 = 25000
)

var b64 = base64.RawStdEncoding

// PassLibBase64Encode encodes using a variant of base64, like Passlib.
// Check https://pythonhosted.org/passlib/lib/passlib.utils.html#passlib.utils.ab64_encode
func PassLibBase64Encode(src []byte) (dst string) {
	dst = b64.EncodeToString(src)
	dst = strings.Replace(dst, "+", ".", -1)
	return
}

// PassLibBase64Decode decodes using a variant of base64, like Passlib.
// Check https://pythonhosted.org/passlib/lib/passlib.utils.html#passlib.utils.ab64_decode
func PassLibBase64Decode(src string) (dst []byte, err error) {
	src = strings.Replace(src, ".", "+", -1)
	dst, err = b64.DecodeString(src)
	return
}

// Base64Encode encodes using a Standard of base64.
// return string base64 encode
func Base64Encode(src []byte) (dst string) {
	return base64.StdEncoding.EncodeToString(src)
}

// Base64Decode decodes using a Standard of base64.
// return string base64 encode
func Base64Decode(src string) (dst []byte, err error) {
	return base64.StdEncoding.DecodeString(src)
}

func HashPassword(password, salt string) string {
	return fmt.Sprintf(
		"$pbkdf2-sha512$%d$%s$%v",
		RecommendedRoundsSHA512,
		PassLibBase64Encode([]byte(salt)),
		PassLibBase64Encode(
			pbkdf2.Key(
				[]byte(password),
				[]byte(salt),
				RecommendedRoundsSHA512,
				sha512.Size, sha512.New,
			),
		),
	)
}
