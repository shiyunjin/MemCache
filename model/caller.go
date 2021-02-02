package mm

import (
	"github.com/golang/protobuf/proto"
)

type CacheCaller func() (proto.Message, error)

type IMemoryCache interface {
	Get(key string, value proto.Message, caller CacheCaller, expireSeconds int) error
}
