package serverpool

type LBStrategy int

const (
	RoundRobin LBStrategy = iota
	LeastConnections
	Random
	IPHash
	URIHash
)

func GetLBStrategy(strategy string) LBStrategy {
	switch strategy {
	case "least_conn":
		return LeastConnections
	case "random":
		return Random
	case "ip_hash":
		return IPHash
	case "uri_hash":
		return URIHash
	default:
		return RoundRobin
	}
}
