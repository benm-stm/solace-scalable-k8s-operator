package controllers

/*
func TestGetSolaceOpenPorts(t *testing.T) {
	got, err := GetSolaceOpenPorts(
		&solaceScalable,
		context.TODO(),
	)

	if got == nil && err != nil {
		t.Errorf("got %v, wanted %v", got, nil)
	}
}
*/
/*
func TestGetEnabledSolaceMsgVpns(t *testing.T) {
	got, err := GetEnabledSolaceMsgVpns(
		&solaceScalable,
		context.TODO(),
	)

	if got[0] == nil && err != nil {
		t.Errorf("got %v, wanted %v", got, nil)
	}
}
*/
/*
func TestMergeSolaceResponses(t *testing.T) {
	var vpn = SolaceMsgVpnsResp{
		Data: []SolaceMsgVpnResp{
			{
				MsgVpnName:                             "test1",
				ServiceAmqpPlainTextListenPort:         1,
				ServiceAmqpTlsListenPort:               2,
				ServiceMqttPlainTextListenPort:         3,
				ServiceMqttTlsListenPort:               4,
				ServiceMqttTlsWebSocketListenPort:      5,
				ServiceMqttWebSocketListenPort:         5,
				ServiceRestIncomingPlainTextListenPort: 6,
				ServiceRestIncomingTlsListenPort:       7,
			},
		},
	}
	var c = SolaceClientUsernamesResp{
		Data: []SolaceClientUsernameResp{
			{
				ClientUsername: "testUser",
				Enabled:        false,
				MsgVpnName:     "test1",
				Ports:          []int32{},
			},
		},
	}
	got := c.MergeSolaceResponses(vpn)
	if got.Data[0].ClientUsername == c.Data[0].ClientUsername &&
		got.Data[0].Ports[6] == 7 {
		t.Errorf("got %v, wanted %v", got, nil)
	}
}
*/
