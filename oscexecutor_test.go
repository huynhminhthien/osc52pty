package main

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeClipboardWriter struct {
	data []byte
}

func (fcw *fakeClipboardWriter) WriteText(data []byte) error {
	fcw.data = append(fcw.data[:0], data...)
	return nil
}

func TestSetClipboard(t *testing.T) {
	text := "hello clipboard\n"
	encodedText := base64.StdEncoding.EncodeToString([]byte(text))
	writer := new(fakeClipboardWriter)

	err := setClipboard(writer, []byte(encodedText))
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.Equal(t, text, string(writer.data))
}

func TestSetClipboardStripsSelectionParameter(t *testing.T) {
	text := "hello clipboard\n"
	encodedText := base64.StdEncoding.EncodeToString([]byte(text))
	writer := new(fakeClipboardWriter)

	err := setClipboard(writer, []byte("c;"+encodedText))
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.Equal(t, text, string(writer.data))
}
