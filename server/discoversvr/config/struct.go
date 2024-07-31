package config

type ServiceInfo struct {
	ServiceName string `json:"serviceName"`
	InstanceId  string `json:"instanceId"`
	Host        string `json:"host"`
	Port        string `json:"port"`
	Protocol    string `json:"protocol"`
	Weight      int64  `json:"weight"`
	Timeout     int64  `json:"timeout"`
	Metadata    string `json:"metadata"`
}

type RouteInfo struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Prefix   string `json:"prefix"`
	Metadata string `json:"metadata"`
}
