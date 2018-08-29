package stream

import (
	"io"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testReader struct {
	chunks []string
	index  int
}

func (t *testReader) Read(p []byte) (n int, err error) {
	if len(t.chunks) == t.index {
		return 0, io.EOF
	}

	s := t.chunks[t.index]
	copy(p, []byte(s))
	t.index++
	if len(t.chunks) == t.index {
		return len(s), io.EOF
	}

	return len(s), nil
}

func (_ *testReader) Close() error {
	return nil
}

func TestConsumer_newLineOrEOF(t *testing.T) {
	c := consumerImpl{}
	s := "hello world"
	x := c.newLineOrEOF([]byte(s))
	if ok := assert.Equal(t, 11, x); !ok {
		t.Fail()
	}

	s = "hello\nworld"
	x = c.newLineOrEOF([]byte(s))
	if ok := assert.Equal(t, 5, x); !ok {
		t.Fail()
	}

	b := []byte("hello\ngood\nworld")
	var r []string
	for i := 0; i < len(b); i++ {
		n := c.newLineOrEOF(b[i:])
		r = append(r, string(b[i:i+n]))
		i += n
	}

	if ok := assert.Equal(t, []string{"hello", "good", "world"}, r); !ok {
		t.Fail()
	}
}

func TestConsumer_processNormalText(t *testing.T) {
	var message *Message
	h := func(m *Message) {
		message = m
	}

	c := consumerImpl{
		level:   1,
		handler: h,
	}

	c.process([]byte("hello world"))

	if ok := assert.NotNil(t, message); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, "hello world", message.Message); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, uint16(1), message.Meta.Level()); !ok {
		t.Fatal()
	}

	message = nil
	txt := "hello world\nthis is output of some program\nthat spans many lines\n"
	c.process([]byte(txt))

	if ok := assert.NotNil(t, message); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, txt, message.Message); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, uint16(1), message.Meta.Level()); !ok {
		t.Fatal()
	}
}

func TestConsumer_processSingleLineMessage(t *testing.T) {
	var messages []*Message
	h := func(m *Message) {
		messages = append(messages, m)
	}

	c := consumerImpl{
		level:   1,
		handler: h,
	}

	c.process([]byte(`hello world
the folowing line is a single line message
2::this is a single line message`))

	if ok := assert.Len(t, messages, 2); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, "hello world\nthe folowing line is a single line message\n", messages[0].Message); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, uint16(1), messages[0].Meta.Level()); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, "this is a single line message", messages[1].Message); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, uint16(2), messages[1].Meta.Level()); !ok {
		t.Error()
	}

	messages = nil

	c.process([]byte(`hello world
the folowing line is a single line message
2::this is a single line message
followed by some more text
that spans multiple lines`))

	if ok := assert.Len(t, messages, 3); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, "hello world\nthe folowing line is a single line message\n", messages[0].Message); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, uint16(1), messages[0].Meta.Level()); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, "this is a single line message\n", messages[1].Message); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, uint16(2), messages[1].Meta.Level()); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, "followed by some more text\nthat spans multiple lines", messages[2].Message); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, uint16(1), messages[2].Meta.Level()); !ok {
		t.Error()
	}

	messages = nil

	c.process([]byte(`2::this is a single line message
3::followed by some more messages
4::that has some message`))

	if ok := assert.Len(t, messages, 3); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, "this is a single line message\n", messages[0].Message); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, uint16(2), messages[0].Meta.Level()); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, "followed by some more messages\n", messages[1].Message); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, uint16(3), messages[1].Meta.Level()); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, "that has some message", messages[2].Message); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, uint16(4), messages[2].Meta.Level()); !ok {
		t.Error()
	}
}

func TestConsumer_processSingleLineMessageMultiBlocks(t *testing.T) {
	var messages []*Message
	h := func(m *Message) {
		messages = append(messages, m)
	}

	c := consumerImpl{
		level:   1,
		handler: h,
	}

	c.process([]byte(`hello world
the folowing line is a single line message
2::this is a single line message`))

	c.process([]byte(`followed by some more text
that spans multiple lines`))

	if ok := assert.Len(t, messages, 3); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, "hello world\nthe folowing line is a single line message\n", messages[0].Message); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, uint16(1), messages[0].Meta.Level()); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, "this is a single line message", messages[1].Message); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, uint16(2), messages[1].Meta.Level()); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, "followed by some more text\nthat spans multiple lines", messages[2].Message); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, uint16(1), messages[2].Meta.Level()); !ok {
		t.Error()
	}
}

