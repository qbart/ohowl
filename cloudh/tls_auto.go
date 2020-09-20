package cloudh

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/hetzner"
	"github.com/go-acme/lego/v4/registration"
	"github.com/qbart/ohowl/owl"
	"github.com/qbart/ohowl/tea"
	"golang.org/x/net/idna"
)

// Issue requests new cert.
func (at *AutoTls) Issue() error {
	user, client, err := at.setup()
	if err != nil {
		return err
	}
	if user.Registration == nil {
		reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err != nil {
			return err
		}
		user.Registration = reg

		if err = at.saveAccount(user); err != nil {
			return err
		}
	}

	request := certificate.ObtainRequest{
		Domains:        at.Config.Domains,
		Bundle:         true,
		MustStaple:     false,
		PreferredChain: "",
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		return err
	}

	return at.saveResource(certificates)
}

func (at *AutoTls) Renew(reuseKey bool) error {
	user, client, err := at.setup()
	if err != nil {
		return err
	}
	if user.Registration == nil {
		return fmt.Errorf("Account is not registered. Issue new certificate.")
	}

	if len(at.Config.Domains) == 0 {
		return errors.New("Domain is not specified")
	}
	domain := at.Config.Domains[0]

	certificates, err := at.readCertificate(domain, ".crt")
	if err != nil {
		return fmt.Errorf("Error while loading the certificate for domain %s\n\t%w", domain, err)
	}

	cert := certificates[0]
	if cert.IsCA {
		return fmt.Errorf("[%s] Certificate bundle starts with a CA certificate", domain)
	}

	if !at.needsRenewal(cert, domain, 30) {
		return nil
	}

	timeLeft := cert.NotAfter.Sub(time.Now().UTC())
	log.Printf("[%s] acme: Trying renewal with %d hours remaining", domain, int(timeLeft.Hours()))

	// technically passed domains might be different than cert domains
	// so it should be merged like lego lib does it but for Owl purpose
	// this behavior is not needed
	// certDomains := certcrypto.ExtractDomains(cert)

	var privateKey crypto.PrivateKey
	if reuseKey {
		keyBytes, err := at.Storage.Read(context.TODO(), at.getFileName(domain, ".key"))
		if err != nil {
			return fmt.Errorf("Error while loading the private key for domain %s\n\t%w", domain, err)
		}

		privateKey, err = certcrypto.ParsePEMPrivateKey(keyBytes)
		if err != nil {
			return err
		}
	} else {
		privateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return err
		}
	}

	request := certificate.ObtainRequest{
		Domains:        at.Config.Domains,
		Bundle:         true,
		PrivateKey:     privateKey,
		MustStaple:     false,
		PreferredChain: "",
	}
	certRes, err := client.Certificate.Obtain(request)
	if err != nil {
		return err
	}

	return at.saveResource(certRes)
}

