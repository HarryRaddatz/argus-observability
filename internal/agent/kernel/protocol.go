package kernel

// portKinds labels an edge by the listening port, which is the only protocol
// hint available without parsing payloads.
var portKinds = map[uint16]string{
	80:    "http",
	443:   "http",
	3000:  "http",
	5432:  "postgres",
	3306:  "mysql",
	6379:  "redis",
	5672:  "amqp",
	15672: "amqp",
	27017: "mongodb",
	9092:  "kafka",
	11211: "memcached",
	9200:  "elasticsearch",
	53:    "dns",
	25:    "smtp",
	1883:  "mqtt",
}

// KindForPort maps a listening port to an edge kind, defaulting to tcp.
func KindForPort(port uint16) string {
	if kind, ok := portKinds[port]; ok {
		return kind
	}
	if port >= 8000 && port <= 8999 {
		return "http"
	}
	return "tcp"
}
