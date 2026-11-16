package vip

import (
	"context"
	"reflect"
	"sort"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	l2keeperv1alpha1 "github.com/l2keeper/l2keeper/api/v1alpha1"
)

func TestBindingSource(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := l2keeperv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "ingress", Namespace: "default"}}
	explicit := &l2keeperv1alpha1.VIPBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "explicit", Namespace: "default"},
		Spec: l2keeperv1alpha1.VIPBindingSpec{
			ServiceRef: l2keeperv1alpha1.ServiceReference{Name: "ingress"},
			Mode:       l2keeperv1alpha1.VIPModeExternal,
			VIPs:       []string{"10.99.50.100"},
		},
		Status: l2keeperv1alpha1.VIPBindingStatus{AllocatedIPs: []string{"10.99.50.100"}},
	}
	invalid := &l2keeperv1alpha1.VIPBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "invalid", Namespace: "default"},
		Spec: l2keeperv1alpha1.VIPBindingSpec{
			ServiceRef: l2keeperv1alpha1.ServiceReference{Name: "ingress"},
			Mode:       l2keeperv1alpha1.VIPModeL2,
			VIPs:       []string{"10.99.50.101"},
			PoolRef:    &l2keeperv1alpha1.PoolReference{Name: "pool"},
		},
	}
	pooled := &l2keeperv1alpha1.VIPBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "pooled", Namespace: "default"},
		Spec: l2keeperv1alpha1.VIPBindingSpec{
			ServiceRef: l2keeperv1alpha1.ServiceReference{Name: "ingress"},
			Mode:       l2keeperv1alpha1.VIPModeL2,
			PoolRef:    &l2keeperv1alpha1.PoolReference{Name: "pool"},
		},
		Status: l2keeperv1alpha1.VIPBindingStatus{AllocatedIPs: []string{"192.168.56.240"}},
	}
	unrelated := &l2keeperv1alpha1.VIPBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "unrelated", Namespace: "default"},
		Spec: l2keeperv1alpha1.VIPBindingSpec{
			ServiceRef: l2keeperv1alpha1.ServiceReference{Name: "other"},
			VIPs:       []string{"10.99.50.200"},
		},
	}

	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(svc, explicit, invalid, pooled, unrelated).
		Build()

	got, err := BindingSource{Client: c}.VIPs(context.Background(), svc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sort.Slice(got, func(i, j int) bool { return got[i].IP < got[j].IP })
	want := []Spec{
		{IP: "10.99.50.100", Mode: ModeExternal},
		{IP: "192.168.56.240", Mode: ModeL2},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected VIPs: got %v, want %v", got, want)
	}
}
