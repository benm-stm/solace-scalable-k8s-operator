package spec

import "github.com/benm-stm/solace-scalable-k8s-operator/handler/solace"

type msgVpn interface {
	GetName() string
	GetMsgVpnProtocolPort(s string, p solace.Protocols) int
	GetAmpqpPort() int
	GetAmpqpsPort() int
	GetMqttPort() int
	GetMqttsPort() int
	GetMqttWebSocket() int
	GetMqttTlsWebSocket() int
	GetRest() int
	GetRests() int
}

type Pppo struct {
	Protocol       string `json:"protocol"`
	Port           int32  `json:"port"`
	PubOrSub       string `json:"pubOrSub"`
	OpeningsNumber int32  `json:"openingsNumber"`
}

// solace svc specification
type SvcSpec struct {
	MsgVpnName     string  `json:"msgVpnName"`
	ClientUsername string  `json:"clientUsername"`
	Pppo           []Pppo  `json:"pppo"`
	AllMsgVpnPorts []int32 `json:"AllMsgVpnPorts"`
}

// solace svc specification
type SvcsSpec struct {
	Data []SvcSpec `json:"data"`
}
