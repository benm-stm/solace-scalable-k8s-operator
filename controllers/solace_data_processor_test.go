package controllers

import (
	"testing"
)

func TestGetEnabledSolaceMsgVpns(t *testing.T) {
	var vpn = `{
		"data":[
			{
				"enabled":true,
				"msgVpnName":"test",
				"serviceAmqpPlainTextListenPort":1100,
				"serviceAmqpTlsListenPort":0,
				"serviceMqttPlainTextListenPort":1050,
				"serviceMqttTlsListenPort":0,
				"serviceMqttTlsWebSocketListenPort":0,
				"serviceMqttWebSocketListenPort":0,
				"serviceRestIncomingPlainTextListenPort":0,
				"serviceRestIncomingTlsListenPort":0
			}
		]
	}`
	got, err := GetEnabledSolaceMsgVpns(
		&solaceScalable,
		vpn,
	)

	if len(got.Data) == 0 && err != nil {
		t.Errorf("got %v, wanted %v", got, nil)
	}
}

func TestGetSolaceClientUsernames(t *testing.T) {
	msgVpns := SolaceMsgVpnsResp{
		Data: []SolaceMsgVpnResp{
			{
				MsgVpnName:                             "testMsgVpn",
				ServiceAmqpPlainTextListenPort:         0,
				ServiceAmqpTlsListenPort:               0,
				ServiceMqttPlainTextListenPort:         0,
				ServiceMqttTlsListenPort:               0,
				ServiceMqttTlsWebSocketListenPort:      0,
				ServiceMqttWebSocketListenPort:         0,
				ServiceRestIncomingPlainTextListenPort: 0,
				ServiceRestIncomingTlsListenPort:       0,
			},
		},
	}
	var data = `{
		"data":[
			{
				"clientUsername":"default",
				"enabled":true,
				"msgVpnName":"default"
			}
		]
	}`
	got, err := msgVpns.GetSolaceClientUsernames(
		&solaceScalable,
		data,
	)

	if len(got.Data) == 0 && err != nil {
		t.Errorf("got %v, wanted %v", got, nil)
	}
}
func TestGetClientUsernameAttributes(t *testing.T) {
	var data = `{
		"data":[
			{
				"attributeName":"pub",
				"attributeValue":"mqtt amqp",
				"clientUsername":"botti",
				"msgVpnName":"test"
			},
			{
				"attributeName":"sub",
				"attributeValue":"amqp",
				"clientUsername":"botti",
				"msgVpnName":"test"
			}
		]
	}`
	got, err := GetClientUsernameAttributes(
		&solaceScalable,
		data,
	)

	if len(got.Data) == 0 && err != nil {
		t.Errorf("got %v, wanted %v", got, nil)
	}
}

func TestAddClientAttributes(t *testing.T) {
	attr := ClientUsernameAttributes{
		Data: []ClientUsernameAttribute{
			{
				AttributeName:  "pub",
				AttributeValue: "mqtt",
				ClientUsername: "testUsername",
				MsgVpnName:     "testMsgVpn",
			},
		},
	}
	resp := SolaceClientUsernamesResp{
		Data: []SolaceClientUsernameResp{
			{
				ClientUsername: "testUsername",
				Enabled:        false,
				MsgVpnName:     "testMsgVpn",
				Attributes:     []ClientUsernameAttribute{},
				Ports:          []int32{},
			},
		},
	}
	got := resp.AddClientAttributes(attr)
	if got[0].Ppp[0].PubOrSub != "pub" {
		t.Errorf("got %v, wanted %v", got[0].Ppp[0].PubOrSub, "pub")
	}
}

func TestAddMsgVpnPorts(t *testing.T) {
	vpn := SolaceMsgVpnResp{
		MsgVpnName:                             "testMsgVpn",
		ServiceAmqpPlainTextListenPort:         1886,
		ServiceAmqpTlsListenPort:               0,
		ServiceMqttPlainTextListenPort:         0,
		ServiceMqttTlsListenPort:               0,
		ServiceMqttTlsWebSocketListenPort:      0,
		ServiceMqttWebSocketListenPort:         0,
		ServiceRestIncomingPlainTextListenPort: 0,
		ServiceRestIncomingTlsListenPort:       0,
	}

	spec := SolaceSvcSpec{
		MsgVpnName:     "",
		ClientUsername: "",
		Ppp: []Ppp{
			{
				Protocol: "mqtt",
				Port: []int32{
					1883,
				},
				PubOrSub: "pub",
			},
		},
		AllMsgVpnPorts: []int32{},
	}

	spec.AddMsgVpnPorts(vpn)
	if len(spec.Ppp[0].Port) != 1 {
		t.Errorf("got %v, wanted %v", len(spec.Ppp[0].Port), 2)
	}
}

func TestGetMsgVpnProtocolPort(t *testing.T) {
	vpn := SolaceMsgVpnResp{
		MsgVpnName:                             "testMsgVpn",
		ServiceAmqpPlainTextListenPort:         1886,
		ServiceAmqpTlsListenPort:               0,
		ServiceMqttPlainTextListenPort:         0,
		ServiceMqttTlsListenPort:               0,
		ServiceMqttTlsWebSocketListenPort:      0,
		ServiceMqttWebSocketListenPort:         0,
		ServiceRestIncomingPlainTextListenPort: 0,
		ServiceRestIncomingTlsListenPort:       0,
	}

	port := GetMsgVpnProtocolPort(
		vpn,
		"amqp",
		protocolsList(),
	)
	if port != 1886 {
		t.Errorf("got %v, wanted %v", port, vpn.ServiceAmqpPlainTextListenPort)
	}
}
