package query

import (
	"sync"

	dnsq "rootinfo/dns"
	"rootinfo/rootservers"
)

// Result holds the queried instance names for one root server.
type Result struct {
	Server     rootservers.Server
	IPv4Result string
	IPv4Err    error
	IPv6Result string
	IPv6Err    error
}

// Runner orchestrates parallel CHAOS queries for a set of root servers.
type Runner struct {
	Querier   dnsq.Querier
	Servers   []rootservers.Server
	DNSServer string // if set, all queries are routed here instead of direct to root server IPs
}

// Run queries all servers concurrently and returns results in the same order as Servers.
func (r *Runner) Run() []Result {
	results := make([]Result, len(r.Servers))
	var wg sync.WaitGroup
	for i, srv := range r.Servers {
		wg.Add(1)
		go func(idx int, s rootservers.Server) {
			defer wg.Done()
			results[idx] = r.queryServer(s)
		}(i, srv)
	}
	wg.Wait()
	return results
}

func (r *Runner) queryServer(srv rootservers.Server) Result {
	res := Result{Server: srv}
	v4target, v6target := srv.IPv4, srv.IPv6
	if r.DNSServer != "" {
		v4target = r.DNSServer
		v6target = r.DNSServer
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		res.IPv4Result, res.IPv4Err = r.Querier.QueryCHAOS(v4target)
	}()
	go func() {
		defer wg.Done()
		res.IPv6Result, res.IPv6Err = r.Querier.QueryCHAOS(v6target)
	}()
	wg.Wait()
	return res
}
