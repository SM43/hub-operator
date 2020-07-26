package hub

import (
	"context"
	"fmt"

	routev1 "github.com/openshift/api/route/v1"
	hubv1alpha1 "github.com/sm43/hub-operator/pkg/apis/hub/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_hub")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Hub Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileHub{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("hub-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Hub
	err = c.Watch(&source.Kind{Type: &hubv1alpha1.Hub{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileHub implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileHub{}

// ReconcileHub reconciles a Hub object
type ReconcileHub struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Hub object and makes changes based on the state read
// and what is in the Hub.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileHub) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Hub")

	// Fetch the Hub instance
	instance := &hubv1alpha1.Hub{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	foundDB := &v1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: "db", Namespace: instance.Namespace}, foundDB)
	if err != nil && errors.IsNotFound(err) {

		dbSecret := r.CreateDBSecret(instance)
		err = r.client.Create(context.TODO(), dbSecret)
		if err != nil {
			reqLogger.Error(err, "Failed to create db secret")
			return reconcile.Result{}, err
		}

		db := r.CreateDBDeployment(instance)
		err = r.client.Create(context.TODO(), db)
		if err != nil {
			reqLogger.Error(err, "Failed to create db deployment")
			return reconcile.Result{}, err
		}

		dbService := r.CreateDBService(instance)
		err = r.client.Create(context.TODO(), dbService)
		if err != nil {
			reqLogger.Error(err, "Failed to create db service")
			return reconcile.Result{}, err
		}

		apiSecret := r.CreateAPISecret(instance)
		err = r.client.Create(context.TODO(), apiSecret)
		if err != nil {
			reqLogger.Error(err, "Failed to create api secret")
			return reconcile.Result{}, err
		}

		api := r.CreateAPIDeployment(instance)
		err = r.client.Create(context.TODO(), api)
		if err != nil {
			reqLogger.Error(err, "Failed to create api deployment")
			return reconcile.Result{}, err
		}

		apiService := r.CreateAPIService(instance)
		err = r.client.Create(context.TODO(), apiService)
		if err != nil {
			reqLogger.Error(err, "Failed to create api service")
			return reconcile.Result{}, err
		}

		apiRoute := r.CreateAPIRoute(instance)
		err = r.client.Create(context.TODO(), apiRoute)
		if err != nil {
			reqLogger.Error(err, "Failed to create api route")
			return reconcile.Result{}, err
		}

		foundAPIRoute := &routev1.Route{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: "api", Namespace: instance.Namespace}, foundAPIRoute)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Error(err, "Service Not found")
			return reconcile.Result{}, err
		}

		route := "https://" + foundAPIRoute.Spec.Host
		uiConfig := r.CreateUIConfigMap(instance, route)
		err = r.client.Create(context.TODO(), uiConfig)
		if err != nil {
			reqLogger.Error(err, "Failed to create ui config map")
			return reconcile.Result{}, err
		}

		ui := r.CreateUIDeployment(instance)
		err = r.client.Create(context.TODO(), ui)
		if err != nil {
			reqLogger.Error(err, "Failed to create ui deployment")
			return reconcile.Result{}, err
		}

		uiService := r.CreateUIService(instance)
		err = r.client.Create(context.TODO(), uiService)
		if err != nil {
			reqLogger.Error(err, "Failed to create ui service")
			return reconcile.Result{}, err
		}

		uiRoute := r.CreateUIRoute(instance)
		err = r.client.Create(context.TODO(), uiRoute)
		if err != nil {
			reqLogger.Error(err, "Failed to create ui route")
			return reconcile.Result{}, err
		}

		dbMigration := r.CreateDBMigration(instance)
		err = r.client.Create(context.TODO(), dbMigration)
		if err != nil {
			reqLogger.Error(err, "Failed to create db migration")
			return reconcile.Result{}, err
		}

	} else {
		return reconcile.Result{}, err
	}

	status := fmt.Sprintf("Hub Deployed")
	if status != instance.Status.Status {
		instance.Status.Status = status
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "Failed to update status")
			return reconcile.Result{Requeue: false}, err
		}
	}

	return reconcile.Result{}, nil
}

