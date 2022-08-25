package controllers

import (
	"context"
	"encoding/json"

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
type SolaceClientUsernameResp struct {
	ClientUsername string  `json:"clientUsername"`
	Enabled        bool    `json:"enabled"`
	MsgVpnName     string  `json:"msgVpnName"`
	Ports          []int32 `json:"ports"`
}
type SolaceClientUsernamesResp struct {
	Data []SolaceClientUsernameResp `json:"data"`
}

//merged datas
type SolaceMergedResp struct {
	MsgVpnName     string  `json:"msgVpnName"`
	ClientUsername string  `json:"clientUsername"`
	Ports          []int32 `json:"ports"`
}
type SolaceMergedResps struct {
	Data []SolaceMergedResp `json:"data"`
}

func GetSolaceOpenPorts(s *scalablev1alpha1.SolaceScalable, ctx context.Context) ([]int32, error) {
	var ports []int32
	bodyText, _, err := CallSolaceSempApi(s, "/config/msgVpns", ctx, solaceAdminPassword)
	if err != nil {
		return nil, err
	}
	ports = Unique(CleanJsonResponse(bodyText, ".*Port\":(.*),"))
	return ports, nil
}

func GetEnabledSolaceMsgVpns(
	s *scalablev1alpha1.SolaceScalable,
	ctx context.Context,
) (SolaceMsgVpnsResp, error) {
	text, _, err := CallSolaceSempApi(
		s,
		"/config/msgVpns?select="+
			"msgVpnName,enabled,*Port"+
			"&where=enabled==true",
		ctx,
		solaceAdminPassword,
	)
	if err != nil {
		return SolaceMsgVpnsResp{}, err
	}
	textBytes := []byte(text)

	resp := SolaceMsgVpnsResp{}
	err = json.Unmarshal(textBytes, &resp)
	if err != nil {
		return SolaceMsgVpnsResp{}, err
	}
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
	return resp, nil
}

func (c *SolaceClientUsernamesResp) MergeSolaceResponses(m SolaceMsgVpnsResp) SolaceMergedResps {
	resp := SolaceMergedResps{}
	res := SolaceMergedResp{}

	for _, m := range m.Data {
		for _, c := range c.Data {
			// remove element if clientusername is disabled
			res = SolaceMergedResp{}
			if c.MsgVpnName == m.MsgVpnName {
				res.MsgVpnName = c.MsgVpnName

				res.ClientUsername = c.ClientUsername

				res.Ports = append(res.Ports, int32(m.ServiceAmqpPlainTextListenPort))
				res.Ports = append(res.Ports, int32(m.ServiceAmqpTlsListenPort))
				res.Ports = append(res.Ports, int32(m.ServiceMqttPlainTextListenPort))
				res.Ports = append(res.Ports, int32(m.ServiceMqttTlsListenPort))
				res.Ports = append(res.Ports, int32(m.ServiceMqttTlsWebSocketListenPort))
				res.Ports = append(res.Ports, int32(m.ServiceMqttWebSocketListenPort))
				res.Ports = append(res.Ports, int32(m.ServiceRestIncomingPlainTextListenPort))
				res.Ports = append(res.Ports, int32(m.ServiceRestIncomingTlsListenPort))
				resp.Data = append(resp.Data, res)
			}
		}
	}
	return resp
}
