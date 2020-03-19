package sonarr

import (
	"context"
	"fmt"
	"github.com/parflesh/sonarr-operator/defaults"
	sonarrv1alpha1 "github.com/parflesh/sonarr-operator/pkg/apis/sonarr/v1alpha1"
	//"github.com/parflesh/sonarr-operator/pkg/image_inspect"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_sonarr")

// Add creates a new Sonarr Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileSonarr{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		//imageInspector: &image_inspect.ImageInspector{},
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("sonarr-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Sonarr
	err = c.Watch(&source.Kind{Type: &sonarrv1alpha1.Sonarr{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &sonarrv1alpha1.Sonarr{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &sonarrv1alpha1.Sonarr{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileSonarr implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileSonarr{}

// ReconcileSonarr reconciles a Sonarr object
type ReconcileSonarr struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	//imageInspector image_inspect.ImageInspectorInterface
}

func (r *ReconcileSonarr) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Sonarr")

	// Fetch the Sonarr instance
	instance := &sonarrv1alpha1.Sonarr{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	newStatus := instance.Status

	err = r.reconcileSpec(instance)
	if err != nil {
		err := r.client.Update(context.TODO(), instance)
		if err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	/*imageManifest, err := r.imageInspector.GetImageLabels(ctx, instance.Spec.Image)
	if err != nil {
		return reconcile.Result{}, err
	}*/

	newStatus.Image = instance.Spec.Image
	if newStatus.Image != instance.Status.Image {
		instance.Status.Image = newStatus.Image
		if err := r.client.Status().Update(context.TODO(), instance); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	newDep, err := r.newDeployment(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	foundDep := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), request.NamespacedName, foundDep)
	if err != nil && errors.IsNotFound(err) {
		err := r.client.Create(context.TODO(), newDep)
		if err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	newSvc, err := r.newService(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	foundSvc := &corev1.Service{}
	err = r.client.Get(context.TODO(), request.NamespacedName, foundSvc)
	if err != nil && errors.IsNotFound(err) {
		err := r.client.Create(context.TODO(), newSvc)
		if err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileSonarr) reconcileSpec(cr *sonarrv1alpha1.Sonarr) error {
	if cr.Spec.Image == "" {
		cr.Spec.Image = defaults.SonarrImage
		return fmt.Errorf("image not set")
	}
	if cr.Spec.WatchFrequency == "" {
		cr.Spec.WatchFrequency = defaults.OperatorRequeuTime
		return fmt.Errorf("watch frequency not set")
	}
	return nil
}

func (r *ReconcileSonarr) newDeployment(cr *sonarrv1alpha1.Sonarr) (*appsv1.Deployment, error) {
	labels := r.labelsForCR(cr)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &[]int32{1}[0],
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name:         "config",
							VolumeSource: cr.Spec.ConfigVolume,
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "sonarr",
							Image: cr.Status.Image,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8989,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Resources: corev1.ResourceRequirements{},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "config",
									MountPath: "/config",
								},
							},
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "",
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: 8989,
											StrVal: "",
										},
										Scheme: corev1.URISchemeHTTP,
									},
								},
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "",
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: 8989,
											StrVal: "",
										},
										Scheme: corev1.URISchemeHTTP,
									},
								},
							},
							ImagePullPolicy: corev1.PullIfNotPresent,
						},
					},
					RestartPolicy:     corev1.RestartPolicyAlways,
					SecurityContext:   &corev1.PodSecurityContext{},
					ImagePullSecrets:  cr.Spec.ImagePullSecrets,
					PriorityClassName: cr.Spec.PriorityClassName,
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			RevisionHistoryLimit: &[]int32{5}[0],
		},
	}

	if cr.Spec.RunAsUser != int64(0) {
		dep.Spec.Template.Spec.Containers[0].SecurityContext.RunAsUser = &cr.Spec.RunAsUser
	}

	if cr.Spec.RunAsGroup != int64(0) {
		dep.Spec.Template.Spec.Containers[0].SecurityContext.RunAsUser = &cr.Spec.RunAsUser
	}

	err := controllerutil.SetControllerReference(cr, dep, r.scheme)
	if err != nil {
		return dep, err
	}
	return dep, nil
}

func (r *ReconcileSonarr) newService(cr *sonarrv1alpha1.Sonarr) (*corev1.Service, error) {
	labels := r.labelsForCR(cr)

	dep := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     "http",
					Protocol: corev1.ProtocolTCP,
					Port:     8989,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 8989,
						StrVal: "",
					},
				},
			},
			Selector: labels,
		},
	}

	err := controllerutil.SetControllerReference(cr, dep, r.scheme)
	if err != nil {
		return dep, err
	}

	return dep, nil
}

func (r *ReconcileSonarr) labelsForCR(cr *sonarrv1alpha1.Sonarr) map[string]string {
	return map[string]string{
		"sonarr": cr.Name,
	}
}