// CreateDBSecret creates db secret
func (r ReconcileHub) CreateDBSecret(instance *hubv1alpha1.Hub) *corev1.Secret {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "db",
			Namespace: instance.Namespace,
			Labels:    map[string]string{"app": "db"},
		},
		StringData: map[string]string{
			"POSTGRESQL_DATABASE": instance.Spec.Db.Name,
			"POSTGRESQL_USER":     instance.Spec.Db.User,
			"POSTGRESQL_PASSWORD": instance.Spec.Db.Password,
		},
		Type: corev1.SecretTypeOpaque,
	}
	secret.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion:         instance.APIVersion,
			Kind:               instance.Kind,
			Name:               instance.Namespace,
			UID:                instance.UID,
			Controller:         nil,
			BlockOwnerDeletion: nil,
		},
	})
	return secret
}

// CreateDBDeployment creates an instance of db deployment
func (r ReconcileHub) CreateDBDeployment(instance *hubv1alpha1.Hub) *v1.Deployment {
	var replicas *int32
	replicas = new(int32)
	*replicas = 1
	d := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "db",
			Namespace: instance.Namespace,
			Labels:    map[string]string{"app": "db"},
		},
		Spec: v1.DeploymentSpec{
			Replicas: replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "db"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "db"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "db",
							Image: "postgres:13",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 5432,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name: "POSTGRES_DB",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "db",
											},
											Key: "POSTGRESQL_DATABASE",
										},
									},
								},
								{
									Name: "POSTGRES_USER",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "db",
											},
											Key: "POSTGRESQL_USER",
										},
									},
								},
								{
									Name: "POSTGRES_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "db",
											},
											Key: "POSTGRESQL_PASSWORD",
										},
									},
								},
								{
									Name:  "PGDATA",
									Value: "/var/lib/postgresql/data/pgdata",
								},
							},
						},
					},
				},
			},
		},
	}
	d.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion:         instance.APIVersion,
			Kind:               instance.Kind,
			Name:               instance.Namespace,
			UID:                instance.UID,
			Controller:         nil,
			BlockOwnerDeletion: nil,
		},
	})
	return d
}

// CreateDBService creates db service
func (r ReconcileHub) CreateDBService(instance *hubv1alpha1.Hub) *corev1.Service {
	s := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "db",
			Namespace: instance.Namespace,
			Labels:    map[string]string{"app": "db"},
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: map[string]string{"app": "db"},
			Ports: []corev1.ServicePort{
				{
					Name:     "postgresql",
					Port:     5432,
					Protocol: corev1.ProtocolTCP,
					TargetPort: intstr.IntOrString{
						IntVal: 5432,
					},
				},
			},
		},
	}
	s.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion:         instance.APIVersion,
			Kind:               instance.Kind,
			Name:               instance.Namespace,
			UID:                instance.UID,
			Controller:         nil,
			BlockOwnerDeletion: nil,
		},
	})
	return s
}

// CreateAPISecret creates api secret
func (r ReconcileHub) CreateAPISecret(instance *hubv1alpha1.Hub) *corev1.Secret {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "api",
			Namespace: instance.Namespace,
			Labels:    map[string]string{"app": "api"},
		},
		StringData: map[string]string{
			"CLIENT_ID":       instance.Spec.API.ClientID,
			"CLIENT_SECRET":   instance.Spec.API.ClientSecret,
			"JWT_SIGNING_KEY": instance.Spec.API.JwtSigningKey,
		},
		Type: corev1.SecretTypeOpaque,
	}
	secret.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion:         instance.APIVersion,
			Kind:               instance.Kind,
			Name:               instance.Namespace,
			UID:                instance.UID,
			Controller:         nil,
			BlockOwnerDeletion: nil,
		},
	})
	return secret
}

