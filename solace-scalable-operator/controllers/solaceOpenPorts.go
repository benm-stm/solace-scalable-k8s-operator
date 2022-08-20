package controllers

import (
	"encoding/json"

	scalablev1alpha1 "solace.io/api/v1alpha1"
)

var blacklistedClientUsernames = []string{"#client-username"}

// message vpns response struct
type solaceMsgVpnResp struct {
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

type solaceMsgVpnsResp struct {
	Data []solaceMsgVpnResp `json:"data"`
}

//******************************************************
//clientUsernames response struct
type solaceClientUsernameResp struct {
	ClientUsername string  `json:"clientUsername"`
	Enabled        bool    `json:"enabled"`
	MsgVpnName     string  `json:"msgVpnName"`
	Ports          []int32 `json:"ports"`
}

type solaceClientUsernamesResp struct {
	Data []solaceClientUsernameResp `json:"data"`
}

//******************************************************
//merged datas
type solaceMergedResp struct {
	MsgVpnName     string  `json:"msgVpnName"`
	ClientUsername string  `json:"clientUsername"`
	Ports          []int32 `json:"ports"`
}
type solaceMergedResps struct {
	Data []solaceMergedResp `json:"data"`
}

//*****************************************************

func getSolaceOpenPorts(s *scalablev1alpha1.SolaceScalable) ([]int32, error) {
	var ports []int32
	bodyText, err := CallSolaceSempApi(s, "/config/msgVpns")
	if err != nil {
		return nil, err
	}
	ports = unique(cleanJsonResponse(bodyText, ".*Port\":(.*),"))
	return ports, nil
}

func getEnabledSolaceMsgVpns(s *scalablev1alpha1.SolaceScalable) (solaceMsgVpnsResp, error) {
	text, err := CallSolaceSempApi(s, "/config/msgVpns?select=msgVpnName,enabled,*Port&where=enabled==true")
	if err != nil {
		return solaceMsgVpnsResp{}, err
	}
	textBytes := []byte(text)

	resp := solaceMsgVpnsResp{}
	err = json.Unmarshal(textBytes, &resp)
	if err != nil {
		return solaceMsgVpnsResp{}, err
	}
	return resp, nil
}

func getSolaceClientUsernames(s *scalablev1alpha1.SolaceScalable, r solaceMsgVpnsResp) (solaceClientUsernamesResp, error) {
	temp := solaceClientUsernamesResp{}
	resp := solaceClientUsernamesResp{}
	for i := 0; i < len(r.Data); i++ {
		//ignore #client-username
		text, err := CallSolaceSempApi(s, "/config/msgVpns/"+r.Data[i].MsgVpnName+"/clientUsernames?select=clientUsername,enabled,msgVpnName&where=clientUsername!=*client-username")
		if err != nil {
			return solaceClientUsernamesResp{}, err
		}
		textBytes := []byte(text)
		err = json.Unmarshal(textBytes, &temp)
		if err != nil {
			return solaceClientUsernamesResp{}, err
		}

		resp.Data = append(resp.Data, temp.Data...)
	}
	return resp, nil
}

func mergeSolaceResponses(m solaceMsgVpnsResp, c solaceClientUsernamesResp) solaceMergedResps {
	resp := solaceMergedResps{}
	res := solaceMergedResp{}

	for i := 0; i < len(m.Data); i++ {
		for j := 0; j < len(c.Data); j++ {
			// remove element if clientusername is disabled
			res = solaceMergedResp{}
			if c.Data[j].MsgVpnName == m.Data[i].MsgVpnName {
				res.MsgVpnName = c.Data[j].MsgVpnName

				res.ClientUsername = c.Data[j].ClientUsername

				res.Ports = append(res.Ports, int32(m.Data[i].ServiceAmqpPlainTextListenPort))
				res.Ports = append(res.Ports, int32(m.Data[i].ServiceAmqpTlsListenPort))
				res.Ports = append(res.Ports, int32(m.Data[i].ServiceMqttPlainTextListenPort))
				res.Ports = append(res.Ports, int32(m.Data[i].ServiceMqttTlsListenPort))
				res.Ports = append(res.Ports, int32(m.Data[i].ServiceMqttTlsWebSocketListenPort))
				res.Ports = append(res.Ports, int32(m.Data[i].ServiceMqttWebSocketListenPort))
				res.Ports = append(res.Ports, int32(m.Data[i].ServiceRestIncomingPlainTextListenPort))
				res.Ports = append(res.Ports, int32(m.Data[i].ServiceRestIncomingTlsListenPort))
				resp.Data = append(resp.Data, res)
			}
		}
	}
	return resp
}
