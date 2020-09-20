package cloudh

import (
	"context"
	"crypto"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/registration"
	consulapi "github.com/hashicorp/consul/api"
)

type TlsStorage interface {
	Exists(ctx context.Context, key string) (bool, error)
	Write(ctx context.Context, key string, b []byte) error
	Read(ctx context.Context, key string) ([]byte, error)
	Find(ctx context.Context, key string, ext string) ([]string, error)
}

type TlsFileStorage struct{}
type TlsConsulFileStorage struct {
	KV *consulapi.KV
}

type TlsConfig struct {
	Token   string
	Email   string
	Path    string
	Domains []string
	Debug   bool
}

type AutoTls struct {
	Config         TlsConfig
	Storage        TlsStorage
	AccountStorage TlsStorage
}

type TlsCert struct {
	CommonName string
	DNS        []string
	Expiry     time.Time
	Path       string
}

type AcmeUser struct {
	Email        string                 `json:"email,omitempty"`
	Registration *registration.Resource `json:"registration,omitempty"`
	key          crypto.PrivateKey
}

// ----- TlsFileStorage -----

func (fs *TlsFileStorage) Exists(ctx context.Context, key string) (bool, error) {
	if _, err := os.Stat(key); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (fs *TlsFileStorage) Write(ctx context.Context, key string, b []byte) error {
	return ioutil.WriteFile(key, b, 0o644)
}

func (fs *TlsFileStorage) Read(ctx context.Context, key string) ([]byte, error) {
	return ioutil.ReadFile(key)
}

func (fs *TlsFileStorage) Find(ctx context.Context, key string, ext string) ([]string, error) {
	return filepath.Glob(filepath.Join(key, fmt.Sprint("*", ext)))
}

// ----- ConsulFileStorage -----

func (fs *TlsConsulFileStorage) Exists(ctx context.Context, key string) (bool, error) {
	if _, _, err := fs.KV.Get(key, &consulapi.QueryOptions{RequireConsistent: true}); err != nil {
		return false, err
	}
	return true, nil
}

func (fs *TlsConsulFileStorage) Write(ctx context.Context, key string, b []byte) error {
	kv := &consulapi.KVPair{Key: key, Value: b}
	if _, err := fs.KV.Put(kv, nil); err != nil {
		return err
	}
	return nil
}

func (fs *TlsConsulFileStorage) Read(ctx context.Context, key string) ([]byte, error) {
	kv, _, err := fs.KV.Get(key, &consulapi.QueryOptions{RequireConsistent: true})
	if err != nil {
		return nil, err
	}
	return kv.Value, nil
}

func (fs *TlsConsulFileStorage) Find(ctx context.Context, key string, ext string) ([]string, error) {
	kvs, _, err := fs.KV.List(fmt.Sprint(key, "/"), nil)
	if err != nil {
		return nil, err
	}
	res := make([]string, 0)
	for _, kv := range kvs {
		if strings.HasSuffix(kv.Key, ext) {
			res = append(res, kv.Key)
		}
	}

	return res, nil
}

// ----- AcmeUser -----

func (u *AcmeUser) GetEmail() string {
	return u.Email
}
func (u AcmeUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *AcmeUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}
