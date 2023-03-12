package solace

import (
	"strings"

	"github.com/rung/go-safecast"
)

// solace svc specification
type SvcSpec struct {
	MsgVpnName     string  `json:"msgVpnName"`
	ClientUsername string  `json:"clientUsername"`
	Pppo           []Pppo  `json:"ppp"`
	AllMsgVpnPorts []int32 `json:"AllMsgVpnPorts"`
}

// solace svc specification
type SvcSpecs struct {
	Data []SvcSpec `json:"data"`
}

// Protocol, Port, PuborSub and OpeningsNumber
type Pppo struct {
	Protocol       string `json:"protocol"`
	Port           int32  `json:"port"`
	PubOrSub       string `json:"pubOrSub"`
	OpeningsNumber int32  `json:"openingsNumber"`
}

// Add message vpn ports in solaceSpec struct
func (s *SvcSpec) MergeMsgVpnPortsInSpec(m *SolaceMsgVpn) {
	protocolsExist := false

	if s.MsgVpnName == m.MsgVpnName {
		for k, v := range s.Pppo {
			protocolsExist = true
			s.Pppo[k].Port = int32(
				GetMsgVpnProtocolPort(
					*m,
					v.Protocol,
					protocolsList(),
				),
			)
		}

		if !protocolsExist {
			// Add all ports
			s.AllMsgVpnPorts = append(s.AllMsgVpnPorts, int32(m.ServiceAmqpPlainTextListenPort))
			s.AllMsgVpnPorts = append(s.AllMsgVpnPorts, int32(m.ServiceAmqpTlsListenPort))
			s.AllMsgVpnPorts = append(s.AllMsgVpnPorts, int32(m.ServiceMqttPlainTextListenPort))
			s.AllMsgVpnPorts = append(s.AllMsgVpnPorts, int32(m.ServiceMqttTlsListenPort))
			s.AllMsgVpnPorts = append(s.AllMsgVpnPorts, int32(m.ServiceMqttTlsWebSocketListenPort))
			s.AllMsgVpnPorts = append(s.AllMsgVpnPorts, int32(m.ServiceMqttWebSocketListenPort))
			s.AllMsgVpnPorts = append(
				s.AllMsgVpnPorts,
				int32(m.ServiceRestIncomingPlainTextListenPort),
			)
			s.AllMsgVpnPorts = append(s.AllMsgVpnPorts, int32(m.ServiceRestIncomingTlsListenPort))
		}
	}
}

// Add client username attributes in solaceSpec struct
func (s *SvcSpecs) MergeClientAttributesInSpec(
	a *ClientUsernameAttributes,
	c *ClientUsernames,
) error {
	for _, c := range c.Data {
		svcSpec := SvcSpec{}
		// add client username and msgvpn
		svcSpec.MsgVpnName = c.MsgVpnName
		svcSpec.ClientUsername = c.ClientUsername
		for _, attr := range a.Data {
			// Add attributes if they exist
			if attr.MsgVpnName == c.MsgVpnName && attr.ClientUsername == c.ClientUsername {
				if attr.AttributeName == "pub" || attr.AttributeName == "sub" {
					for _, protocol := range strings.Fields(attr.AttributeValue) {
						//split protocol and number of openings
						po := strings.Split(protocol, ":")
						// if user didn't provide the openings number 1 is the default value
						var openingsNumber int32 = 1
						var err error
						if len(po) == 2 {
							openingsNumber, err = safecast.Atoi32(po[1])
							if err != nil {
								return err
							}
						}
						svcSpec.Pppo = append(svcSpec.Pppo, Pppo{
							Protocol:       po[0],
							PubOrSub:       attr.AttributeName,
							OpeningsNumber: int32(openingsNumber),
						})
					}
				}
			}
		}
		s.Data = append(s.Data, svcSpec)
	}
	return nil
}

/*
returns the correspondance of protocol given by the clientusername attributes
inside solace
*/
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
