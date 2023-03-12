package solace

import (
	"context"
	"encoding/json"

	scalablev1alpha1 "github.com/benm-stm/solace-scalable-k8s-operator/api/v1alpha1"
	handler "github.com/benm-stm/solace-scalable-k8s-operator/handler"
)

type AboutApi struct {
	Data struct {
		Platform    string
		SempVersion string
	}
}

const aboutApiPath = "/monitor/about/api"

func (a *AboutApi) GetInfos(
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
	//fmt.Println(string(textBytes))
	if err := json.Unmarshal(textBytes, &a); err != nil {
		return err
	}
	return nil
}
