package controllers

import (
	"testing"

	scalablev1alpha1 "github.com/benm-stm/solace-scalable-k8s-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var solaceScalable = scalablev1alpha1.SolaceScalable{
	TypeMeta: metav1.TypeMeta{},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test",
		Namespace: "test",
		Labels:    map[string]string{},
	},
	Spec:   scalablev1alpha1.SolaceScalableSpec{},
	Status: scalablev1alpha1.SolaceScalableStatus{},
}

func TestLabels(t *testing.T) {
	got := Labels(&solaceScalable)
	want := map[string]string{
		"app": "test",
	}
	if got["app"] != want["app"] {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

func TestStringInSlice(t *testing.T) {
	got := StringInSlice("a", []string{"a", "b", "c"})
	want := true
	if got != want {
		t.Errorf("got %v, wanted %v", got, want)
	}
}

func TestAsSha256(t *testing.T) {
	got := AsSha256("test")
	want := "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
	if got != want {
		t.Errorf("got %s, wanted %s", got, want)
	}
}

func TestUnique(t *testing.T) {
	got := UniqueAndNonZero([]int32{1, 2, 3, 2, 0})
	want := []int32{1, 2, 3}
	if len(got) != len(want) {
		t.Errorf("got %v, wanted %v", got, want)
	}
	for i, v := range got {
		if v != want[i] {
			t.Errorf("got %v, wanted %v", got, want)
		}
	}
}

func TestCleanJsonResponse(t *testing.T) {
	got := CleanJsonResponse("testPort\":8000,test2Port\":8001",
		".*Port\":(.*),",
	)
	want := []int32{8000, 8001}
	if CheckEq(got, want) {
		t.Errorf("got %v, wanted %v", got, want)
	}
}

func CheckEq(a, b []int32) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
