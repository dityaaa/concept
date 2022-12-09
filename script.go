package concept

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
)

type Direction string

const (
	AdvanceDirection = "ADV"
	ReverseDirection = "REV"
)

type Script struct {
	Version     string
	Identifier  string
	Description string
	Direction   Direction

	content  io.ReadCloser
	checksum string
}

func (i *Script) Read(p []byte) (n int, err error) {
	return i.content.Read(p)
}

func (i *Script) Close() error {
	return i.content.Close()
}

func (i *Script) SetContent(rd io.ReadCloser) {
	i.content = rd
}

func (i *Script) Checksum() string {
	if i.checksum != "" {
		return i.checksum
	}

	rawContent, err := io.ReadAll(i.content)
	if err != nil {
		panic(err)
	}

	// does not use defer because i.content is replaced by NopCloser
	err = i.Close()
	if err != nil {
		panic(err)
	}

	i.content = io.NopCloser(bytes.NewReader(rawContent))
	i.checksum = fmt.Sprintf("%x", md5.Sum(rawContent))

	return i.checksum
}
