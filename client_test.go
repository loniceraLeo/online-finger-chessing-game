package main

import (
	"testing"
)

func TestMsg(t *testing.T) {
	msg := doCreatePlayRoom()
	t.Log(msg)
}
