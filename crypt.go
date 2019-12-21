package jcrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

type Options struct {
	Entropy                 []byte
	GetEncryptionKeyHandler EncryptionKeySource
}

type EncryptionKeySource func() []byte

type cryptBlock struct {
	Mode       string `json:"mode"`
	DataBase64 string `json:"data"`
	Data       []byte `json:"-"`
}

func newCryptBlock(mode string, data []byte) cryptBlock {
	dataBase64 := base64.URLEncoding.EncodeToString(data)
	return cryptBlock{mode, dataBase64, data}
}

func parseCryptBlock(v interface{}) (cryptBlock, error) {
	block, ok := v.(map[string]interface{})
	if !ok {
		return cryptBlock{}, fmt.Errorf("unexpected content %T for encrypted value", v)
	}

	rawMode, ok := block["mode"]
	if !ok {
		return cryptBlock{}, fmt.Errorf("encrypted block without mode")
	}
	mode, ok := rawMode.(string)
	if !ok {
		return cryptBlock{}, fmt.Errorf("expected mode of encrypted block to be of type string")
	}

	rawData, ok := block["data"]
	if !ok {
		return cryptBlock{}, fmt.Errorf("encrypted block without data")
	}
	strData, ok := rawData.(string)
	if !ok {
		return cryptBlock{}, fmt.Errorf("expected data of encrypted block to be of type string")
	}
	data, err := base64.URLEncoding.DecodeString(strData)
	if err != nil {
		return cryptBlock{}, fmt.Errorf("expected data of encrypted block to be base64 encoded")
	}

	return cryptBlock{mode, strData, data}, nil
}

func Marshal(v interface{}, options *Options) ([]byte, error) {
	return jsonMarshal(v, marshalCryptHandler)
}

func marshalCryptHandler(src srcValue) (interface{}, bool, error) {
	if src.StructField != nil {
		jcryptTag := strings.Split(src.StructField.Tag.Get("jcrypt"), ",")
		if len(jcryptTag[0]) > 0 {
			if jcryptTag[0] == "aes" {
				val, err := marshalCryptAES(src)
				if err != nil {
					return nil, false, err
				}
				return val, true, nil
			}
			return nil, false, fmt.Errorf("unknown encryption mode %q", jcryptTag[0])
		}
	}

	return nil, false, nil
}

func marshalCryptAES(src srcValue) (interface{}, error) {
	//TODO encrypt json representation of arbitrary type

	str, ok := src.Interface().(string)
	if !ok {
		return nil, fmt.Errorf("encrypted values must be of type string")
	}

	encData, err := encryptAES([]byte(str), nil, nil)
	if err != nil {
		return nil, err
	}

	return newCryptBlock("aes", encData), nil
}

func Unmarshal(data []byte, v interface{}, options *Options) error {
	return jsonUnmarshal(data, v, unmarshalCryptHandler)
}

func unmarshalCryptHandler(src interface{}, dst dstValue) (bool, error) {
	if dst.StructField != nil {
		jcryptTag := strings.Split(dst.StructField.Tag.Get("jcrypt"), ",")
		if len(jcryptTag[0]) > 0 {
			if jcryptTag[0] == "aes" {
				return true, unmarshalCrypt(src, dst)
			}
			return false, fmt.Errorf("unknown encryption mode %q", jcryptTag[0])
		}
	}

	return false, nil
}

func unmarshalCrypt(src interface{}, dst dstValue) error {
	//TODO decrypt json representation of arbitrary type

	if dst.Kind() != reflect.String {
		return fmt.Errorf("encrypted values must be of type string")
	}
	val, ok := src.(string)
	if ok {
		// raw value present
		dst.Assign(val)
		return nil
	}

	block, err := parseCryptBlock(src)
	if err != nil {
		return err
	}

	switch block.Mode {
	case "none":
		dst.Assign(string(block.Data))
		return nil

	case "aes":
		raw, err := decryptAES(block.Data, nil, nil)
		if err != nil {
			return err
		}
		dst.Assign(string(raw))
		return nil

	default:
		return fmt.Errorf("unknown encryption mode %q", block.Mode)
	}
}

func encryptAES(data, key, entropy []byte) ([]byte, error) {
	c, err := aes.NewCipher(deriveKeyAES(key, entropy))
	if err != nil {
		return nil, err
	}

	safeData := packSafeData(data)
	encData := make([]byte, c.BlockSize()+len(safeData))
	if _, err := rand.Read(encData[:c.BlockSize()]); err != nil {
		return nil, fmt.Errorf("failed to read random initialization vector: %s", err.Error())
	}

	stream := cipher.NewCFBEncrypter(c, encData[:c.BlockSize()])
	stream.XORKeyStream(encData[c.BlockSize():], safeData)

	return encData, nil
}

func decryptAES(data, key, entropy []byte) ([]byte, error) {
	c, err := aes.NewCipher(deriveKeyAES(key, entropy))
	if err != nil {
		return nil, err
	}

	if len(data) < c.BlockSize() {
		return nil, fmt.Errorf("data block corrupt")
	}

	decData := make([]byte, len(data)-c.BlockSize())

	stream := cipher.NewCFBDecrypter(c, data[:c.BlockSize()])
	stream.XORKeyStream(decData, data[c.BlockSize():])

	return unpackSafeData(decData)
}

func deriveKeyAES(key, entropy []byte) []byte {
	if key == nil {
		key = []byte{}
	}
	if entropy == nil {
		entropy = []byte{}
	}

	return pbkdf2.Key(key, entropy, 4096, 32, sha256.New)
}

func packSafeData(data []byte) []byte {
	h := sha256.New()
	h.Write(data)
	hash := h.Sum(nil)
	safeData := make([]byte, 32+len(data))
	copy(safeData[:32], hash)
	copy(safeData[32:], data)
	return safeData
}

func unpackSafeData(safeData []byte) ([]byte, error) {
	if len(safeData) < 32 {
		return nil, fmt.Errorf("data block corrupt")
	}

	h := sha256.New()
	h.Write(safeData[32:])
	hash := h.Sum(nil)
	for i := 0; i < 32; i++ {
		if safeData[i] != hash[i] {
			return nil, fmt.Errorf("checksum mismatch")
		}
	}

	return safeData[32:], nil
}
