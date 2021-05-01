package valkyrie

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase64(t *testing.T) {
	var flagtests = []struct {
		title string
		word  string
	}{
		{"hash password", "sekretuwhw8w8ewyewueibxw74h747hwuwywe74wuey7273y23ebebyd6773yye3456"},
		{"ascii word", "Hello, 世界"},
	}
	for _, tt := range flagtests {
		tt := tt // pin it
		t.Run(tt.title, func(t *testing.T) {
			encode := Base64Encode([]byte(tt.word))
			// assert.NoError(t, err)
			decode, err := Base64Decode(encode)
			assert.NoError(t, err)
			assert.Equal(t, tt.word, string(decode))
		})
	}
}

func TestPassLibBase64(t *testing.T) {
	var flagtests = []struct {
		title string
		word  string
	}{
		{"hash password", "sekretuwhw8w8ewyewueibxw74h747hwuwywe74wuey7273y23ebebyd6773yye3456"},
		{"ascii word", "Hello, 世界"},
	}
	for _, tt := range flagtests {
		tt := tt // pin it
		t.Run(tt.title, func(t *testing.T) {
			encode := PassLibBase64Encode([]byte(tt.word))
			// assert.NoError(t, err)
			decode, err := PassLibBase64Decode(encode)
			assert.NoError(t, err)
			assert.Equal(t, tt.word, string(decode))
		})
	}
}

func TestHashPassword(t *testing.T) {
	var flagtests = []struct {
		title string
		hash  string
		pass  string
		salt  string
	}{
		{"Hash 1", "$pbkdf2-sha512$25000$c2VrcmV0$pXDTYx14sKdpiPP5kV8eqU74rNGucdLxohyYzjKK6Gl9jWkG97dgEtk.LE50IXpg8Cd1YH.I98EA1FRFIuG6mQ", "Pass1234", "sekret"},
		{"Hash 2", "$pbkdf2-sha512$25000$dGVya2Vkam9ldA$ZxJx.AQ7CVNvgDZAGBC4IZhlp7V.3Q4ljtsVlljWRephWZoJzmTweoPCoK6p13o9A9IdiRHfhWXCoDwkRrtwHA", "Pass1234", "terkedjoet"},
	}
	for _, tt := range flagtests {
		tt := tt // pin it
		t.Run(tt.title, func(t *testing.T) {
			hash := HashPassword(tt.pass, tt.salt)

			assert.Equal(t, hash, tt.hash)
		})
	}
}
