package dns

import (
	"fmt"
	"net"
	"time"

	mdns "github.com/miekg/dns"
)

// Querier sends a CHAOS TXT query to a given server address.
type Querier interface {
	QueryCHAOS(serverAddr string) (instance string, rtt time.Duration, err error)
}

// RealQuerier implements Querier using miekg/dns, querying hostname.bind CH TXT
// directly against the given IP:53.
type RealQuerier struct {
	Timeout time.Duration
}

func (q *RealQuerier) QueryCHAOS(serverAddr string) (string, time.Duration, error) {
	c := &mdns.Client{
		Net:     "udp",
		Timeout: q.Timeout,
	}
	m := new(mdns.Msg)
	m.SetQuestion("hostname.bind.", mdns.TypeTXT)
	m.Question[0].Qclass = mdns.ClassCHAOS
	m.RecursionDesired = false

	resp, rtt, err := c.Exchange(m, net.JoinHostPort(serverAddr, "53"))
	if err != nil {
		return "", rtt, err
	}
	if resp.Rcode != mdns.RcodeSuccess {
		return "", rtt, fmt.Errorf("rcode %s", mdns.RcodeToString[resp.Rcode])
	}
	for _, rr := range resp.Answer {
		if txt, ok := rr.(*mdns.TXT); ok && len(txt.Txt) > 0 {
			return txt.Txt[0], rtt, nil
		}
	}
	return "", rtt, fmt.Errorf("no TXT record in response")
}
