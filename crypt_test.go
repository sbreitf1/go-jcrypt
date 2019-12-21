package jcrypt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testData struct {
	RawValue string `json:"raw"`
	Pass     string `json:"pass" jcrypt:"aes"`
	Number   int    `json:"num"`
	Ignored  bool   `json:"-"`
}

func TestMarshalRaw(t *testing.T) {
	input := `{"raw":"just a string","pass":"secret","Ignored":true}`
	expected := testData{"just a string", "secret", 0, false}
	var d testData
	assert.NoError(t, Unmarshal([]byte(input), &d, nil))
	assert.Equal(t, expected, d)
}
