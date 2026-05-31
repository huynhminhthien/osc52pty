package main

import (
	"encoding/base64"
	"strings"
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

func TestOSCExecutorDropsOversizedClipboardSequence(t *testing.T) {
	writer := new(fakeClipboardWriter)
	var outputData []byte
	outputDataSender := func(data []byte) bool {
		outputData = append(outputData, data...)
		return true
	}

	oe := &oscExecutor{clipboardWriter: writer}
	oe.Init(nil, outputDataSender)

	validText := "ok"
	validEncodedText := base64.StdEncoding.EncodeToString([]byte(validText))
	oversizedText := strings.Repeat("A", maxOSC52PayloadSize+1)
	data := append([]byte("before"), escapeSequenceBegin...)
	data = append(data, []byte("c;"+oversizedText)...)
	data = append(data, escapeSequenceEnd...)
	data = append(data, []byte("after")...)
	data = append(data, escapeSequenceBegin...)
	data = append(data, []byte("c;"+validEncodedText)...)
	data = append(data, escapeSequenceEnd...)

	ok := oe.HandleOutputData(data)

	assert.True(t, ok)
	assert.Equal(t, "beforeafter", string(outputData))
	assert.Equal(t, validText, string(writer.data))
}
