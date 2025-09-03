package chunk

import (
	"encoding/binary"
	"math"
	"math/rand"
	"miistream/rtmpconn"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var address = "127.0.0.1:" + strconv.Itoa(1000+rand.Intn(9999-1000))

var listener, _ = net.Listen("tcp", address)

var rtmpConn = rtmpconn.RtmpConn{
	MaxChunkSize:   128,
	NetworkTimeout: 10 * time.Second,
}

var chunkProcessed = make(chan bool)

func TestMain(m *testing.M) {
	defer listener.Close()
	go func() {
		for {
			conn, _ := listener.Accept()
			rtmpConn.Conn = conn
			err := Accept(&rtmpConn)
			if err != nil {
				panic(err)
			}
			chunkProcessed <- true
		}
	}()
	m.Run()
}

func TestType0Chunk(t *testing.T) {
	conn, _ := net.Dial("tcp", address)
	chunk := Chunk{
		Header{
			BasicHeader{
				uint8(0),
				uint32(2),
			},
			MessageHeader{
				uint32(0),
				uint32(32),
				uint8(1),
				uint32(0),
			},
		},
		binary.BigEndian.AppendUint32(make([]byte, 0), uint32(256)),
	}
	_, err := conn.Write(chunk.Buffer())
	assert.Nil(t, err)
	<-chunkProcessed
}

func TestType1Chunk(t *testing.T) {
	conn, _ := net.Dial("tcp", address)
	chunk := Chunk{
		Header{
			BasicHeader{
				uint8(1),
				uint32(2),
			},
			MessageHeader{
				uint32(0),
				uint32(32),
				uint8(1),
				uint32(0),
			},
		},
		binary.BigEndian.AppendUint32(make([]byte, 0), uint32(256)),
	}
	_, err := conn.Write(chunk.Buffer())
	assert.Nil(t, err)
	<-chunkProcessed
}

func TestChunkExtendedTimestamp(t *testing.T) {
	conn, _ := net.Dial("tcp", address)
	chunk := Chunk{
		Header{
			BasicHeader{
				uint8(0),
				uint32(2),
			},
			MessageHeader{
				math.MaxUint32,
				uint32(32),
				uint8(1),
				uint32(0),
			},
		},
		binary.BigEndian.AppendUint32(make([]byte, 0), uint32(256)),
	}
	_, err := conn.Write(chunk.Buffer())
	assert.Nil(t, err)
	<-chunkProcessed
}
