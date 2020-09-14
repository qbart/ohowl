package cloudh

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/go-acme/lego/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/hetzner"
	"github.com/go-acme/lego/v4/registration"
	"github.com/qbart/ohowl/owl"
	"github.com/qbart/ohowl/tea"
	"golang.org/x/net/idna"
)

func (at *AutoTls) IssueNew() error {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	user := AcmeUser{
		Email: at.Config.Email,
		key:   privateKey,
	}

	config := lego.NewConfig(&user)
	config.UserAgent = owl.UserAgent
	config.CADirURL = lego.LEDirectoryProduction
	if at.Config.Debug {
		log.Println("!!! STAGING MODE !!!")
		config.CADirURL = lego.LEDirectoryStaging
	}

	client, err := lego.NewClient(config)
	if err != nil {
		return err
	}
	if err = at.setupDnsChallenge(client); err != nil {
		return err
	}

	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return err
	}
	user.Registration = reg

	request := certificate.ObtainRequest{
		Domains: at.Config.Domains,
		Bundle:  true,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		return err
	}

	return tea.ErrCoalesce(
		at.Storage.Write(filepath.Join(at.Config.Path, at.escapeFileName(fmt.Sprint(certificates.Domain, ".key"))), certificates.PrivateKey),
		at.Storage.Write(filepath.Join(at.Config.Path, at.escapeFileName(fmt.Sprint(certificates.Domain, ".crt"))), certificates.Certificate),
		at.Storage.Write(filepath.Join(at.Config.Path, at.escapeFileName(fmt.Sprint(certificates.Domain, ".ca"))), certificates.IssuerCertificate),
	)
}

func (at *AutoTls) List() ([]TlsCert, error) {
	matches, err := at.Storage.Find(at.Config.Path, "*.crt")
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

func (at *AutoTls) escapeFileName(f string) string {
	safe, err := idna.ToASCII(strings.Replace(f, "*", "_", -1))
	if err != nil {
		log.Fatal(err)
	}
	return safe
}

// func needsRenewal(x509Cert *x509.Certificate, domain string, days int) (bool, error) {
// 	if x509Cert.IsCA {
// 		return false, fmt.Errorf("[%s] Certificate bundle starts with a CA certificate", domain)
// 	}

// 	if days >= 0 {
// 		notAfter := int(time.Until(x509Cert.NotAfter).Hours() / 24.0)
// 		if notAfter > days {
// 			return false, fmt.Errorf("[%s] The certificate expires in %d days, the number of days defined to perform the renewal is %d: no renewal.", domain, notAfter, days)
// 		}
// 	}
// 	account, client := setup(ctx, NewAccountsStorage(ctx))
// 	setupChallenges(ctx, client)

// 	if account.Registration == nil {
// 		log.Fatalf("Account %s is not registered. Use 'run' to register a new account.\n", account.Email)
// 	}

// 	certsStorage := NewCertificatesStorage(ctx)

// 	bundle := !ctx.Bool("no-bundle")

// 	meta := map[string]string{renewEnvAccountEmail: account.Email}

// 	// CSR
// 	if ctx.GlobalIsSet("csr") {
// 		return renewForCSR(ctx, client, certsStorage, bundle, meta)
// 	}

// 	// Domains
// 	return renewForDomains(ctx, client, certsStorage, bundle, meta)
// 	return true, nil
// }