// CreateAPIDeployment creates an instance of api deployment
func (r ReconcileHub) CreateAPIDeployment(instance *hubv1alpha1.Hub) *v1.Deployment {
	var replicas *int32
	replicas = new(int32)
	*replicas = 1
	d := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "api",
			Namespace: instance.Namespace,
			Labels:    map[string]string{"app": "api"},
		},
		Spec: v1.DeploymentSpec{
			Replicas: replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "api"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "api"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "api",
							Image: instance.Spec.API.Image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8000,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "POSTGRESQL_HOST",
									Value: "db",
								},
								{
									Name: "POSTGRESQL_DATABASE",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "db",
											},
											Key: "POSTGRESQL_DATABASE",
										},
									},
								},
								{
									Name:  "POSTGRESQL_PORT",
									Value: "5432",
								},
								{
									Name: "POSTGRESQL_USER",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "db",
											},
											Key: "POSTGRESQL_USER",
										},
									},
								},
								{
									Name: "POSTGRESQL_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "db",
											},
											Key: "POSTGRESQL_PASSWORD",
										},
									},
								},
								{
									Name:  "GITHUB_TOKEN",
									Value: "",
								},
								{
									Name: "CLIENT_ID",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "api",
											},
											Key: "CLIENT_ID",
										},
									},
								},
								{
									Name: "CLIENT_SECRET",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "api",
											},
											Key: "CLIENT_SECRET",
										},
									},
								},
								{
									Name: "JWT_SIGNING_KEY",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "api",
											},
											Key: "JWT_SIGNING_KEY",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	d.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion:         instance.APIVersion,
			Kind:               instance.Kind,
			Name:               instance.Namespace,
			UID:                instance.UID,
			Controller:         nil,
			BlockOwnerDeletion: nil,
		},
	})
	return d
}

// CreateAPIService creates api service
func (r ReconcileHub) CreateAPIService(instance *hubv1alpha1.Hub) *corev1.Service {
	s := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "api",
			Namespace: instance.Namespace,
			Labels:    map[string]string{"app": "api"},
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeNodePort,
			Selector: map[string]string{"app": "api"},
			Ports: []corev1.ServicePort{
				{
					Port: 5000,
					TargetPort: intstr.IntOrString{
						IntVal: 5000,
					},
				},
			},
		},
	}
	s.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion:         instance.APIVersion,
			Kind:               instance.Kind,
			Name:               instance.Namespace,
			UID:                instance.UID,
			Controller:         nil,
			BlockOwnerDeletion: nil,
		},
	})
	return s
}

// CreateAPIRoute creates api route
func (r ReconcileHub) CreateAPIRoute(instance *hubv1alpha1.Hub) *routev1.Route {
	route := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "api",
			Namespace: instance.Namespace,
			Labels:    map[string]string{"app": "api"},
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: "api",
			},
			TLS: &routev1.TLSConfig{
				InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
				Termination:                   routev1.TLSTerminationEdge,
			},
		},
	}
	route.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion:         instance.APIVersion,
			Kind:               instance.Kind,
			Name:               instance.Namespace,
			UID:                instance.UID,
			Controller:         nil,
			BlockOwnerDeletion: nil,
		},
	})
	return route
}

// CreateDBMigration creates db migration
func (r ReconcileHub) CreateDBMigration(instance *hubv1alpha1.Hub) *batchv1.Job {
	var backOffLimit *int32
	backOffLimit = new(int32)
	*backOffLimit = 3
	j := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "db-migration",
			Namespace: instance.Namespace,
			Labels:    map[string]string{"app": "db"},
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "db-migration",
							Image: instance.Spec.Migration.Image,
							Env: []corev1.EnvVar{
								{
									Name:  "POSTGRESQL_HOST",
									Value: "db",
								},
								{
									Name: "POSTGRESQL_DATABASE",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "db",
											},
											Key: "POSTGRESQL_DATABASE",
										},
									},
								},
								{
									Name:  "POSTGRESQL_PORT",
									Value: "5432",
								},
								{
									Name: "POSTGRESQL_USER",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "db",
											},
											Key: "POSTGRESQL_USER",
										},
									},
								},
								{
									Name: "POSTGRESQL_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "db",
											},
											Key: "POSTGRESQL_PASSWORD",
										},
									},
								},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
			BackoffLimit: backOffLimit,
		},
	}
	j.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion:         instance.APIVersion,
			Kind:               instance.Kind,
			Name:               instance.Namespace,
			UID:                instance.UID,
			Controller:         nil,
			BlockOwnerDeletion: nil,
		},
	})
	return j
}

