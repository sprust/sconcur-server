package connection

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sconcur/internal/services/dto"
	"sconcur/pkg/foundation/errs"
)

type Transport struct {
	conn net.Conn
}

func NewTransport(conn net.Conn) *Transport {
	return &Transport{
		conn: conn,
	}
}

func (t *Transport) IsConnected() bool {
	buffer := make([]byte, 1024)

	_, err := t.conn.Read(buffer)

	if err != nil {
		return false
	}

	return true
}

func (t *Transport) ReadMessage() (*dto.Message, error) {
	sizeBuf := make([]byte, 4)

	_, err := io.ReadFull(t.conn, sizeBuf)

	if err != nil {
		return nil, errs.Err(err)
	}

	size := int(sizeBuf[0])<<24 | int(sizeBuf[1])<<16 | int(sizeBuf[2])<<8 | int(sizeBuf[3])

	payload := make([]byte, size)
	bytesRead := 0

	const bufferSize = 32 * 1024 // 32KB buffer
	buffer := make([]byte, bufferSize)
	bytesRemaining := int64(size)

	for bytesRemaining > 0 {
		bytesToRead := int64(bufferSize)

		if bytesToRead > bytesRemaining {
			bytesToRead = bytesRemaining
		}

		readBytes, err := t.conn.Read(buffer[:bytesToRead])

		if err != nil {
			return nil, errs.Err(err)
		}

		copy(payload[bytesRead:], buffer[:readBytes])
		bytesRead += readBytes
		bytesRemaining -= int64(readBytes)
	}

	var message dto.Message

	err = json.Unmarshal(payload, &message)

	if err != nil {
		return nil, errs.Err(err)
	}

	return &message, nil
}

func (t *Transport) WriteResult(result *dto.Result) error {
	payload, err := json.Marshal(result)

	if err != nil {
		return errs.Err(err)
	}

	dataLength := uint32(len(payload))

	lengthBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBuf, dataLength)

	fmt.Println(uint32(len(payload)))
	fmt.Println(string(dataLength) + string(payload))

	_, err = t.conn.Write(
		[]byte(
			string(lengthBuf) + string(payload),
		),
	)

	return err
}

func (t *Transport) Close() error {
	if t.conn == nil {
		return nil
	}

	return t.conn.Close()
}
