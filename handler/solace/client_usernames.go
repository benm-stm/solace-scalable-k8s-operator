package solace

import (
	"context"
	"encoding/json"
	"strings"

	scalablev1alpha1 "github.com/benm-stm/solace-scalable-k8s-operator/api/v1alpha1"
	libs "github.com/benm-stm/solace-scalable-k8s-operator/common"
	handler "github.com/benm-stm/solace-scalable-k8s-operator/handler"
)

// ClientUsername response struct
type ClientUsername struct {
	ClientUsername string                    `json:"clientUsername"`
	Enabled        bool                      `json:"enabled"`
	MsgVpnName     string                    `json:"msgVpnName"`
	Attributes     []ClientUsernameAttribute `json:"attributes"`
	Ports          []int32                   `json:"ports"`
}

// ClientUsernames array response struct
type ClientUsernames struct {
	Data []ClientUsername `json:"data"`
}

// const clientUsernamePath = "/monitor/about/api"
// var clientUsernamePath = "/config/msgVpns/" + m.MsgVpnName + "/clientUsernames"
var clientUsernamePath = "/config/msgVpns/<MsgVpn>/clientUsernames"

var urlValues = map[string]string{
	"select": "clientUsername,enabled,msgVpnName",
	"where":  "clientUsername!=*client-username",
}

func NewClientUsernames() *ClientUsernames {
	return &ClientUsernames{}
}

// Returns the solace's clientUsernames per msgVpn in Json format
func (c *ClientUsernames) Add(
	s *scalablev1alpha1.SolaceScalable,
	node int,
	msgVpn *msgVpn,
	pwd string,
	ctx context.Context,
) error {
	path := strings.Replace(clientUsernamePath, "<MsgVpn>", msgVpn.MsgVpnName, 1)
	url := handler.ConstructSempUrl(*s, node, path, urlValues)
	textBytes, _, _ := handler.CallSolaceSempApi(
		url,
		ctx,
		pwd,
	)

	resp := ClientUsernames{}
	err := json.Unmarshal(textBytes, &resp)
	if err != nil {
		return err
	}
	// sanityze clientUsername
	for i := range resp.Data {
		resp.Data[i].ClientUsername =
			libs.SanityzeForSvcName(resp.Data[i].ClientUsername)
	}
	c.Data = append(
		c.Data,
		resp.Data...,
	)
	return nil
}
