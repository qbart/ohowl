package cloudh

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-acme/lego/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/hetzner"
	"github.com/go-acme/lego/v4/registration"
	"github.com/qbart/ohowl/tea"
	"golang.org/x/net/idna"
)

func AutoTls(dns Dns) error {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	user := AcmeUser{
		Email: dns.Email,
		key:   privateKey,
	}

	config := lego.NewConfig(&user)
	config.CADirURL = lego.LEDirectoryProduction
	if dns.Debug {
		log.Println("!!! STAGING MODE !!!")
		config.CADirURL = lego.LEDirectoryStaging
	}
	client, err := lego.NewClient(config)
	if err != nil {
		return err
	}

	hc := hetzner.NewDefaultConfig()
	hc.APIKey = dns.Token
	provider, err := hetzner.NewDNSProviderConfig(hc)
	if err != nil {
		return err
	}
	client.Challenge.SetDNS01Provider(provider)

	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return err
	}
	user.Registration = reg

	request := certificate.ObtainRequest{
		Domains: dns.Domains,
		Bundle:  true,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		return err
	}

	fs := TlsFS{Path: dns.Path}
	return fs.Write(certificates)
}

func ListTls(dns Dns) ([]TlsCert, error) {
	fs := TlsFS{Path: dns.Path}
	return fs.List()
}

type TlsFS struct {
	Path string
}

func (fs *TlsFS) Write(res *certificate.Resource) error {
	return tea.ErrCoalesce(
		ioutil.WriteFile(filepath.Join(fs.Path, fs.escapeFileName(fmt.Sprint(res.Domain, ".key"))), res.PrivateKey, 0o644),
		ioutil.WriteFile(filepath.Join(fs.Path, fs.escapeFileName(fmt.Sprint(res.Domain, ".crt"))), res.Certificate, 0o644),
		ioutil.WriteFile(filepath.Join(fs.Path, fs.escapeFileName(fmt.Sprint(res.Domain, ".ca"))), res.IssuerCertificate, 0o644),
	)
}

func (fs *TlsFS) List() ([]TlsCert, error) {
	matches, err := filepath.Glob(filepath.Join(fs.Path, "*.crt"))
	if err != nil {
		return nil, err
	}

	certs := make([]TlsCert, 0)
	for _, filename := range matches {
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return certs, err
		}
		cert, err := certcrypto.ParsePEMCertificate(data)
		if err != nil {
			return certs, err
		}
		certs = append(certs, TlsCert{
			CommonName: cert.Subject.CommonName,
			DNS:        cert.DNSNames,
			Expiry:     cert.NotAfter,
			Path:       filename,
		})
	}
	return certs, nil
}

func (fs *TlsFS) escapeFileName(f string) string {
	safe, err := idna.ToASCII(strings.Replace(f, "*", "_", -1))
	if err != nil {
		log.Fatal(err)
	}
	return safe
}

type TlsCert struct {
	CommonName string
	DNS        []string
	Expiry     time.Time
	Path       string
}

type TlsStorage interface {
	Write(res *certificate.Resource) error
	List() []TlsCert
}

type Dns struct {
	Token   string
	Email   string
	Path    string
	Domains []string
	Debug   bool
}

type AcmeUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *AcmeUser) GetEmail() string {
	return u.Email
}
func (u AcmeUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *AcmeUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}
