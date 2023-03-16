package solace

import (
	"context"
	"encoding/json"

	scalablev1alpha1 "github.com/benm-stm/solace-scalable-k8s-operator/api/v1alpha1"
	handler "github.com/benm-stm/solace-scalable-k8s-operator/handler"
)

// Message vpn response struct
type msgVpn struct {
	MsgVpnName                             string `json:"msgVpnName"`
	ServiceAmqpPlainTextListenPort         int    `json:"serviceAmqpPlainTextListenPort"`
	ServiceAmqpTlsListenPort               int    `json:"serviceAmqpTlsListenPort"`
	ServiceMqttPlainTextListenPort         int    `json:"serviceMqttPlainTextListenPort"`
	ServiceMqttTlsListenPort               int    `json:"serviceMqttTlsListenPort"`
	ServiceMqttTlsWebSocketListenPort      int    `json:"serviceMqttTlsWebSocketListenPort"`
	ServiceMqttWebSocketListenPort         int    `json:"serviceMqttWebSocketListenPort"`
	ServiceRestIncomingPlainTextListenPort int    `json:"serviceRestIncomingPlainTextListenPort"`
	ServiceRestIncomingTlsListenPort       int    `json:"serviceRestIncomingTlsListenPort"`
}

// Message vpns array response struct
type msgVpns struct {
	Data []msgVpn `json:"data"`
}

// protocols struct used for protocol's mapping
type Protocols struct {
	ServiceAmqpPlainTextListenPort         string `json:"amqpPlainText"`
	ServiceAmqpTlsListenPort               string `json:"amqpTls"`
	ServiceMqttPlainTextListenPort         string `json:"mqttPlainText"`
	ServiceMqttTlsListenPort               string `json:"mqttTls"`
	ServiceMqttWebSocketListenPort         string `json:"mqttWebSocket"`
	ServiceMqttTlsWebSocketListenPort      string `json:"mqttTlsWebSocket"`
	ServiceRestIncomingPlainTextListenPort string `json:"restIncomingPlainText"`
	ServiceRestIncomingTlsListenPort       string `json:"restIncomingTls"`
}

const msgVpnspath = "/config/msgVpns"

var options = map[string]string{
	"select": "msgVpnName,enabled,*Port",
	"where":  "enabled==true",
}

// Returns the solace's enabled msgVpns in Json format
func (a *msgVpns) GetEnabledMsgVpns(
	i int,
	s *scalablev1alpha1.SolaceScalable,
	ctx context.Context,
	p string,
) error {
	url := handler.ConstructSempUrl(*s, i, msgVpnspath, options)
	textBytes, _, _ := handler.CallSolaceSempApi(
		url,
		ctx,
		p,
	)
	if err := json.Unmarshal(textBytes, &a); err != nil {
		return err
	}
	return nil
}

// Protocls mapping
func (m *msgVpn) GetMsgVpnProtocolPort(s string, p Protocols) int {
	// supportes protocols
	if p.ServiceAmqpPlainTextListenPort == s {
		return m.ServiceAmqpPlainTextListenPort

	} else if p.ServiceAmqpTlsListenPort == s {
		return m.ServiceAmqpTlsListenPort

	} else if p.ServiceMqttPlainTextListenPort == s {
		return m.ServiceMqttPlainTextListenPort

	} else if p.ServiceMqttTlsListenPort == s {
		return m.ServiceMqttTlsListenPort

	} else if p.ServiceMqttTlsWebSocketListenPort == s {
		return m.ServiceMqttTlsWebSocketListenPort

	} else if p.ServiceRestIncomingPlainTextListenPort == s {
		return m.ServiceRestIncomingPlainTextListenPort

	} else if p.ServiceRestIncomingTlsListenPort == s {
		return m.ServiceRestIncomingTlsListenPort
	}
	return 0
}

func NewMsgVpns() *msgVpns {
	return &msgVpns{}
}

func (a *msgVpn) GetName() string {
	return a.MsgVpnName
}

func (a *msgVpn) GetMsgVpn() msgVpn {
	return *a
}

func (a *msgVpn) GetAmpqpPort() int {
	return a.ServiceAmqpPlainTextListenPort
}
func (a *msgVpn) GetAmpqpsPort() int {
	return a.ServiceAmqpTlsListenPort
}
func (a *msgVpn) GetMqttPort() int {
	return a.ServiceMqttPlainTextListenPort
}
func (a *msgVpn) GetMqttsPort() int {
	return a.ServiceMqttTlsListenPort
}
func (a *msgVpn) GetMqttWebSocket() int {
	return a.ServiceMqttWebSocketListenPort
}
func (a *msgVpn) GetMqttTlsWebSocket() int {
	return a.ServiceMqttTlsWebSocketListenPort
}
func (a *msgVpn) GetRest() int {
	return a.ServiceRestIncomingPlainTextListenPort
}
func (a *msgVpn) GetRests() int {
	return a.ServiceRestIncomingTlsListenPort
}
