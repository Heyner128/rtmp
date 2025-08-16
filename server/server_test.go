package server

import (
	"math/rand"
	"net"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

var address = "127.0.0.1:" + strconv.Itoa(1000+rand.Intn(9999-1000))

func TestMain(m *testing.M) {
	go Listen(address)
	m.Run()
}

func TestStartServer(t *testing.T) {
	_, err := net.Dial("tcp", address)
	assert.Nil(t, err)
}

func TestStartServerFail(t *testing.T) {
	var invalidAddress = "notanip:0000"
	go Listen(invalidAddress)
	_, err := net.Dial("tcp", invalidAddress)
	assert.NotNil(t, err)
}
