package kong

import (
	"io"

	jsoniter "github.com/json-iterator/go"
)

type Log struct {
	Service             Service             `json:"service"`
	Route               Route               `json:"route"`
	Request             Request             `json:"request"`
	Response            Response            `json:"response"`
	Latencies           Latencies           `json:"latencies"`
	Tries               []Try               `json:"tries"`
	ClientIP            string              `json:"client_ip"`
	Workspace           string              `json:"workspace"`
	WorkspaceName       string              `json:"workspace_name"`
	UpstreamURI         string              `json:"upstream_uri"`
	AuthenticatedEntity AuthenticatedEntity `json:"authenticated_entity"`
	Consumer            Consumer            `json:"consumer"`
	StartedAt           int                 `json:"started_at"`
}

type AuthenticatedEntity struct {
	ID string `json:"id"`
}

type Consumer struct {
	ID            string `json:"id"`
	CreatedAt     int    `json:"created_at"`
	UsernameLower string `json:"username_lower"`
	Username      string `json:"username"`
	Type          int    `json:"type"`
}

type Latencies struct {
	Request int `json:"request"`
	Kong    int `json:"kong"`
	Proxy   int `json:"proxy"`
}

type Request struct {
	Querystring Querystring       `json:"querystring"`
	Size        int               `json:"size"`
	URI         string            `json:"uri"`
	URL         string            `json:"url"`
	Headers     map[string]string `json:"headers"`
	Method      string            `json:"method"`
}

type Querystring struct{}

type Response struct {
	Headers ResponseHeaders `json:"headers"`
	Status  int             `json:"status"`
	Size    int             `json:"size"`
}

type ResponseHeaders struct {
	ContentType                   string `json:"content-type"`
	Date                          string `json:"date"`
	Connection                    string `json:"connection"`
	AccessControlAllowCredentials string `json:"access-control-allow-credentials"`
	ContentLength                 string `json:"content-length"`
	Server                        string `json:"server"`
	Via                           string `json:"via"`
	XKongProxyLatency             string `json:"x-kong-proxy-latency"`
	XKongUpstreamLatency          string `json:"x-kong-upstream-latency"`
	AccessControlAllowOrigin      string `json:"access-control-allow-origin"`
}

type Route struct {
	ID                      string              `json:"id"`
	Paths                   []string            `json:"paths"`
	Protocols               []string            `json:"protocols"`
	StripPath               bool                `json:"strip_path"`
	CreatedAt               int                 `json:"created_at"`
	WsID                    string              `json:"ws_id"`
	RequestBuffering        bool                `json:"request_buffering"`
	UpdatedAt               int                 `json:"updated_at"`
	PreserveHost            bool                `json:"preserve_host"`
	RegexPriority           int                 `json:"regex_priority"`
	ResponseBuffering       bool                `json:"response_buffering"`
	HTTPSRedirectStatusCode int                 `json:"https_redirect_status_code"`
	PathHandling            string              `json:"path_handling"`
	Service                 AuthenticatedEntity `json:"service"`
}

type Service struct {
	Host           string `json:"host"`
	CreatedAt      int    `json:"created_at"`
	ConnectTimeout int    `json:"connect_timeout"`
	ID             string `json:"id"`
	Protocol       string `json:"protocol"`
	ReadTimeout    int    `json:"read_timeout"`
	Port           int    `json:"port"`
	Path           string `json:"path"`
	UpdatedAt      int    `json:"updated_at"`
	WriteTimeout   int    `json:"write_timeout"`
	Retries        int    `json:"retries"`
	WsID           string `json:"ws_id"`
}

type Try struct {
	BalancerLatency int    `json:"balancer_latency"`
	Port            int    `json:"port"`
	BalancerStart   int    `json:"balancer_start"`
	IP              string `json:"ip"`
}

func ParseLog(body io.Reader) (*Log, error) {
	var log Log

	err := jsoniter.NewDecoder(body).Decode(&log)
	if err != nil {
		return nil, err
	}

	return &log, nil
}
