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
	IPv4Error  string `json:"ipv4_error,omitempty"`
	IPv6       string `json:"ipv6"`
	IPv6Result string `json:"ipv6_result,omitempty"`
	IPv6Error  string `json:"ipv6_error,omitempty"`
}

type jsonOutput struct {
	Timestamp time.Time    `json:"timestamp"`
	Refresh   int          `json:"refresh"`
	Servers   []jsonServer `json:"servers"`
}

// JSON writes one newline-delimited JSON object for the given results to w.
func JSON(w io.Writer, results []query.Result, refresh int) {
	out := jsonOutput{
		Timestamp: time.Now().UTC(),
		Refresh:   refresh,
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
		}
		if r.IPv6Err != nil {
			js.IPv6Error = r.IPv6Err.Error()
		} else {
			js.IPv6Result = r.IPv6Result
		}
		out.Servers[i] = js
	}
	data, _ := json.Marshal(out)
	fmt.Fprintf(w, "%s\n", data)
}
