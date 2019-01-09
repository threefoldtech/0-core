package stream

import (
	"fmt"
)

type Flag uint16

const (
	//StreamFlag Job is running in stream
	StreamFlag Flag = 1 << iota
	//ExitSuccessFlag success
	ExitSuccessFlag
	//ExitErrorFlag error
	ExitErrorFlag
)

type Meta uint64

func NewMeta(level uint16, flag ...Flag) Meta {
	m := uint32(level) << 16
	for _, f := range flag {
		m |= uint32(f)
	}

	return Meta(m)
}

func NewMetaWithCode(code uint32, level uint16, flag ...Flag) Meta {
	meta := NewMeta(level, flag...)
	return (Meta(code) << 32) | meta
}

func (m Meta) Level() uint16 {
	return uint16(uint64(m) >> 16 & 0xffff)
}

func (m Meta) Assert(level ...uint16) bool {
	l := uint16(uint64(m) >> 16 & 0xffff)
	for _, lv := range level {
		if l == lv {
			return true
		}
	}

	return false
}

//Is checks if a flag is set on the meta object
func (m Meta) Is(flag Flag) bool {
	return (uint16(m) & uint16(flag)) != 0
}

//Set sets a flag on meta object
func (m Meta) Set(flag Flag) Meta {
	return Meta(uint64(m) | uint64(flag))
}

//Code exit code
func (m Meta) Code() uint32 {
	return uint32(m >> 32)
}

//Base gets meta without the code part (used for backward compatibility)
func (m Meta) Base() Meta {
	return m & 0xffffffff
}

//Message is a message from running process
type Message struct {
	Message string `json:"message"`
	Epoch   int64  `json:"epoch"`
	Meta    Meta   `json:"meta"`
}

//MessageHandler represents a callback type
type MessageHandler func(*Message)

//String represents a message as a string
func (msg *Message) String() string {
	return fmt.Sprintf("%d|%s", msg.Meta.Level(), msg.Message)
}
