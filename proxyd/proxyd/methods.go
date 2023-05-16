package proxyd

import (
	"context"
	"encoding/json"
	"github.com/ethereum/go-ethereum/log"
	"strings"
	"sync"
)

type RPCMethodHandler interface {
	GetRPCMethod(context.Context, *RPCReq) (*RPCRes, error)
	PutRPCMethod(context.Context, *RPCReq, *RPCRes) error
}

type StaticMethodHandler struct {
	cache Cache
	m     sync.RWMutex
}

func (e *StaticMethodHandler) key(req *RPCReq) string {
	// signature is a cache-friendly base64-encoded string with json.RawMessage param contents
	signature := string(req.Params)
	return strings.Join([]string{"cache", req.Method, signature}, ":")
}

func (e *StaticMethodHandler) GetRPCMethod(ctx context.Context, req *RPCReq) (*RPCRes, error) {
	if e.cache == nil {
		return nil, nil
	}
	e.m.RLock()
	defer e.m.RUnlock()

	key := e.key(req)
	val, err := e.cache.Get(ctx, key)
	if err != nil {
		log.Error("error reading from cache", "key", key, "method", req.Method, "err", err)
		return nil, err
	}
	if val == "" {
		return nil, nil
	}

	var result interface{}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		log.Error("error unmarshalling value from cache", "key", key, "method", req.Method, "err", err)
		return nil, err
	}
	return &RPCRes{
		JSONRPC: req.JSONRPC,
		Result:  result,
		ID:      req.ID,
	}, nil
}

func (e *StaticMethodHandler) PutRPCMethod(ctx context.Context, req *RPCReq, res *RPCRes) error {
	if e.cache == nil {
		return nil
	}

	e.m.Lock()
	defer e.m.Unlock()

	key := e.key(req)
	value := mustMarshalJSON(res.Result)

	err := e.cache.Put(ctx, key, string(value))
	if err != nil {
		log.Error("error marshalling value to put into cache", "key", key, "method", req.Method, "err", err)
		return err
	}
	return nil
}
