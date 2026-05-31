package main

import (
	"bytes"
	"encoding/base64"
	"log"
	"sync"

	"golang.design/x/clipboard"
)

type oscExecutor struct {
	inputDataSender dataSender
	clipboardWriter clipboardWriter
	parser          parser
}

var _ shellInterceptor = (*oscExecutor)(nil)

const maxOSC52PayloadSize = 1024 * 1024

func (oe *oscExecutor) Init(inputDataSender, outputDataSender dataSender) *oscExecutor {
	oe.inputDataSender = inputDataSender
	if oe.clipboardWriter == nil {
		oe.clipboardWriter = new(nativeClipboardWriter)
	}
	oe.parser.Init(escapeSequenceBegin, escapeSequenceEnd, oe.handleDataToCopy, dataHandler(outputDataSender)).
		SetMaxCapturedDataSize(maxOSC52PayloadSize)
	return oe
}

func (oe *oscExecutor) HandleInputData(data []byte) bool {
	return oe.inputDataSender(data)
}

func (oe *oscExecutor) HandleOutputData(data []byte) bool {
	return oe.parser.FeedData(data)
}

func (oe *oscExecutor) handleDataToCopy(data []byte) bool {
	if err := setClipboard(oe.clipboardWriter, data); err != nil {
		log.Printf("set clipboard failed: %v", err)
	}

	return true
}

var (
	escapeSequenceBegin = []byte("\x1b]52;")
	escapeSequenceEnd   = []byte("\x07")
)

type clipboardWriter interface {
	WriteText([]byte) error
}

type nativeClipboardWriter struct {
	initOnce sync.Once
	initErr  error
}

func (ncw *nativeClipboardWriter) WriteText(data []byte) error {
	ncw.initOnce.Do(func() {
		ncw.initErr = clipboard.Init()
	})

	if ncw.initErr != nil {
		return ncw.initErr
	}

	clipboard.Write(clipboard.FmtText, data)
	return nil
}

func setClipboard(writer clipboardWriter, rawData []byte) error {
	// At this point, string will still be prepended by command and separator.
	// https://invisible-island.net/xterm/ctlseqs/ctlseqs.html
	//        The first, Pc, may contain zero or more characters from the
	//        set c , p , q , s , 0 , 1 , 2 , 3 , 4 , 5 , 6 , and 7 .  It is
	//        used to construct a list of selection parameters for
	//        clipboard, primary, secondary, select, or cut-buffers 0
	//        through 7 respectively, in the order given.  If the parameter
	//        is empty, xterm uses s 0 , to specify the configurable
	//        primary/clipboard selection and cut-buffer 0.
	// (thank you https://github.com/tmux/tmux/issues/4847#issuecomment-3863645137)

	if idx := bytes.IndexByte(rawData, ';'); idx != -1 {
		rawData = rawData[idx+1:]
	}

	buffer := make([]byte, base64.StdEncoding.DecodedLen(len(rawData)))
	n, err := base64.StdEncoding.Decode(buffer, rawData)

	if err != nil {
		return err
	}

	data := buffer[:n]
	return writer.WriteText(data)
}
