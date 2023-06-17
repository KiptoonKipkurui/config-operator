/*
Copyright 2023.

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

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	configv1alpha1 "github.com/kiptoonkipkurui/config-operator/api/v1alpha1"
)

// ConfigOpReconciler reconciles a ConfigOp object
type ConfigOpReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=config.configop.com,resources=configops,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=config.configop.com,resources=configops/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=config.configop.com,resources=configops/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ConfigOp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *ConfigOpReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	_ = log.FromContext(ctx)
	var crd configv1alpha1.ConfigOp

	if err := r.Get(ctx, req.NamespacedName, &crd); err != nil {
		log.Log.Error(err, "unable to fetch Custom resource")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// name of our custom finalizer
	myFinalizerName := "batch.tutorial.kubebuilder.io/finalizer"
	// examine DeletionTimestamp to determine if object is under deletion
	if crd.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !controllerutil.ContainsFinalizer(&crd, myFinalizerName) {
			controllerutil.AddFinalizer(&crd, myFinalizerName)
			if err := r.Update(ctx, &crd); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(&crd, myFinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.CleanUpResources(ctx, crd); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(&crd, myFinalizerName)
			if err := r.Update(ctx, &crd); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	// loop through all namespaces ensuring the resources map exits or create otherwise
	for _, namespace := range crd.Spec.Namespaces {

		// start with config maps
		for _, v := range crd.Spec.ConfigMaps {
			var configMap v1.ConfigMap
			namespacedName := types.NamespacedName{Namespace: namespace, Name: v.Name}
			//create config map
			v.ConfigMap.Namespace = namespace
			v.ConfigMap.Name = v.Name
			if err := r.Get(ctx, namespacedName, &configMap); err != nil {

				//create the resource
				err = r.Create(ctx, &v.ConfigMap)
				if err != nil {
					log.Log.Error(err, "unable to create config map")
				}
			} else {
				// this is an update and so update the existing config map
				err = r.Update(ctx, &v.ConfigMap)

				// log any errors
				if err != nil {
					log.Log.Error(err, "unable to update config map")
				}
			}
		}

		// replicate secrets
		for _, v := range crd.Spec.Secrets {
			var configMap v1.Secret
			namespacedName := types.NamespacedName{Namespace: namespace, Name: v.Name}
			//create config map
			v.Secret.Namespace = namespace
			v.Secret.Name = v.Name
			if err := r.Get(ctx, namespacedName, &configMap); err != nil {

				// create the secret
				err = r.Create(ctx, &v.Secret)

				// log any errors
				if err != nil {
					log.Log.Error(err, "unable to create config map")
				}
			} else {

				// update the existing resource
				err = r.Update(ctx, &v.Secret)

				// log any errors
				if err != nil {
					log.Log.Error(err, "unable to create config map")
				}

			}

		}
	}

	return ctrl.Result{}, nil
}

func (r *ConfigOpReconciler) CleanUpResources(ctx context.Context, crd configv1alpha1.ConfigOp) error {
	for _, namespace := range crd.Spec.Namespaces {

		// start with config maps
		for _, v := range crd.Spec.ConfigMaps {
			var configMap v1.ConfigMap
			namespacedName := types.NamespacedName{Namespace: namespace, Name: v.Name}
			//create config map
			v.ConfigMap.Namespace = namespace
			v.ConfigMap.Name = v.Name

			if err := r.Get(ctx, namespacedName, &configMap); err == nil {

				//create the resource
				err = r.Delete(ctx, &v.ConfigMap)
				if err != nil {
					log.Log.Error(err, "unable to delete config map")
					return err
				}
			}
		}

		// replicate secrets
		for _, v := range crd.Spec.Secrets {
			var configMap v1.Secret
			namespacedName := types.NamespacedName{Namespace: namespace, Name: v.Name}
			//create config map
			v.Secret.Namespace = namespace
			v.Secret.Name = v.Name
			if err := r.Get(ctx, namespacedName, &configMap); err == nil {

				// create the secret
				err = r.Delete(ctx, &v.Secret)

				// log any errors
				if err != nil {
					log.Log.Error(err, "unable to delete config map")
					return err
				}
			}
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConfigOpReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1alpha1.ConfigOp{}).
		Complete(r)
}
