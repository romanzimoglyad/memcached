package memcached

import (
	"bytes"
	"fmt"
	"github.com/romanzimoglyad/memcached/internal/model"
	"io"
	"strings"
)

type Client struct {
	*TcpConnPool
}

func NewClient(tcpConnPool *TcpConnPool) (*Client, error) {

	return &Client{

		TcpConnPool: tcpConnPool,
	}, nil
}
func (c *Client) Get(keys ...string) (out []model.Record, err error) {
	conn, err := c.get()

	if err != nil {
		return out, err
	}
	command := Command{
		Key: strings.Join(keys, " "),
	}

	_, err = conn.rw.WriteString(command.BuildGet())
	if err != nil {
		return out, err
	}
	if err := conn.rw.Flush(); err != nil {
		return out, err
	}
	var result string
	for !strings.EqualFold(result, End.String()) {
		line, err := conn.rw.ReadSlice('\n')
		if bytes.Equal(line, End) {
			break
		}

		if err != nil {
			return out, err
		}
		var sz int
		dest := []interface{}{&command.Key, &command.Flags, &sz}
		pattern := "VALUE %s %d %d\r\n"
		_, err = fmt.Sscanf(string(line), pattern, dest...)
		if err != nil {
			return out, err
		}
		command.Value = make([]byte, sz+2)
		_, err = io.ReadFull(conn.rw, command.Value)
		if err != nil {
			return out, err
		}
		out = append(out, model.Record{
			Key:   command.Key,
			Value: string(command.Value[:len(command.Value)-2]),
		})
	}

	if err != nil {
		return out, err
	}
	c.put(conn)
	return out, nil
}

func (c *Client) Set(record model.Record) error {
	if record.Key == "" {
		return fmt.Errorf("key shouldn't be empty")
	}
	conn, err := c.get()

	if err != nil {
		return err
	}
	command := Command{

		Key:     record.Key,
		Value:   []byte(record.Value),
		Flags:   0,
		Exptime: 0,
		Bytes:   len(record.Value),
	}

	_, err = conn.rw.WriteString(command.BuildSet())
	if err != nil {
		return err
	}
	if _, err = conn.rw.Write(command.Value); err != nil {
		return err
	}
	if _, err := conn.rw.Write([]byte("\r\n")); err != nil {
		return err
	}
	if err := conn.rw.Flush(); err != nil {
		return err
	}
	line, err := conn.rw.ReadSlice('\n')

	if err != nil {
		return err
	}
	if !bytes.Equal(line, Stored) {
		return fmt.Errorf("Bad  status from memCached: %v", string(line))
	}
	c.put(conn)
	return nil
}

func (c *Client) Delete(key string) error {
	conn, err := c.get()
	if err != nil {
		return err
	}
	command := Command{
		Key: key,
	}
	_, err = conn.rw.WriteString(command.BuildDelete())
	if err != nil {
		return err
	}
	if err := conn.rw.Flush(); err != nil {
		return err
	}
	line, err := conn.rw.ReadSlice('\n')
	if err != nil {
		return err
	}
	switch {
	case bytes.Equal(line, Deleted):
		c.put(conn)
		return nil
	case bytes.Equal(line, NotFound):
		return fmt.Errorf("record with key %s not found", key)
	case bytes.Equal(line, Error):
		return fmt.Errorf("error while delete record with key %s ", key)
	default:
		return fmt.Errorf("unknown error")
	}
}
