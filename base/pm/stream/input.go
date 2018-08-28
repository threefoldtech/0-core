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
	pmMsgPattern  = regexp.MustCompile("^(\\d+)(:{2,3})(.*)$")
	headerPattern = regexp.MustCompile(`^(\d+)(:{2,3})`)

	errNotHeader = fmt.Errorf("not header")
)

type consumerImpl struct {
	level   uint16
	handler MessageHandler

	multi *Message

	source io.Reader
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

		if err := c.consume(); err != nil {
			log.Errorf("failed to read stream: %s", err)
		}

		source.Close()
	}()
}

func (c *consumerImpl) isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func (c *consumerImpl) parseHead(head []byte) (*header, error) {
	//parse the header without regex
	//valid pattern is
	// `dd?:::?`
	if len(head) < 3 {
		return nil, errNotHeader //noway this is a header
	}

	skip := bytes.Index(head, []byte("::"))

	if skip > 2 || skip <= 0 {
		return nil, errNotHeader
	}

	level, _ := strconv.ParseUint(string(head[0:skip]), 10, 16)
	h := header{
		level: uint16(level),
	}

	if bytes.HasPrefix(head[skip:], []byte(":::")) {
		//multi line
		h.multiline = true
		h.length = skip + 3
	} else if bytes.HasPrefix(head[skip:], []byte("::")) {
		h.length = skip + 2
	} else {
		return nil, errNotHeader
	}

	return &h, nil
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

// func (c *consumerImpl) readLine(out io.Writer) error {
// 	line, err := c.buffer.ReadBytes('\n')
// 	if err != nil && err != io.EOF {
// 		return err
// 	}
// 	_, err = out.Write(line)
// 	return err
// }

//newLineOrEOF will return index of the next \n or EOF (end of file or string)
func (c *consumerImpl) newLineOrEOF(b []byte) int {
	i := 0
	for ; i < len(b) && b[i] != '\n'; i++ {
	}

	return i
}

//process process the buffer and return th
func (c *consumerImpl) process(buffer []byte) int {
	start := 0
	for i := 0; i < len(buffer); i++ {
		//fmt.Printf("trying: %s\n", string(buffer[i:]))
		m := headerPattern.FindSubmatch(buffer[i:])
		if m == nil {
			//no header was found at this position
			i += c.newLineOrEOF(buffer[i+1:]) + 1
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
			j := i + h.length               //start of text after the header
			j += c.newLineOrEOF(buffer[j:]) //seek to new line or end of text

			c.handler(&Message{
				Meta:    NewMeta(h.level),
				Message: string(buffer[i+h.length : j]),
			})
			i = j + 1
			start = i
			//TODO: what if the same line output is split !!
			//I think it's better if the single line message must end with new line, in that case we
			//need to seek only to new line termination, if not found we save the current state and
			//wait for the next feedback
			continue
		}

		//TODO:
		//if we are here we must find the \n::: termination string
	}

	if start < len(buffer) {
		c.handler(&Message{
			Meta:    NewMeta(c.level),
			Message: string(buffer[start:]),
		})
	}

	return 0
}

func (c *consumerImpl) consume() error {
	buffer := make([]byte, ioBufferSize)
	offset := 0
	for {
		size, err := c.source.Read(buffer[offset:]) //fill what left of the buffer
		if err == io.EOF {
			return nil
		}

		all := offset + size
		reminder := c.process(buffer[:all])
		if reminder != 0 {
			//if some data remains at the end of the buffer that
			//wasn't able to be processed. we need to move it at
			//the head of the buffer
			copy(buffer, buffer[all-reminder:all])
			offset = reminder
		}
	}
}

func (c *consumerImpl) processLine(line string) {
	line = strings.TrimRight(line, "\n")

	if c.multi != nil {
		if line == ":::" {
			//last, flush mult
			c.handler(c.multi)
			c.multi = nil
			return
		}

		c.multi.Message += "\n" + line
		return
	}

	matches := pmMsgPattern.FindStringSubmatch(line)

	if matches == nil {
		//use default level.
		c.handler(&Message{
			Meta:    NewMeta(c.level),
			Message: line,
		})
	} else {
		l, _ := strconv.ParseUint(matches[1], 10, 16)
		level := uint16(l)
		message := matches[3]

		if matches[2] == ":::" {
			c.multi = &Message{
				Meta:    NewMeta(level),
				Message: message,
			}
		} else {
			//single line message
			c.handler(&Message{
				Meta:    NewMeta(level),
				Message: message,
			})
		}
	}

}
