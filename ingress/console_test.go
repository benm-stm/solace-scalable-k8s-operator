package ingress

import (
	"context"
	"strconv"
	"testing"

	libs "github.com/benm-stm/solace-scalable-k8s-operator/common"
	"github.com/benm-stm/solace-scalable-k8s-operator/tests"
)

func TestNewConsole(t *testing.T) {
	got := NewConsole(
		&tests.SolaceScalable,
		libs.Labels(&tests.SolaceScalable),
	)
	if got == nil {
		t.Errorf("got %v, wanted %v", got, nil)
	}
}

func TestCreateConsoleRules(t *testing.T) {
	got := CreateConsoleRules(
		&tests.SolaceScalable,
	)
	for i := 0; i < int(tests.SolaceScalable.Spec.Replicas); i++ {
		if got[0].IngressRuleValue.HTTP.Paths[i].Backend.Service.Name != tests.SolaceScalable.ObjectMeta.Namespace+"-"+strconv.Itoa(
			i,
		) {
			t.Errorf("got %v, wanted %v", got, nil)
		}
	}
}

func TestCreateConsole(t *testing.T) {
	r, _, err := MockHaproxyReconciler()
	if err != nil {
		t.Errorf("object mock fail")
	}

	err = CreateConsole(
		&tests.SolaceScalable,
		NewConsole(&tests.SolaceScalable, libs.Labels(&tests.SolaceScalable)),
		r,
		context.TODO(),
	)
	if err != nil {
		t.Errorf("got %v, wanted %v", err, nil)
	}

}