func (at *AutoTls) List() ([]TlsCert, error) {
	matches, err := at.Storage.Find(context.TODO(), at.Config.Path, ".crt")
	if err != nil {
		return nil, err
	}

	certs := make([]TlsCert, 0)
	for _, filename := range matches {
		data, err := at.Storage.Read(context.TODO(), filename)
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

func (at *AutoTls) saveResource(res *certificate.Resource) error {
	return tea.ErrCoalesce(
		at.Storage.Write(context.TODO(), at.getFileName(res.Domain, ".key"), res.PrivateKey),
		at.Storage.Write(context.TODO(), at.getFileName(res.Domain, ".crt"), res.Certificate),
		at.Storage.Write(context.TODO(), at.getFileName(res.Domain, ".ca"), res.IssuerCertificate),
	)
}

func (at *AutoTls) readCertificate(domain, ext string) ([]*x509.Certificate, error) {
	content, err := at.Storage.Read(context.TODO(), at.getFileName(domain, ext))
	if err != nil {
		return nil, err
	}

	return certcrypto.ParsePEMBundle(content)
}

func (at *AutoTls) getFileName(domain, ext string) string {
	safe, err := idna.ToASCII(strings.Replace(fmt.Sprint(domain, ext), "*", "_", -1))
	if err != nil {
		log.Fatal(err)
	}
	return filepath.Join(at.Config.Path, safe)
}

func (at *AutoTls) setupDnsChallenge(client *lego.Client) error {
	hc := hetzner.NewDefaultConfig()
	hc.APIKey = at.Config.Token
	provider, err := hetzner.NewDNSProviderConfig(hc)
	if err != nil {
		return err
	}

	client.Challenge.SetDNS01Provider(provider)
	return nil
}

func (at *AutoTls) needsRenewal(x509Cert *x509.Certificate, domain string, days int) bool {
	if days >= 0 {
		notAfter := int(time.Until(x509Cert.NotAfter).Hours() / 24.0)
		if notAfter > days {
			log.Printf("[%s] The certificate expires in %d days, the number of days defined to perform the renewal is %d: no renewal.", domain, notAfter, days)
			return false
		}
	}
	return true
}

func (at *AutoTls) newClient(acc registration.User, keyType certcrypto.KeyType) (*lego.Client, error) {
	config := lego.NewConfig(acc)
	config.UserAgent = owl.UserAgent
	config.CADirURL = at.caDirUrl()

	config.Certificate = lego.CertificateConfig{
		KeyType: keyType,
		Timeout: 30 * time.Second,
	}

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("Could not create client: %w", err)
	}

	return client, nil
}

func (at *AutoTls) setup() (*AcmeUser, *lego.Client, error) {
	privateKey, err := at.accountPrivateKey()

	user := &AcmeUser{Email: at.Config.Email, key: privateKey}

	if exists, err := at.AccountStorage.Exists(context.TODO(), at.accountFilePath()); exists {
		if user, err = at.readAccount(privateKey); err != nil {
			return nil, nil, err
		}
	}

	client, err := at.newClient(user, certcrypto.EC384)
	if err != nil {
		return nil, nil, err
	}
	if err = at.setupDnsChallenge(client); err != nil {
		return nil, nil, err
	}

	return user, client, nil
}

func (at *AutoTls) saveAccount(user *AcmeUser) error {
	jsonBytes, err := json.MarshalIndent(user, "", "\t")
	if err != nil {
		return err
	}

	return at.AccountStorage.Write(context.TODO(), at.accountFilePath(), jsonBytes)
}

func (at *AutoTls) accountFilePath() string {
	return filepath.Join(at.Config.Path, at.Config.Email+".json")
}

func (at *AutoTls) accountPrivateKey() (crypto.PrivateKey, error) {
	path := filepath.Join(at.Config.Path, at.Config.Email+".key")

	exists, err := at.AccountStorage.Exists(context.TODO(), path)
	if err != nil {
		return nil, err
	}

	if !exists {
		privateKey, err := certcrypto.GeneratePrivateKey(certcrypto.EC384)
		if err != nil {
			return nil, err
		}

		pemKey := certcrypto.PEMBlock(privateKey)
		b := pem.EncodeToMemory(pemKey)

		err = at.Storage.Write(context.TODO(), path, b)
		if err != nil {
			return nil, err
		}
	}

	b, err := at.Storage.Read(context.TODO(), path)
	if err != nil {
		return nil, fmt.Errorf("Failed to load user private key %w", err)
	}
	keyBlock, _ := pem.Decode(b)

	switch keyBlock.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	case "EC PRIVATE KEY":
		return x509.ParseECPrivateKey(keyBlock.Bytes)
	}

	return nil, errors.New("Unknown private key type")
}

func (at *AutoTls) readAccount(key crypto.PrivateKey) (*AcmeUser, error) {
	b, err := at.AccountStorage.Read(context.TODO(), at.accountFilePath())
	if err != nil {
		return nil, err
	}
	var account AcmeUser
	err = json.Unmarshal(b, &account)
	if err != nil {
		return nil, err
	}
	account.key = key

	if account.Registration == nil || account.Registration.Body.Status == "" {
		reg, err := at.tryRecoverRegistration(key)
		if err != nil {
			return nil, fmt.Errorf("Could not load account Registration is nil: %w", err)
		}

		account.Registration = reg
		err = at.saveAccount(&account)
		if err != nil {
			return nil, fmt.Errorf("Could not save account. Registration is nil: %w", err)
		}
	}

	return &account, nil
}

func (at *AutoTls) tryRecoverRegistration(key crypto.PrivateKey) (*registration.Resource, error) {
	config := lego.NewConfig(&AcmeUser{key: key})
	config.UserAgent = owl.UserAgent
	config.CADirURL = at.caDirUrl()

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, err
	}

	reg, err := client.Registration.ResolveAccountByKey()
	if err != nil {
		return nil, err
	}
	return reg, nil
}

func (at *AutoTls) caDirUrl() string {
	result := lego.LEDirectoryProduction
	if at.Config.Debug {
		log.Println("!!! STAGING MODE !!!")
		result = lego.LEDirectoryStaging
	}
	return result
}
