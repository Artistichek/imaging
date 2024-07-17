package base64

import "encoding/base64"

// Decode img in base64 format to byte array
func Decode(s string) []byte {
	buffer, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return buffer
}
