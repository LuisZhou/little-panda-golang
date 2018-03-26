package main

import (
	"encoding/binary"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:3563")
	if err != nil {
		panic(err)
	}

	// Hello message (JSON-encoded)
	// The structure of the message
	data := []byte(`{
        "Hello": {
            "Name": "leaf"
        }
    }`)

	// len + data
	m := make([]byte, 2+len(data))

	// BigEndian encoded
	binary.BigEndian.PutUint16(m, uint16(len(data)))

	copy(m[2:], data)

	// Send message
	conn.Write(m)
}
