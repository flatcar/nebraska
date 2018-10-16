package sessions

import (
	"bytes"
	"encoding/gob"
	"log"

	"github.com/fzzy/radix/extra/pool"
	"github.com/fzzy/radix/redis"
)

type RedisStore struct {
	Prefix     string
	Network    string
	Address    string
	ClientPool *pool.Pool
}

func NewRedisStore(network, addr string) *RedisStore {
	p, err := pool.NewPool(network, addr, 1)
	if err != nil {
		log.Fatalln("Redis connection pool initialization failed.", err)
	}
	return &RedisStore{
		Prefix:     "session",
		Network:    network,
		Address:    addr,
		ClientPool: p,
	}
}

func (r RedisStore) getClient() *redis.Client {
	client, err := r.ClientPool.Get()
	if err != nil {
		log.Panicf("Redis Get Client error: %v\n", err)
	}
	return client
}

func (r RedisStore) Get(key string) (map[string]interface{}, error) {
	client := r.getClient()
	defer client.Close()
	// GET from redis.
	res := client.Cmd("GET", r.Prefix+key)
	if res.Err != nil {
		log.Panicf("Redis GET error:", res.Err)
	}
	// Check if key is empty
	if res.Type == redis.NilReply {
		return nil, NotFoundError
	}
	// Get bytes from redis response.
	b, err := res.Bytes()
	if err != nil {
		log.Panicf("Redis GET error:", err)
	}
	// Decode bytes.
	var obj map[string]interface{}
	buffer := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buffer)
	err = dec.Decode(&obj)
	if err != nil {
		log.Panicf("Decoding error:", err)
	}
	return obj, nil
}

func (r RedisStore) Save(key string, object map[string]interface{}) {
	client := r.getClient()
	defer client.Close()
	b := bytes.Buffer{}
	enc := gob.NewEncoder(&b)
	err := enc.Encode(object)
	if err != nil {
		log.Panicln("Encoding Error:", err)
	}
	res := client.Cmd("SET", r.Prefix+key, b.Bytes())
	if res.Err != nil {
		log.Panicln("Redis SET error:", res.Err)
	}

}

func (r RedisStore) Destroy(key string) {
	client := r.getClient()
	defer client.Close()
	err := client.Cmd("DEL", r.Prefix+key).Err
	if err != nil {
		panic(err)
	}
}
