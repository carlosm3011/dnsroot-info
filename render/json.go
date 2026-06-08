package render

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"rootinfo/query"
)

type jsonServer struct {
	Letter     string `json:"letter"`
	IPv4       string `json:"ipv4"`
	IPv4Result string `json:"ipv4_result,omitempty"`
	IPv4RTTms  int64  `json:"ipv4_rtt_ms,omitempty"`
	IPv4Error  string `json:"ipv4_error,omitempty"`
	IPv6       string `json:"ipv6"`
	IPv6Result string `json:"ipv6_result,omitempty"`
	IPv6RTTms  int64  `json:"ipv6_rtt_ms,omitempty"`
	IPv6Error  string `json:"ipv6_error,omitempty"`
}

type jsonMeta struct {
	Author    string `json:"author,omitempty"`
	Version   string `json:"version,omitempty"`
	BuildDate string `json:"build_date,omitempty"`
	Arch      string `json:"arch,omitempty"`
}

type jsonOutput struct {
	Timestamp time.Time    `json:"timestamp"`
	Refresh   int          `json:"refresh"`
	Meta      jsonMeta     `json:"meta,omitempty"`
	Servers   []jsonServer `json:"servers"`
}

// JSON writes one newline-delimited JSON object for the given results to w.
func JSON(w io.Writer, results []query.Result, refresh int, meta Meta) {
	out := jsonOutput{
		Timestamp: time.Now().UTC(),
		Refresh:   refresh,
		Meta:      jsonMeta{Author: meta.Author, Version: meta.Version, BuildDate: meta.BuildDate, Arch: meta.Arch},
		Servers:   make([]jsonServer, len(results)),
	}
	for i, r := range results {
		js := jsonServer{
			Letter: r.Server.Letter,
			IPv4:   r.Server.IPv4,
			IPv6:   r.Server.IPv6,
		}
		if r.IPv4Err != nil {
			js.IPv4Error = r.IPv4Err.Error()
		} else {
			js.IPv4Result = r.IPv4Result
			js.IPv4RTTms = r.IPv4RTT.Milliseconds()
		}
		if r.IPv6Err != nil {
			js.IPv6Error = r.IPv6Err.Error()
		} else {
			js.IPv6Result = r.IPv6Result
			js.IPv6RTTms = r.IPv6RTT.Milliseconds()
		}
		out.Servers[i] = js
	}
	data, _ := json.Marshal(out)
	fmt.Fprintf(w, "%s\n", data)
}
