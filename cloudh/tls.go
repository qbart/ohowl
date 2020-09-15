package cloudh

import (
	"crypto"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/go-acme/lego/v4/registration"
)

type TlsStorage interface {
	Exists(key string) (bool, error)
	Write(key string, b []byte) error
	Read(key string) ([]byte, error)
	Find(key string, filter string) ([]string, error)
}

type TlsFileStorage struct{}

type TlsConfig struct {
	Token   string
	Email   string
	Path    string
	Domains []string
	Debug   bool
}

type AutoTls struct {
	Config  TlsConfig
	Storage TlsStorage
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

func (fs *TlsFileStorage) Exists(key string) (bool, error) {
	if _, err := os.Stat(key); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (fs *TlsFileStorage) Write(key string, b []byte) error {
	return ioutil.WriteFile(key, b, 0o644)
}

func (fs *TlsFileStorage) Read(key string) ([]byte, error) {
	return ioutil.ReadFile(key)
}

func (fs *TlsFileStorage) Find(key string, filter string) ([]string, error) {
	return filepath.Glob(filepath.Join(key, filter))
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
