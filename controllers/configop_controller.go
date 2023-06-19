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
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

		// check if namespace exists
		ns := &v1.Namespace{}
		err := r.Client.Get(ctx, types.NamespacedName{Name: namespace}, ns)

		if err != nil {
			// create namespace since it doesnt exist
			ns.Name = namespace
			err = r.Create(ctx, ns)
			log.Log.Error(err, fmt.Sprintf("unable to create namespace %s", namespace))
		}
		// start with config maps
		for _, v := range crd.Spec.ConfigMaps {
			configMap := v1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      v.Name,
					Namespace: namespace,
				},
				Data:       v.Data,
				BinaryData: v.BinaryData,
				Immutable:  v.Immutable,
			}

			namespacedName := types.NamespacedName{Namespace: namespace, Name: v.Name}
			//create config map

			if err := r.Get(ctx, namespacedName, &configMap); err != nil {

				//create the resource
				err = r.Create(ctx, &configMap)
				if err != nil {
					log.Log.Error(err, fmt.Sprintf("unable to create config map %s", configMap.Name))
				}
			} else {
				// this is an update and so update the existing config map
				err = r.Update(ctx, &configMap)

				// log any errors
				if err != nil {
					log.Log.Error(err, "unable to update config map")
				}
			}
		}

		// replicate secrets
		for _, v := range crd.Spec.Secrets {
			secret := v1.Secret{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      v.Name,
					Namespace: namespace,
				},
				Data:       v.Data,
				StringData: v.StringData,
				Immutable:  v.Immutable,
				Type:       v.Type,
			}
			namespacedName := types.NamespacedName{Namespace: namespace, Name: v.Name}
			//create secret
			if err := r.Get(ctx, namespacedName, &secret); err != nil {

				// create the secret
				err = r.Create(ctx, &secret)

				// log any errors
				if err != nil {
					log.Log.Error(err, fmt.Sprintf("unable to create secret %s", secret.Name))
				}
			} else {

				// update the existing resource
				err = r.Update(ctx, &secret)

				// log any errors
				if err != nil {
					log.Log.Error(err, fmt.Sprintf("unable to create secret %s", secret.Name))
				}
			}
		}
	}

	return ctrl.Result{}, nil
}

// CleanUpResources is a sweeper function that runs just before a crd is deleted
// to remove any resources created by the CRD
func (r *ConfigOpReconciler) CleanUpResources(ctx context.Context, crd configv1alpha1.ConfigOp) error {
	for _, namespace := range crd.Spec.Namespaces {

		// start with config maps
		for _, v := range crd.Spec.ConfigMaps {
			configMap := v1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      v.Name,
					Namespace: namespace,
				},
				Data:       v.Data,
				BinaryData: v.BinaryData,
				Immutable:  v.Immutable,
			}
			namespacedName := types.NamespacedName{Namespace: namespace, Name: v.Name}
			//create config map

			if err := r.Get(ctx, namespacedName, &configMap); err == nil {

				//create the resource
				err = r.Delete(ctx, &configMap)
				if err != nil {
					log.Log.Error(err, fmt.Sprintf("unable to delete configMap %s", configMap.Name))
					return err
				}
			}
		}

		// replicate secrets
		for _, v := range crd.Spec.Secrets {
			secret := v1.Secret{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      v.Name,
					Namespace: namespace,
				},
				Data:       v.Data,
				StringData: v.StringData,
				Immutable:  v.Immutable,
				Type:       v.Type,
			}
			namespacedName := types.NamespacedName{Namespace: namespace, Name: v.Name}
			//delete secret

			if err := r.Get(ctx, namespacedName, &secret); err == nil {

				// create the secret
				err = r.Delete(ctx, &secret)

				// log any errors
				if err != nil {
					log.Log.Error(err, fmt.Sprintf("unable to delete secret %s", secret.Name))
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
