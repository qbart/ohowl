package cloudh

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/caddyserver/certmagic"
	"github.com/imroc/req"
	"github.com/libdns/libdns"
	"github.com/qbart/ohowl/tea"
)

func ConfigAutoTls(dns Dns, debug bool) *AutoTlsProvider {
	provider := AutoTlsProvider{
		DNS: dns,
	}

	certmagic.DefaultACME.DNS01Solver = &certmagic.DNS01Solver{
		DNSProvider: &provider,
	}

	if dns.Path != "" {
		certmagic.Default.Storage = &certmagic.FileStorage{Path: dns.Path}
	}

	certmagic.Default.OnDemand = nil
	certmagic.DefaultACME.Agreed = true
	certmagic.DefaultACME.Email = dns.Email

	if debug {
		certmagic.DefaultACME.CA = certmagic.LetsEncryptStagingCA
	}

	// magic := certmagic.NewDefault()

	return &provider
}

type AutoTlsProvider struct {
	DNS Dns

	libdns.RecordGetter
	libdns.RecordAppender
	// libdns.RecordSetter
	libdns.RecordDeleter
}

type Dns struct {
	Token string
	Email string
	Path  string
}

type schemaZone struct {
	Zones []dnsZone `json:"zones,omitempty"`
}

type schemaRecords struct {
	Records []dnsRecord `json:"records,omitempty"`
}

type schemaBulkCreateRecords struct {
	Records        []dnsRecord `json:"records,omitempty"`
	ValidRecords   []dnsRecord `json:"valid_records,omitempty"`
	InvalidRecords []dnsRecord `json:"invalid_records,omitempty"`
}

type dnsRecord struct {
	ID     string `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
	Type   string `json:"type,omitempty"`
	Value  string `json:"value,omitempty"`
	TTL    int    `json:"ttl,omitempty"`
	ZoneID string `json:"zone_id,omitempty"`
}

func (d dnsRecord) String() string {
	return fmt.Sprint(
		d.Name,
		" ",
		d.TTL,
		" IN ",
		d.Type,
		" ",
		d.Value,
	)
}

func (d *dnsRecord) ToLibdns() libdns.Record {
	return libdns.Record{
		ID:    d.ID,
		Name:  d.Name,
		Type:  d.Type,
		Value: d.Value,
		TTL:   time.Duration(d.TTL) * time.Second,
	}
}

func (d *dnsRecord) FromLibdns(rec libdns.Record, zone string) *dnsRecord {
	d.ID = rec.ID
	// d.Name = strings.TrimSuffix(rec.Name, fmt.Sprint(".", ZoneName(zone)))
	d.Name = rec.Name
	d.Type = rec.Type
	d.Value = rec.Value
	d.TTL = int(rec.TTL.Seconds())

	return d
}

type dnsZone struct {
	ID           string   `json:"id,omitempty"`
	Name         string   `json:"name,omitempty"`
	Ns           []string `json:"ns,omitempty"`
	TTL          int      `json:"ttl,omitempty"`
	RecordsCount int      `json:"records_count,omitempty"`
}

func ZoneName(s string) string {
	return strings.TrimSuffix(s, ".")
}

func (atp *AutoTlsProvider) Start(domains []string) error {
	return certmagic.ManageSync(domains)
	// tls, err := certmagic.TLS(domains)
	// log.Println(len(tls.Certificates))
	// return err
}

func (atp *AutoTlsProvider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	log.Println("TLS getting records")
	var (
		schemaZone schemaZone
		schema     schemaRecords
	)

	headers := req.Header{
		"Content-Type":   "application/json",
		"Auth-API-Token": atp.DNS.Token,
	}

	r := tea.HttpGet("https://dns.hetzner.com/api/v1/zones", req.Param{"name": ZoneName(zone)}, headers).ToJSON(&schemaZone)
	if r.Err != nil {
		return nil, r.Err
	}

	r = tea.HttpGet(
		"https://dns.hetzner.com/api/v1/records",
		headers,
		req.Param{"zone_id": schemaZone.Zones[0].ID},
	).ToJSON(&schema)
	if r.Err != nil {
		return nil, r.Err
	}

	recs := make([]libdns.Record, 0)
	for _, rec := range schema.Records {
		log.Println(rec)
		recs = append(recs, rec.ToLibdns())
	}

	return recs, nil
}

func (atp *AutoTlsProvider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	log.Println("TLS appending records")
	headers := req.Header{
		"Content-Type":   "application/json",
		"Auth-API-Token": atp.DNS.Token,
	}

	var schemaZone schemaZone
	r := tea.HttpGet("https://dns.hetzner.com/api/v1/zones", req.Param{"name": ZoneName(zone)}, headers).ToJSON(&schemaZone)
	if r.Err != nil {
		return nil, r.Err
	}

	input := []dnsRecord{}
	for _, rec := range records {
		dnsRec := dnsRecord{ZoneID: schemaZone.Zones[0].ID}
		dnsRec.FromLibdns(rec, zone)
		log.Println(dnsRec)

		input = append(input, dnsRec)
	}

	var schema schemaBulkCreateRecords
	r = tea.HttpPost(
		"https://dns.hetzner.com/api/v1/records/bulk",
		req.Header{
			"Content-Type":   "application/json",
			"Auth-API-Token": atp.DNS.Token,
		},
		req.BodyJSON(schemaRecords{Records: input}),
	).ToJSON(&schema)
	if r.Err != nil {
		return nil, r.Err
	}

	var created []libdns.Record
	for _, r := range schema.ValidRecords {
		created = append(created, r.ToLibdns())
	}

	return created, nil
}

// DeleteRecords deletes the records from the zone. If a record does not have an ID,
// it will be looked up. It returns the records that were deleted.
func (atp *AutoTlsProvider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	log.Println("TLS deleting records")
	existingRecs, err := atp.GetRecords(ctx, zone)
	if err != nil {
		return nil, err
	}
	var todo []libdns.Record

	var wg sync.WaitGroup
	wg.Add(len(records))

	ch := make(chan libdns.Record)

	for _, rec := range records {
		if rec.ID != "" {
			todo = append(todo, rec)
		} else {
			for _, m := range existingRecs {
				tmp := dnsRecord{}
				rec = tmp.FromLibdns(rec, zone).ToLibdns()
				if rec.Name == m.Name && rec.Value == m.Value {
					todo = append(todo, rec)
				}
			}
		}
	}

	for _, rec := range todo {
		go func(rec libdns.Record) {
			defer wg.Done()
			r := tea.HttpDelete(
				fmt.Sprint("https://dns.hetzner.com/api/v1/records/", rec.ID),
				req.Header{
					"Content-Type":   "application/json",
					"Auth-API-Token": atp.DNS.Token,
				},
			)
			if r.Err == nil {
				ch <- rec
			} else {
				log.Println(r.Err)
			}
		}(rec)
	}

	var recs []libdns.Record
	go func() {
		for rec := range ch {
			recs = append(recs, rec)
		}
	}()
	wg.Wait()

	return recs, nil
}

func (atp *AutoTlsProvider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	log.Println("TLS upserting records")
	var results []libdns.Record
	return results, nil
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*AutoTlsProvider)(nil)
	_ libdns.RecordAppender = (*AutoTlsProvider)(nil)
	// _ libdns.RecordSetter   = (*AutoTlsProvider)(nil)
	_ libdns.RecordDeleter = (*AutoTlsProvider)(nil)
)
