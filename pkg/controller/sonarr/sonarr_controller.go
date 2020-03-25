package sonarr

import (
	"context"
	"fmt"
	"github.com/parflesh/sonarr-operator/defaults"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"reflect"
	"strings"
	"time"

	sonarrv1alpha1 "github.com/parflesh/sonarr-operator/pkg/apis/sonarr/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	return &ReconcileSonarr{client: mgr.GetClient(), scheme: mgr.GetScheme()}
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
		newStatus.Phase = "Initializing"
		newStatus.Reason = "Setting default spec settings"
		newStatus.Deployments = r.checkDeploymentStatus(&appsv1.Deployment{})
		newStatus.Image = instance.Spec.Image
		_ = r.updateStatus(newStatus, instance)
		return reconcile.Result{Requeue: true}, nil
	}

	/*imageManifest, err := r.imageInspector.GetImageLabels(ctx, instance.Spec.Image)
	if err != nil {
		return reconcile.Result{}, err
	}*/

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
		newStatus.Phase = "Initializing"
		newStatus.Reason = "Created deployment"
		_ = r.updateStatus(newStatus, instance)
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	newStatus.Deployments = r.checkDeploymentStatus(foundDep)
	_ = r.updateStatus(newStatus, instance)

	if err := r.reconcileDeployment(foundDep, newDep); err != nil {
		reqLogger.Error(err, "Deployment.Namespace", foundDep.Namespace, "Deployment.Name", foundDep.Name)
		if err := r.client.Update(context.TODO(), foundDep); err != nil {
			return reconcile.Result{}, err
		}
		newStatus.Image = instance.Spec.Image
		newStatus.Phase = "Updating"
		newStatus.Reason = "Updating deployment"
		_ = r.updateStatus(newStatus, instance)
		return reconcile.Result{Requeue: true}, nil
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
		newStatus.Phase = "Initializing"
		newStatus.Reason = "Created service"
		_ = r.updateStatus(newStatus, instance)
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	if len(newStatus.Deployments[appsv1.DeploymentAvailable]) > 0 {
		newStatus.Phase = string(appsv1.DeploymentAvailable)
		newStatus.Reason = ""
	} else if len(newStatus.Deployments[appsv1.DeploymentProgressing]) > 0 {
		newStatus.Phase = string(appsv1.DeploymentProgressing)
		newStatus.Reason = "Deployment progressing"
	} else if len(newStatus.Deployments[appsv1.DeploymentReplicaFailure]) > 0 {
		newStatus.Phase = string(appsv1.DeploymentReplicaFailure)
		newStatus.Reason = "Deployment replica failure"
	}
	_ = r.updateStatus(newStatus, instance)

	requeueTime, err := time.ParseDuration(instance.Spec.WatchFrequency)
	if err != nil {
		return reconcile.Result{RequeueAfter: time.Second * 60}, nil
	}
	return reconcile.Result{RequeueAfter: requeueTime}, nil
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

	volumes, volumeMounts, err := r.parseVolumes(cr.Spec.Volumes)
	if err != nil {
		return &appsv1.Deployment{}, err
	}

	var imagePullSecrets []corev1.LocalObjectReference
	for _, s := range cr.Spec.ImagePullSecrets {
		imagePullSecrets = append(imagePullSecrets, corev1.LocalObjectReference{Name: s})
	}

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
					Volumes: volumes,
					Containers: []corev1.Container{
						{
							Name:  "sonarr",
							Image: cr.Spec.Image,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8989,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Resources:    corev1.ResourceRequirements{},
							VolumeMounts: volumeMounts,
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
					ImagePullSecrets:  imagePullSecrets,
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
		dep.Spec.Template.Spec.SecurityContext.RunAsUser = &cr.Spec.RunAsUser
	}

	if cr.Spec.RunAsGroup != int64(0) {
		dep.Spec.Template.Spec.SecurityContext.RunAsUser = &cr.Spec.RunAsUser
	}

	if cr.Spec.FSGroup != int64(0) {
		dep.Spec.Template.Spec.SecurityContext.FSGroup = &cr.Spec.FSGroup
	}

	err = controllerutil.SetControllerReference(cr, dep, r.scheme)
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

func (r *ReconcileSonarr) reconcileDeployment(f *appsv1.Deployment, p *appsv1.Deployment) error {
	if !reflect.DeepEqual(f.Spec.Template.Spec.Volumes, p.Spec.Template.Spec.Volumes) || !reflect.DeepEqual(f.Spec.Template.Spec.Containers[0].VolumeMounts, p.Spec.Template.Spec.Containers[0].VolumeMounts) {
		f.Spec.Template.Spec.Volumes = p.Spec.Template.Spec.Volumes
		f.Spec.Template.Spec.Containers[0].VolumeMounts = p.Spec.Template.Spec.Containers[0].VolumeMounts
		return fmt.Errorf("deployment volumes/volumemounts mismatch")
	}

	if f.Spec.Template.Spec.PriorityClassName != p.Spec.Template.Spec.PriorityClassName {
		f.Spec.Template.Spec.PriorityClassName = p.Spec.Template.Spec.PriorityClassName
		return fmt.Errorf("priority class name mismatch")
	}

	if !reflect.DeepEqual(f.Spec.Template.Spec.SecurityContext.RunAsUser, p.Spec.Template.Spec.SecurityContext.RunAsUser) {
		f.Spec.Template.Spec.SecurityContext.RunAsUser = p.Spec.Template.Spec.SecurityContext.RunAsUser
		return fmt.Errorf("user mismatch")
	}

	if !reflect.DeepEqual(f.Spec.Template.Spec.SecurityContext.RunAsGroup, p.Spec.Template.Spec.SecurityContext.RunAsGroup) {
		f.Spec.Template.Spec.SecurityContext.RunAsGroup = p.Spec.Template.Spec.SecurityContext.RunAsGroup
		return fmt.Errorf("group mismatch")
	}

	if !reflect.DeepEqual(f.Spec.Template.Spec.SecurityContext.FSGroup, p.Spec.Template.Spec.SecurityContext.FSGroup) {
		f.Spec.Template.Spec.SecurityContext.FSGroup = p.Spec.Template.Spec.SecurityContext.FSGroup
		return fmt.Errorf("filesystem group mismatch")
	}

	if f.Spec.Template.Spec.Containers[0].Image != p.Spec.Template.Spec.Containers[0].Image {
		f.Spec.Template.Spec.Containers[0].Image = p.Spec.Template.Spec.Containers[0].Image
		return fmt.Errorf("image mismatch")
	}

	if !reflect.DeepEqual(f.Spec.Template.Spec.ImagePullSecrets, p.Spec.Template.Spec.ImagePullSecrets) {
		f.Spec.Template.Spec.ImagePullSecrets = p.Spec.Template.Spec.ImagePullSecrets
		return fmt.Errorf("image pull secrets mismatch")
	}

	if !reflect.DeepEqual(f.Labels, p.Labels) || !reflect.DeepEqual(f.Spec.Template.Labels, p.Spec.Template.Labels) || !reflect.DeepEqual(f.Spec.Selector.MatchLabels, p.Spec.Selector.MatchLabels) {
		f.Labels = p.Labels
		f.Spec.Template.Labels = p.Spec.Template.Labels
		f.Spec.Selector.MatchLabels = p.Spec.Selector.MatchLabels
		return fmt.Errorf("labels mismatch")
	}

	if *f.Spec.Replicas != *p.Spec.Replicas {
		f.Spec.Replicas = p.Spec.Replicas
		return fmt.Errorf("replicas mismatch")
	}
	return nil
}

func (r *ReconcileSonarr) labelsForCR(cr *sonarrv1alpha1.Sonarr) map[string]string {
	return map[string]string{
		"sonarr": cr.Name,
	}
}

func (r *ReconcileSonarr) parseVolumes(cr []sonarrv1alpha1.SonarrSpecVolume) ([]corev1.Volume, []corev1.VolumeMount, error) {
	var volumes []corev1.Volume
	var volumeMounts []corev1.VolumeMount

	for _, vol := range cr {
		volumes = append(volumes, corev1.Volume{
			Name: vol.Name,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: vol.Claim,
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      vol.Name,
			MountPath: vol.MountPath,
			SubPath:   vol.SubPath,
		})
	}

	return volumes, volumeMounts, nil
}

func (r *ReconcileSonarr) updateStatus(status sonarrv1alpha1.SonarrStatus, cr *sonarrv1alpha1.Sonarr) error {
	if !reflect.DeepEqual(status, cr.Status) {
		cr.Status = status
		if err := r.client.Status().Update(context.TODO(), cr); err != nil {
			reqLogger := log.WithValues("Request.Namespace", cr.Namespace, "Request.Name", cr.Name)
			reqLogger.Error(err, "Status", status)
			return err
		}
	}
	return nil
}

func (r *ReconcileSonarr) checkDeploymentStatus(dep *appsv1.Deployment) map[appsv1.DeploymentConditionType][]string {
	output := map[appsv1.DeploymentConditionType][]string{
		appsv1.DeploymentAvailable:      {},
		appsv1.DeploymentReplicaFailure: {},
		appsv1.DeploymentProgressing:    {},
	}

	for _, c := range dep.Status.Conditions {
		if c.Status == corev1.ConditionTrue {
			if c.Type == appsv1.DeploymentProgressing && strings.Contains(c.Message, "has successfully progressed") {
				continue
			}
			output[c.Type] = append(output[c.Type], dep.Name)
		}
	}

	if len(output[appsv1.DeploymentProgressing]) > 0 {
		output[appsv1.DeploymentAvailable] = []string{}
	}
	if len(output[appsv1.DeploymentReplicaFailure]) > 0 {
		output[appsv1.DeploymentAvailable] = []string{}
		output[appsv1.DeploymentProgressing] = []string{}

	}

	return output
}
