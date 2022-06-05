package memcached

import "fmt"

type MemCachedFlags []byte

func (c MemCachedFlags) String() string {
	return string(c)
}

var (
	End      MemCachedFlags = []byte("END\r\n")
	Stored   MemCachedFlags = []byte("STORED\r\n")
	Deleted  MemCachedFlags = []byte("DELETED\r\n")
	NotFound MemCachedFlags = []byte("NOT_FOUND\r\n")
	Error    MemCachedFlags = []byte("ERROR\r\n")
)

type CommandName string

func (c CommandName) String() string {
	return string(c)
}

const (
	get       CommandName = "get"
	set       CommandName = "set"
	deleteKey CommandName = "delete"
)

type Command struct {
	Key     string
	Value   []byte
	Flags   int
	Exptime int
	Bytes   int
}

func (c Command) BuildSet() string {
	return fmt.Sprintf("%s %s %d %d %d\n",
		set, c.Key, c.Flags, c.Exptime, len(c.Value))
}
func (c Command) BuildGet() string {
	return fmt.Sprintf("%s %s\n",
		get, c.Key)
}
func (c Command) BuildDelete() string {
	return fmt.Sprintf("%s %s\n",
		deleteKey, c.Key)
}
