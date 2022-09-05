package controllers

import (
	"context"
	"encoding/json"
	"strings"

	scalablev1alpha1 "github.com/benm-stm/solace-scalable-k8s-operator/api/v1alpha1"
)

// message vpns response struct
type SolaceMsgVpnResp struct {
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
type SolaceMsgVpnsResp struct {
	Data []SolaceMsgVpnResp `json:"data"`
}

//clientUsernames response struct
type ClientUsernameAttribute struct {
	AttributeName  string `json:"attributeName"`
	AttributeValue string `json:"attributeValue"`
	ClientUsername string `json:"clientUsername"`
	MsgVpnName     string `json:"msgVpnName"`
}
type ClientUsernameAttributes struct {
	Data []ClientUsernameAttribute `json:"data"`
}

type SolaceClientUsernameResp struct {
	ClientUsername string                    `json:"clientUsername"`
	Enabled        bool                      `json:"enabled"`
	MsgVpnName     string                    `json:"msgVpnName"`
	Attributes     []ClientUsernameAttribute `json:"attributes"`
	Ports          []int32                   `json:"ports"`
}
type SolaceClientUsernamesResp struct {
	Data []SolaceClientUsernameResp `json:"data"`
}
type SolaceSvcSpec struct {
	MsgVpnName     string  `json:"msgVpnName"`
	ClientUsername string  `json:"clientUsername"`
	Ppp            []Ppp   `json:"ppp"`
	AllMsgVpnPorts []int32 `json:"AllMsgVpnPorts"`
}
type Ppp struct {
	Protocol string  `json:"protocol"`
	Port     []int32 `json:"port"`
	PubOrSub string  `json:"pubOrSub"`
}

type Attributes struct {
	AttributeName  string `json:"attributeName"`
	AttributeValue string `json:"attributeValue"`
}

type protocols struct {
	ServiceAmqpPlainTextListenPort         string `json:"amqpPlainText"`
	ServiceAmqpTlsListenPort               string `json:"amqpTls"`
	ServiceMqttPlainTextListenPort         string `json:"mqttPlainText"`
	ServiceMqttTlsListenPort               string `json:"mqttTls"`
	ServiceMqttTlsWebSocketListenPort      string `json:"mqttWebSocket"`
	ServiceRestIncomingPlainTextListenPort string `json:"restIncomingPlainText"`
	ServiceRestIncomingTlsListenPort       string `json:"restIncomingTls"`
}

func protocolsList() protocols {
	return protocols{
		ServiceAmqpPlainTextListenPort:         "amqp",
		ServiceAmqpTlsListenPort:               "amqps",
		ServiceMqttPlainTextListenPort:         "mqtt",
		ServiceMqttTlsListenPort:               "mqtts",
		ServiceMqttTlsWebSocketListenPort:      "mqttws",
		ServiceRestIncomingPlainTextListenPort: "rest",
		ServiceRestIncomingTlsListenPort:       "rests",
	}
}

func GetEnabledSolaceMsgVpns(
	s *scalablev1alpha1.SolaceScalable,
	data string,
) (SolaceMsgVpnsResp, error) {
	textBytes := []byte(data)

	resp := SolaceMsgVpnsResp{}
	if err := json.Unmarshal(textBytes, &resp); err != nil {
		return SolaceMsgVpnsResp{}, err
	}
	//fmt.Printf("GetEnabledSolaceMsgVpns : %v\n", resp)
	return resp, nil
}

func (m *SolaceMsgVpnsResp) GetSolaceClientUsernames(
	s *scalablev1alpha1.SolaceScalable,
	ctx context.Context,
) (SolaceClientUsernamesResp, error) {
	temp := SolaceClientUsernamesResp{}
	resp := SolaceClientUsernamesResp{}
	for _, m := range m.Data {
		//ignore #client-username
		text, _, err := CallSolaceSempApi(
			s, "/config/msgVpns/"+m.MsgVpnName+
				"/clientUsernames?select="+
				"clientUsername,enabled,msgVpnName"+
				"&where=clientUsername!=*client-username",
			ctx,
			solaceAdminPassword,
		)
		if err != nil {
			return SolaceClientUsernamesResp{}, err
		}
		textBytes := []byte(text)
		err = json.Unmarshal(textBytes, &temp)
		if err != nil {
			return SolaceClientUsernamesResp{}, err
		}

		resp.Data = append(resp.Data, temp.Data...)
	}
	//fmt.Printf("GetSolaceClientUsernames : %v\n", resp)
	return resp, nil
}

func GetClientUsernameAttributes(
	s *scalablev1alpha1.SolaceScalable,
	data string,
) (ClientUsernameAttributes, error) {
	textBytes := []byte(data)
	resp := ClientUsernameAttributes{}
	if err := json.Unmarshal(textBytes, &resp); err != nil {
		return ClientUsernameAttributes{}, err
	}
	//fmt.Printf("GetClientUsernameAttributes : %v\n", resp)
	return resp, nil
}

func (c *SolaceClientUsernamesResp) AddClientAttributes(a ClientUsernameAttributes) []SolaceSvcSpec {
	svcSpecs := []SolaceSvcSpec{}
	for _, c := range c.Data {
		svcSpec := SolaceSvcSpec{}
		// add client username and msgvpn
		svcSpec.MsgVpnName = c.MsgVpnName
		svcSpec.ClientUsername = c.ClientUsername
		for _, attr := range a.Data {
			// Add attributes if they exist
			if attr.MsgVpnName == c.MsgVpnName && attr.ClientUsername == c.ClientUsername {
				if attr.AttributeName == "pub" || attr.AttributeName == "sub" {
					for _, protocol := range strings.Fields(attr.AttributeValue) {
						svcSpec.Ppp = append(svcSpec.Ppp, Ppp{
							Protocol: protocol,
							PubOrSub: attr.AttributeName,
						})
					}
				}
			}
		}
		svcSpecs = append(svcSpecs, svcSpec)
	}
	return svcSpecs
}

func (s *SolaceSvcSpec) AddMsgVpnPorts(m SolaceMsgVpnResp) {
	protocolsExist := false

	if s.MsgVpnName == m.MsgVpnName {
		protocols := protocolsList()
		for k, v := range s.Ppp {
			protocolsExist = true
			s.Ppp[k].Port = append(s.Ppp[k].Port, int32(GetMsgVpnProtocolPort(m, v.Protocol, protocols)))
			//fmt.Printf("vpn :%v\nprotocol :%v\nnature :%v\n\n", m.MsgVpnName, v.Protocol, v.PubOrSub)
		}

		if !protocolsExist {
			// App ports
			s.AllMsgVpnPorts = append(s.AllMsgVpnPorts, int32(m.ServiceAmqpPlainTextListenPort))
			s.AllMsgVpnPorts = append(s.AllMsgVpnPorts, int32(m.ServiceAmqpTlsListenPort))
			s.AllMsgVpnPorts = append(s.AllMsgVpnPorts, int32(m.ServiceMqttPlainTextListenPort))
			s.AllMsgVpnPorts = append(s.AllMsgVpnPorts, int32(m.ServiceMqttTlsListenPort))
			s.AllMsgVpnPorts = append(s.AllMsgVpnPorts, int32(m.ServiceMqttTlsWebSocketListenPort))
			s.AllMsgVpnPorts = append(s.AllMsgVpnPorts, int32(m.ServiceMqttWebSocketListenPort))
			s.AllMsgVpnPorts = append(s.AllMsgVpnPorts, int32(m.ServiceRestIncomingPlainTextListenPort))
			s.AllMsgVpnPorts = append(s.AllMsgVpnPorts, int32(m.ServiceRestIncomingTlsListenPort))
		}
	}
}

func GetMsgVpnProtocolPort(m SolaceMsgVpnResp, s string, p protocols) int {
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

//GetEnabledSolaceMsgVpns : {[{default 5672 5671 1883 8883 8443 8000 9000 9443} {test 0 0 1884 1885 0 1886 0 0}]}
//GetSolaceClientUsernames : {[{default true default   []} {botti true test   []} {default true test   []}]}
//MergeSolaceResponses : {[{default default [5672 5671 1883 8883 8443 8000 9000 9443]} {test botti [0 0 1884 1885 0 1886 0 0]} {test default [0 0 1884 1885 0 1886 0 0]}]}
