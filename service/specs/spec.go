package spec

import (
	"strings"

	"github.com/benm-stm/solace-scalable-k8s-operator/handler/solace"
	"github.com/rung/go-safecast"
)

func NewSvcsSpec() *SvcsSpec {
	return &SvcsSpec{}
}

// Add message vpn ports in solaceSpec struct
func (s *SvcSpec) WithMsgVpnPorts(m msgVpn) {
	protocolsExist := false

	if s.MsgVpnName == m.GetName() {
		for k, v := range s.Pppo {
			protocolsExist = true
			s.Pppo[k].Port = int32(
				m.GetMsgVpnProtocolPort(
					v.Protocol,
					getProtocols(),
				),
			)
		}

		if !protocolsExist {
			// Add all ports
			s.AllMsgVpnPorts = append(s.AllMsgVpnPorts, int32(m.GetAmpqpPort()))
			s.AllMsgVpnPorts = append(s.AllMsgVpnPorts, int32(m.GetAmpqpsPort()))
			s.AllMsgVpnPorts = append(s.AllMsgVpnPorts, int32(m.GetMqttPort()))
			s.AllMsgVpnPorts = append(s.AllMsgVpnPorts, int32(m.GetMqttWebSocket()))
			s.AllMsgVpnPorts = append(s.AllMsgVpnPorts, int32(m.GetMqttTlsWebSocket()))
			s.AllMsgVpnPorts = append(s.AllMsgVpnPorts, int32(m.GetMqttsPort()))
			s.AllMsgVpnPorts = append(s.AllMsgVpnPorts, int32(m.GetRest()))
			s.AllMsgVpnPorts = append(s.AllMsgVpnPorts, int32(m.GetRests()))
		}
	}
}

/*
returns the correspondance of protocol given by the clientusername attributes
for solace
*/
func getProtocols() solace.Protocols {
	return solace.Protocols{
		ServiceAmqpPlainTextListenPort:         "amqp",
		ServiceAmqpTlsListenPort:               "amqps",
		ServiceMqttPlainTextListenPort:         "mqtt",
		ServiceMqttTlsListenPort:               "mqtts",
		ServiceMqttTlsWebSocketListenPort:      "mqttws",
		ServiceRestIncomingPlainTextListenPort: "rest",
		ServiceRestIncomingTlsListenPort:       "rests",
	}
}

// Add client username attributes in solaceSpec struct
func (s *SvcsSpec) WithClientAttributes(
	a *solace.ClientUsernameAttributes,
	c *solace.ClientUsernames,
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
