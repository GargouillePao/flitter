package flitter

import (
	"testing"
)

type Hello struct {
	intval int
}

func TestMessageProcessor(t *testing.T) {
	p := &MessageProcessor{}
	data := Hello{}
	p.Register("Hello", data)
	t.Fatalf("Should Be Hello")
}
