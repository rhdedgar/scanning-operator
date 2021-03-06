/*
Copyright 2020 Doug Edgar.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pingcap/errors"
	"github.com/rhdedgar/scanning-operator/k8s"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	managedv1alpha1 "github.com/rhdedgar/scanning-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
)

// ScannerReconciler reconciles a Scanner object
type ScannerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=managed.openshift.io,namespace="openshift-scanning-operator",resources=scanners,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=managed.openshift.io,namespace="openshift-scanning-operator",resources=scanners/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=managed.openshift.io,namespace="openshift-scanning-operator",resources=scanners/finalizers,verbs=update

// +kubebuilder:rbac:groups=security.openshift.io,namespace="openshift-scanning-operator",resources=securitycontextconstraints,verbs=use
// +kubebuilder:rbac:groups=security.openshift.io,namespace="openshift-scanning-operator",resources=securitycontextconstraints,verbs=get;create;update
// +kubebuilder:rbac:groups=security.openshift.io,namespace="openshift-scanning-operator",resourceNames=privileged,resources=securitycontextconstraints,verbs=get;create;update

// +kubebuilder:rbac:groups=core,namespace="openshift-scanning-operator",resources=pods,verbs=get;list;
// +kubebuilder:rbac:groups=core,namespace="openshift-scanning-operator",resources=events,verbs=create;watch;list;patch
// +kubebuilder:rbac:groups=core,namespace="openshift-scanning-operator",resources=secrets;services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,namespace="openshift-scanning-operator",resources=deployments;daemonsets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Scanner object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *ScannerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := r.Log.WithValues("scanner", req.NamespacedName)
	reqLogger.Info("Reconciling Scanner")

	// Fetch the Scanner instance
	instance := &managedv1alpha1.Scanner{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
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

	// Define a new DaemonSet object
	daemonSet := k8s.ScannerDaemonSet(instance)

	// Set Scanner instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, daemonSet, r.Scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this DaemonSet already exists
	found := &appsv1.DaemonSet{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: daemonSet.Name, Namespace: daemonSet.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new DaemonSet", "Daemonset.Namespace", daemonSet.Namespace, "DaemonSet.Name", daemonSet.Name)
		err = r.Client.Create(context.TODO(), daemonSet)
		if err != nil {
			return reconcile.Result{}, err
		}

		// DaemonSet created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// DaemonSet already exists - don't requeue
	reqLogger.Info("Skip reconcile: DaemonSet already exists", "DaemonSet.Namespace", found.Namespace, "DaemonSet.Name", found.Name)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScannerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&managedv1alpha1.Scanner{}).
		Owns(&managedv1alpha1.Scanner{}).
		Complete(r)
}