func TestConsumer_processMultiLineMessage(t *testing.T) {
	var messages []*Message
	h := func(m *Message) {
		messages = append(messages, m)
	}

	c := consumerImpl{
		level:   1,
		handler: h,
	}

	c.process([]byte(`3:::multi line message
with full termination
in one block
:::
`))

	if ok := assert.Len(t, messages, 1); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, "multi line message\nwith full termination\nin one block", messages[0].Message); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, uint16(3), messages[0].Meta.Level()); !ok {
		t.Error()
	}

	messages = nil
	c.process([]byte(`a multi line block is coming
3:::multi line message
with full termination
in one block
:::
which is surrounded by normal text`))

	if ok := assert.Len(t, messages, 3); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, "a multi line block is coming\n", messages[0].Message); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, uint16(1), messages[0].Meta.Level()); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, "multi line message\nwith full termination\nin one block", messages[1].Message); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, uint16(3), messages[1].Meta.Level()); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, "which is surrounded by normal text", messages[2].Message); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, uint16(1), messages[2].Meta.Level()); !ok {
		t.Error()
	}
}

func TestConsumer_processMultiLineMessageMultiBlock(t *testing.T) {
	var messages []*Message
	h := func(m *Message) {
		messages = append(messages, m)
	}

	c := consumerImpl{
		level:   1,
		handler: h,
	}

	c.process([]byte(`a multi line block is coming
30:::multi line message
with full termination`))
	c.process([]byte(`
in two blocks
:::
which is surrounded by normal text`))

	if ok := assert.Len(t, messages, 3); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, "a multi line block is coming\n", messages[0].Message); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, uint16(1), messages[0].Meta.Level()); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, "multi line message\nwith full termination\nin two blocks", messages[1].Message); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, uint16(30), messages[1].Meta.Level()); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, "which is surrounded by normal text", messages[2].Message); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, uint16(1), messages[2].Meta.Level()); !ok {
		t.Error()
	}
}

func TestConsumer_OneLine(t *testing.T) {
	var message *Message
	h := func(m *Message) {
		message = m
	}

	var wg sync.WaitGroup
	wg.Add(1)
	Consume(&wg, &testReader{
		chunks: []string{"hello world\n"},
	}, 1, h)

	wg.Wait()

	if ok := assert.NotNil(t, message); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, "hello world\n", message.Message); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, uint16(1), message.Meta.Level()); !ok {
		t.Fatal()
	}
}

func TestConsumer_TwoLines(t *testing.T) {
	var messages []*Message
	h := func(m *Message) {
		messages = append(messages, m)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	Consume(&wg, &testReader{
		chunks: []string{
			"hello world\n",
			"10::bye bye world\n",
		},
	}, 1, h)

	wg.Wait()

	if ok := assert.Len(t, messages, 2); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, "hello world\n", messages[0].Message); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, uint16(1), messages[0].Meta.Level()); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, "bye bye world\n", messages[1].Message); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, uint16(10), messages[1].Meta.Level()); !ok {
		t.Error()
	}
}

func TestConsumer_MultiLine(t *testing.T) {
	var messages []*Message
	h := func(m *Message) {
		messages = append(messages, m)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	Consume(&wg, &testReader{
		chunks: []string{
			"30:::hello\nworld",
			"\n:::\n",
		},
	}, 1, h)

	wg.Wait()

	if ok := assert.Len(t, messages, 1); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, "hello\nworld", messages[0].Message); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, uint16(30), messages[0].Meta.Level()); !ok {
		t.Error()
	}
}

func TestConsumer_Complex(t *testing.T) {
	var messages []*Message
	h := func(m *Message) {
		messages = append(messages, m)
	}

	chunk1 := `Hello world
20::this is a single line message
30:::but this is a multi line
that spans`

	chunk2 := ` two blocks of data
:::
`
	var wg sync.WaitGroup
	wg.Add(1)
	Consume(&wg, &testReader{
		chunks: []string{
			chunk1,
			chunk2,
		},
	}, 2, h)

	wg.Wait()

	if ok := assert.Len(t, messages, 3); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, "Hello world\n", messages[0].Message); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, uint16(2), messages[0].Meta.Level()); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, "this is a single line message\n", messages[1].Message); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, uint16(20), messages[1].Meta.Level()); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, "but this is a multi line\nthat spans two blocks of data", messages[2].Message); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, uint16(30), messages[2].Meta.Level()); !ok {
		t.Error()
	}
}

func TestConsumerNoNewLine(t *testing.T) {
	var messages []*Message
	h := func(m *Message) {
		messages = append(messages, m)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	Consume(&wg, &testReader{
		chunks: []string{
			"hello world",
		},
	}, 1, h)

	wg.Wait()

	if ok := assert.Len(t, messages, 1); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, "hello world", messages[0].Message); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, uint16(1), messages[0].Meta.Level()); !ok {
		t.Fatal()
	}
}
