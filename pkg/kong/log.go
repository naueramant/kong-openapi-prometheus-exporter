package kong

type Log struct {
	ClientIP  string    `json:"client_ip"`
	Request   Request   `json:"request"`
	Response  Response  `json:"response"`
	Latencies Latencies `json:"latencies"`
	Service   Service   `json:"service"`
}

type Request struct {
	URI         string            `json:"uri"`
	Headers     map[string]string `json:"headers"`
	Method      string            `json:"method"`
	Size        int               `json:"size"`
	URL         string            `json:"url"`
	QueryString map[string]string `json:"querystring"`
}

type Response struct {
	Size   int `json:"size"`
	Status int `json:"status"`
}

type Latencies struct {
	Request int `json:"request"`
	Proxy   int `json:"proxy"`
	Kong    int `json:"kong"`
	Receive int `json:"receive"`
}

type Service struct {
	Host string `json:"host"`
	Name string `json:"name"`
	Port int    `json:"port"`
}
