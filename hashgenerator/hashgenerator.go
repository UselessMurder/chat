package hashgenerator

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"time"
)

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func GenerateRandomString(s int) (string, error) {
	b, err := GenerateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}

func GenerateHash28(Value, solt string) (string, error) {

	var bytes28 [28]byte

	str, err := GenerateRandomString(16)

	if err != nil {
		return string(bytes28[:]), err
	}

	bytes28 = sha512.Sum512_224([]byte(time.Now().String() + str + Value + solt))

	return string(bytes28[:]), nil
}

func GenerateHash64(Value, solt string) (string, error) {

	var bytes64 [64]byte

	str, err := GenerateRandomString(16)

	if err != nil {
		return string(bytes64[:]), err
	}

	bytes64 = sha512.Sum512([]byte(time.Now().String() + str + Value + solt))

	return string(bytes64[:]), nil
}

func GetHashSum28(Value, solt string) (string, error) {

	bytes28 := sha512.Sum512_224([]byte(Value + solt))

	return string(bytes28[:]), nil
}

func GetHashSum64(Value, solt string) (string, error) {

	bytes64 := sha512.Sum512([]byte(Value + solt))

	return string(bytes64[:]), nil
}
