package vault

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_AesEncrypt(t *testing.T) {
	ve := NewVaultElement()
	ve.MetaKey = []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")

	assert.Equal(t, 32, len(ve.MetaKey))

	src := bytes.NewBuffer([]byte("Hello World!!"))
	dst := bytes.NewBuffer([]byte(""))

	AesOFBEncrypter(ve.MetaKey, src, dst)

	assert.NotEmpty(t, dst.String())

	decrypted := bytes.NewBuffer([]byte(""))

	AesOFBDecrypter(ve.MetaKey, dst, decrypted)

	assert.Equal(t, []byte("Hello World!!"), decrypted.Bytes())
}
