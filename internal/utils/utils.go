package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"math/rand"

	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/storage"
)

// encKey rand key
type encData struct {
	aesGCM cipher.AEAD
	nonce  []byte
}

// encInstance save encrypt data
var encInstance *encData

var errMathRand = errors.New("encode string error")

// Decode userId  from encrypted cookie
func Decode(shaUserID string, userID *string) error {
	if err := keyInit(); err != nil {
		return err
	}

	dst, err := hex.DecodeString(shaUserID)
	if err != nil {
		return err
	}

	src, err := encInstance.aesGCM.Open(nil, encInstance.nonce, dst, nil)
	if err != nil {
		return err
	}

	*userID = string(src)
	return nil
}

// Encode userId by GCM algorithm and get hex
func Encode(userID string) (string, error) {
	if err := keyInit(); err != nil {
		return "", err
	}
	src := []byte(userID)
	dst := encInstance.aesGCM.Seal(nil, encInstance.nonce, src, nil)
	sha := hex.EncodeToString(dst)
	return sha, nil
}

func keyInit() error {
	if encInstance == nil {
		key, err := generateRandomBytes(aes.BlockSize)
		if err != nil {
			return err
		}

		aesBlock, err := aes.NewCipher(key)
		if err != nil {
			return err
		}

		aesGCM, err := cipher.NewGCM(aesBlock)
		if err != nil {
			return err
		}

		nonce, err := generateRandomBytes(aesGCM.NonceSize())
		if err != nil {
			return err
		}

		encInstance = new(encData)
		encInstance.aesGCM = aesGCM
		encInstance.nonce = nonce
	}
	return nil
}

func generateRandomBytes(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func RandStringBytes() (storage.URLKey, error) {
	b, err := generateRandomBytes(config.LENHASH)
	if err != nil {
		return "", errMathRand
	}

	return storage.URLKey(hex.EncodeToString(b))[:config.LENHASH], nil
}
