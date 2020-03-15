package sonarr

import (
	"testing"

	sonarrv1alpha1 "github.com/parflesh/sonarr-operator/pkg/apis/sonarr/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestSonarrController(t *testing.T) {
	var (
		name      = "sonarr-operator"
		namespace = "sonarr"
	)
	// A Sonarr object with metadata and spec.
	cr := &sonarrv1alpha1.Sonarr{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: sonarrv1alpha1.SonarrSpec{},
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{cr}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(sonarrv1alpha1.SchemeGroupVersion, cr)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)

	// Create a ReconcileSonarr object with the scheme and fake client.
	r := &ReconcileSonarr{client: cl, scheme: s}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource .
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}
	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if res.Requeue {
		t.Error("reconcile requeued even though all should be good")
	}
}
