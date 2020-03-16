package sonarr

import (
	"context"
	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"testing"

	sonarrv1alpha1 "github.com/parflesh/sonarr-operator/pkg/apis/sonarr/v1alpha1"
	"github.com/parflesh/sonarr-operator/pkg/registry_client"

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
	cl := fake.NewFakeClientWithScheme(s, objs...)

	registryClient := &registry_client.MockRegistryClient{}
	// Create a ReconcileSonarr object with the scheme and fake client.
	r := &ReconcileSonarr{
		client:                 cl,
		scheme:                 s,
		registryClientProvider: &registry_client.MockRegistryClientProvider{Client: registryClient},
	}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource .
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	// Check spec updates
	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue")
	}
	err = r.client.Get(context.TODO(), req.NamespacedName, cr)
	if cr.Spec.Image == "" {
		t.Error("Image spec not updated")
	}

	res, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue")
	}
	err = r.client.Get(context.TODO(), req.NamespacedName, cr)
	if cr.Spec.WatchFrequency == "" {
		t.Error("Watch Frequency not updated")
	}

	registryClient.ManifestV2Output = &schema2.DeserializedManifest{
		Manifest: schema2.Manifest{
			Config: distribution.Descriptor{
				Digest: "abc123",
			},
		},
	}
	res, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue")
	}
	err = r.client.Get(context.TODO(), req.NamespacedName, cr)
	if cr.Status.Image != "quay.io/parflesh/sonarr@sha256:abc123" {
		t.Error("status image mismatch")
	}

	depDep := &appsv1.Deployment{}
	res, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue")
	}
	err = r.client.Get(context.TODO(), req.NamespacedName, depDep)
	if err != nil {
		t.Error("Deployment not created")
	}

	depSvc := &corev1.Service{}
	res, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue")
	}
	err = r.client.Get(context.TODO(), req.NamespacedName, depSvc)
	if err != nil {
		t.Error("Service not created")
	}

	// Everything should be good.  Lets check
	res, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if res.Requeue {
		t.Error("reconcile requeued even though all should be good")
	}
}