// CreateUIConfigMap creates ui configmap
func (r ReconcileHub) CreateUIConfigMap(instance *hubv1alpha1.Hub, route string) *corev1.ConfigMap {
	config := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ui",
			Namespace: instance.Namespace,
			Labels:    map[string]string{"app": "ui"},
		},
		Data: map[string]string{
			"API_URL":      route,
			"GH_CLIENT_ID": instance.Spec.API.ClientID,
		},
	}
	config.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion:         instance.APIVersion,
			Kind:               instance.Kind,
			Name:               instance.Namespace,
			UID:                instance.UID,
			Controller:         nil,
			BlockOwnerDeletion: nil,
		},
	})
	return config
}

// CreateUIDeployment creates an instance of ui deployment
func (r ReconcileHub) CreateUIDeployment(instance *hubv1alpha1.Hub) *v1.Deployment {
	var replicas *int32
	replicas = new(int32)
	*replicas = 1
	d := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ui",
			Namespace: instance.Namespace,
			Labels:    map[string]string{"app": "ui"},
		},
		Spec: v1.DeploymentSpec{
			Replicas: replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "ui"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "ui"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "ui",
							Image: instance.Spec.UI.Image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name: "GH_CLIENT_ID",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "ui",
											},
											Key: "GH_CLIENT_ID",
										},
									},
								},
								{
									Name: "API_URL",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "ui",
											},
											Key: "API_URL",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	d.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion:         instance.APIVersion,
			Kind:               instance.Kind,
			Name:               instance.Namespace,
			UID:                instance.UID,
			Controller:         nil,
			BlockOwnerDeletion: nil,
		},
	})
	return d
}

// CreateUIService creates ui service
func (r ReconcileHub) CreateUIService(instance *hubv1alpha1.Hub) *corev1.Service {
	s := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ui",
			Namespace: instance.Namespace,
			Labels:    map[string]string{"app": "ui"},
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeNodePort,
			Selector: map[string]string{"app": "ui"},
			Ports: []corev1.ServicePort{
				{
					Port: 8080,
					TargetPort: intstr.IntOrString{
						IntVal: 8080,
					},
				},
			},
		},
	}
	s.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion:         instance.APIVersion,
			Kind:               instance.Kind,
			Name:               instance.Namespace,
			UID:                instance.UID,
			Controller:         nil,
			BlockOwnerDeletion: nil,
		},
	})
	return s
}

// CreateUIRoute creates ui route
func (r ReconcileHub) CreateUIRoute(instance *hubv1alpha1.Hub) *routev1.Route {
	route := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ui",
			Namespace: instance.Namespace,
			Labels:    map[string]string{"app": "ui"},
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: "ui",
			},
			TLS: &routev1.TLSConfig{
				InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
				Termination:                   routev1.TLSTerminationEdge,
			},
		},
	}
	route.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion:         instance.APIVersion,
			Kind:               instance.Kind,
			Name:               instance.Namespace,
			UID:                instance.UID,
			Controller:         nil,
			BlockOwnerDeletion: nil,
		},
	})
	return route
}

// // CreateDBPvc creates db pvc
// func (r ReconcileHub) CreateDBPvc(instance *hubv1alpha1.Hub) *corev1.PersistentVolumeClaim {
// 	pvc := &corev1.PersistentVolumeClaim{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      "db",
// 			Namespace: instance.Namespace,
// 			Labels:    map[string]string{"app": "db"},
// 		},
// 		Spec: corev1.PersistentVolumeClaimSpec{
// 			AccessModes: []corev1.PersistentVolumeAccessMode{
// 				corev1.ReadWriteOnce,
// 			},
// 			Resources: corev1.ResourceRequirements{
// 				Requests: map[corev1.ResourceName]resource.Quantity{
// 					corev1.ResourceStorage: resource.Quantity{
// 						Format: resource.BinarySI,
// 					},
// 				},
// 			},
// 		},
// 	}
// 	pvc.SetOwnerReferences([]metav1.OwnerReference{
// 		{
// 			APIVersion:         instance.APIVersion,
// 			Kind:               instance.Kind,
// 			Name:               instance.Namespace,
// 			UID:                instance.UID,
// 			Controller:         nil,
// 			BlockOwnerDeletion: nil,
// 		},
// 	})
// 	return pvc
// }
