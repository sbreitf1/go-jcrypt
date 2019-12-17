package jcrypt

import "testing"

type testData struct {
	RawValue string `json:"raw"`
	Pass     string `json:"pass" jcrypt:"des"`
	Number   int    `json:"num"`
	Ignored  bool   `json:-`
}

func TestMarshal(t *testing.T) {

}
