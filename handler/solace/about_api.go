package solace

import (
	"context"
	"encoding/json"

	scalablev1alpha1 "github.com/benm-stm/solace-scalable-k8s-operator/api/v1alpha1"
	handler "github.com/benm-stm/solace-scalable-k8s-operator/handler"
)

type aboutApi struct {
	Data struct {
		Platform    string `json:"platform"`
		SempVersion string `json:"sempVersion"`
	}
}

const aboutApiPath = "/monitor/about/api"

func NewAboutApi() *aboutApi {
	return &aboutApi{}
}

func (a *aboutApi) GetInfos(
	i int,
	s *scalablev1alpha1.SolaceScalable,
	ctx context.Context,
	p string,
) error {
	url := handler.ConstructSempUrl(*s, i, aboutApiPath, nil)
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

func (a *aboutApi) GetPlatform() string {
	return a.Data.Platform
}

func (a *aboutApi) GetSempVersion() string {
	return a.Data.SempVersion
}
