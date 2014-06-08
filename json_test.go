package bayeux

import (
	_ "bytes"
	"encoding/json"
	"testing"
)

type A struct {
	First string `json:"first"`
}

type B struct {
	A
	Last string `json:"last"`
}

func TestExtenedEncoding(t *testing.T) {
	out, err := Encode(&B{A{"Eric"}, "Bittleman"})

	if err != nil {
		t.Error(err)
	}

	t.Log(out)
}

func Encode(a *A) (string, error) {
	data, err := json.Marshal(a)

	return string(data), err
}
