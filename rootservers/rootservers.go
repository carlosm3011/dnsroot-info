package rootservers

import "strings"

// Server represents a DNS root server.
type Server struct {
	Letter string
	IPv4   string
	IPv6   string
}

// All contains the 13 root servers in letter order.
var All = []Server{
	{Letter: "A", IPv4: "198.41.0.4", IPv6: "2001:503:ba3e::2:30"},
	{Letter: "B", IPv4: "170.247.170.2", IPv6: "2801:1b8:10::b"},
	{Letter: "C", IPv4: "192.33.4.12", IPv6: "2001:500:2::c"},
	{Letter: "D", IPv4: "199.7.91.13", IPv6: "2001:500:2d::d"},
	{Letter: "E", IPv4: "192.203.230.10", IPv6: "2001:500:a8::e"},
	{Letter: "F", IPv4: "192.5.5.241", IPv6: "2001:500:2f::f"},
	{Letter: "G", IPv4: "192.112.36.4", IPv6: "2001:500:12::d0d"},
	{Letter: "H", IPv4: "198.97.190.53", IPv6: "2001:500:1::53"},
	{Letter: "I", IPv4: "192.36.148.17", IPv6: "2001:7fe::53"},
	{Letter: "J", IPv4: "192.58.128.30", IPv6: "2001:503:c27::2:30"},
	{Letter: "K", IPv4: "193.0.14.129", IPv6: "2001:7fd::1"},
	{Letter: "L", IPv4: "199.7.83.42", IPv6: "2001:500:9f::42"},
	{Letter: "M", IPv4: "202.12.27.33", IPv6: "2001:dc3::35"},
}

// Filter returns only the servers matching the given letters (case-insensitive).
// Returns All when letters is empty.
func Filter(letters []string) []Server {
	if len(letters) == 0 {
		return All
	}
	set := make(map[string]struct{}, len(letters))
	for _, l := range letters {
		set[strings.ToUpper(l)] = struct{}{}
	}
	var result []Server
	for _, s := range All {
		if _, ok := set[s.Letter]; ok {
			result = append(result, s)
		}
	}
	return result
}
