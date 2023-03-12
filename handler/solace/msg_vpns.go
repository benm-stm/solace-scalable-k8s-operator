package solace

import (
	"context"
	"encoding/json"

	scalablev1alpha1 "github.com/benm-stm/solace-scalable-k8s-operator/api/v1alpha1"
	handler "github.com/benm-stm/solace-scalable-k8s-operator/handler"
)

// Message vpn response struct
type SolaceMsgVpn struct {
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
type SolaceMsgVpns struct {
	Data []SolaceMsgVpn `json:"data"`
}

// protocols struct used for protocol's mapping
type protocols struct {
	ServiceAmqpPlainTextListenPort         string `json:"amqpPlainText"`
	ServiceAmqpTlsListenPort               string `json:"amqpTls"`
	ServiceMqttPlainTextListenPort         string `json:"mqttPlainText"`
	ServiceMqttTlsListenPort               string `json:"mqttTls"`
	ServiceMqttTlsWebSocketListenPort      string `json:"mqttWebSocket"`
	ServiceRestIncomingPlainTextListenPort string `json:"restIncomingPlainText"`
	ServiceRestIncomingTlsListenPort       string `json:"restIncomingTls"`
}

const msgVpnspath = "/config/msgVpns"

var options = map[string]string{
	"select": "msgVpnName,enabled,*Port",
	"where":  "enabled==true",
}

// Returns the solace's enabled msgVpns in Json format
func (a *SolaceMsgVpns) GetSolaceEnabledMsgVpns(
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

/*
func GetEnabledSolaceMsgVpns(
	s *scalablev1alpha1.SolaceScalable,
	data string,
) (SolaceMsgVpnsResp, error) {
	textBytes := []byte(data)

	resp := SolaceMsgVpnsResp{}
	if err := json.Unmarshal(textBytes, &resp); err != nil {
		return SolaceMsgVpnsResp{}, err
	}
	return resp, nil
}
*/

// Protocls mapping
func GetMsgVpnProtocolPort(m SolaceMsgVpn, s string, p protocols) int {
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
