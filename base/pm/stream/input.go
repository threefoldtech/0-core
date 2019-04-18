package stream

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"sync"

	logging "github.com/op/go-logging"
)

const (
	ioBufferSize = 32 * 1024
	//messageBufferSize = 10 * 1024
)

var (
	log           = logging.MustGetLogger("stream")
	headerPattern = regexp.MustCompile(`^(\d+)(:{2,3})`)

	multiLineTerm = "\n:::\n"
	// multiLineHead = ":::\n"

	errNotHeader = fmt.Errorf("not header")
)

type consumerImpl struct {
	level   uint16
	handler MessageHandler
	source  io.Reader

	multi      *bytes.Buffer
	multiLevel uint16
}

type header struct {
	level     uint16
	multiline bool
	length    int
}

//Consume consumes a stream to the end, and calls the handler with the parsed stream messages
func Consume(wg *sync.WaitGroup, source io.ReadCloser, level uint16, handler MessageHandler) {
	c := &consumerImpl{
		level:   level,
		handler: handler,
		source:  source,
	}

	//source.Read
	go func() {
		if wg != nil {
			defer wg.Done()
		}

		defer source.Close()

		if err := c.consume(); err != nil {
			log.Errorf("failed to read stream: %s", err)
		}
	}()
}

func (c *consumerImpl) getHeaderFromMatch(m [][]byte) *header {
	level, _ := strconv.ParseUint(string(m[1]), 10, 16)
	h := header{
		level: uint16(level),
	}
	if len(m[2]) == 3 {
		h.multiline = true
	}
	h.length = len(m[0])
	return &h
}

//newLineOrEOF will return index of the next \n or EOF (end of file or string)
func (c *consumerImpl) newLineOrEOF(b []byte) int {
	var i int
	var x byte
	for i, x = range b {
		if x == '\n' {
			break
		}
	}
	if i+1 == len(b) {
		i += 1
	}
	return i
}

//process process the buffer and return th
func (c *consumerImpl) process(buffer []byte) {
	if c.multi != nil {
		//we are in a middle of a multi line message
		if c.multi.Len() > 0 && c.multi.Bytes()[c.multi.Len()-1] == '\n' && bytes.HasPrefix(buffer, []byte(":::\n")) {
			c.handler(&Message{
				Meta:    NewMeta(c.multiLevel),
				Message: c.multi.String(),
			})
			c.multi = nil
			buffer = buffer[4:]
		} else if end := strings.Index(string(buffer), multiLineTerm); end != -1 {
			//we found the termination string
			c.multi.Write(buffer[:end])
			c.handler(&Message{
				Meta:    NewMeta(c.multiLevel),
				Message: c.multi.String(),
			})
			c.multi = nil
			buffer = buffer[end+len(multiLineTerm):]
		} else {
			c.multi.Write(buffer)
			return
		}
	}

	start := 0
	for i := 0; i < len(buffer); i++ {
		m := headerPattern.FindSubmatch(buffer[i:])
		if m == nil {
			//no header was found at this position
			i += c.newLineOrEOF(buffer[i:])
			continue
		}

		//if we reach here then we can safely flush what we have in buffer as a message
		if i > start {
			c.handler(&Message{
				Meta:    NewMeta(c.level),
				Message: string(buffer[start:i]),
			})
		}

		h := c.getHeaderFromMatch(m)
		if !h.multiline {
			//find next new line or end of line
			start = i + h.length                       //start of text after the header
			i = start + c.newLineOrEOF(buffer[start:]) //seek to new line or end of text

			var msg string
			if i == len(buffer) {
				msg = string(buffer[start:i])
			} else {
				msg = string(buffer[start : i+1]) //include the new line
			}
			c.handler(&Message{
				Meta:    NewMeta(h.level),
				Message: msg,
			})

			start = i + 1

			//TODO: what if the same line output is split !!

			//I think it's better if the single line message must end with new line, in that case we
			//need to seek only to new line termination, if not found we save the current state and
			//wait for the next feedback
			continue
		}

		log.Debugf("starting multiline")
		//multiline message
		//read in multi until eof or \n:::\n termination
		j := i + h.length
		if end := strings.Index(string(buffer[j:]), multiLineTerm); end != -1 {
			//we found the termination string
			c.handler(&Message{
				Meta:    NewMeta(h.level),
				Message: string(buffer[j : j+end]),
			})
			i = j + end + len(multiLineTerm) // 4 is the width of the termination string
			start = i
		} else {
			c.multiLevel = h.level
			c.multi = bytes.NewBuffer(nil)
			c.multi.Write(buffer[j:])
			start = len(buffer)
			break
		}
	}

	if start < len(buffer) {
		c.handler(&Message{
			Meta:    NewMeta(c.level),
			Message: string(buffer[start:]),
		})
	}
}

func (c *consumerImpl) consume() error {
	buffer := make([]byte, ioBufferSize)

	for {
		size, err := c.source.Read(buffer) //fill what left of the buffer
		if err != nil && err != io.EOF {
			return err
		}
		c.process(buffer[:size])

		if err == io.EOF {
			return nil
		}
	}
}
