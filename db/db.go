package db

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gomodule/redigo/redis"
)

type DB struct {
	conn redis.Conn
}

type KanameMadoka struct {
	UserId   string
	UserName string
}

var Keywords []string
var KanameMadokas []KanameMadoka

func NewDB(url string, logger *log.Logger) (*DB, error) {
	return &DB{}, nil
}

func (d *DB) AddKeyword(w string) error {
	c, err := redis.DialURL(os.Getenv("REDIS_URL"))
	if err != nil {
		return err
	}
	_, err = c.Do("SET", "keyword:"+w, nil)
	if err != nil {
		return fmt.Errorf("set keyword failed: %w", err)
	}
	c.Close()

	Keywords = append(Keywords, w)

	return nil
}

func (d *DB) AddKanameMadoka(m KanameMadoka) error {
	c, err := redis.DialURL(os.Getenv("REDIS_URL"))
	if err != nil {
		return err
	}
	_, err = c.Do("SET", "kaname-madoka:"+m.UserId, m.UserName)
	if err != nil {
		return fmt.Errorf("set kaname-madoka failed: %w", err)
	}
	c.Close()

	KanameMadokas = append(KanameMadokas, m)

	return nil
}

func (d *DB) ListKeywords() ([]string, error) {
	c, err := redis.DialURL(os.Getenv("REDIS_URL"))
	if err != nil {
		return nil, err
	}
	keys, err := redis.Strings(c.Do("KEYS", "keyword:*"))
	if err != nil {
		return nil, fmt.Errorf("get keys failed: %w", err)
	}
	c.Close()
	var keywords []string
	for _, k := range keys {
		k = strings.TrimPrefix(k, "keyword:")
		keywords = append(keywords, k)
	}
	return keywords, nil
}

func (d *DB) ListKanameMadokas() ([]KanameMadoka, error) {
	c, err := redis.DialURL(os.Getenv("REDIS_URL"))
	if err != nil {
		return nil, err
	}
	keys, err := redis.Strings(c.Do("KEYS", "kaname-madoka:*"))
	if err != nil {
		return nil, fmt.Errorf("get keys failed: %w", err)
	}
	kanameMadokas := make([]KanameMadoka, len(keys))
	if len(kanameMadokas) == 0 {
		return kanameMadokas, nil
	}

	var args []interface{}
	for _, k := range keys {
		args = append(args, k)
	}

	bs, err := redis.ByteSlices(c.Do("MGET", args...))
	if err == redis.ErrNil {
		return make([]KanameMadoka, 0), nil
	}
	if err != nil {
		return nil, fmt.Errorf("mget failed: %w", err)
	}
	c.Close()

	for i, b := range bs {
		kanameMadokas[i] = KanameMadoka{
			UserId:   keys[i],
			UserName: string(b[i]),
		}
	}
	return kanameMadokas, nil
}
