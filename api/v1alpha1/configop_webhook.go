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

package v1alpha1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	validationutils "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var configoplog = logf.Log.WithName("configop-resource")

func (r *ConfigOp) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-config-configop-com-v1alpha1-configop,mutating=true,failurePolicy=fail,sideEffects=None,groups=config.configop.com,resources=configops,verbs=create;update,versions=v1alpha1,name=mconfigop.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &ConfigOp{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *ConfigOp) Default() {
	configoplog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.

	// if no namespace has been provided then give use default namespace
	if len(r.Spec.Namespaces) == 0 {
		r.Spec.Namespaces = append(r.Spec.Namespaces, "default")
	}

	//add recommended labels
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/#labels
	r.Labels = make(map[string]string)
	r.Labels["app.kubernetes.io/name"] = "configOp"
	r.Labels["app.kubernetes.io/instance"] = "configOp-instance"
	r.Labels["app.kubernetes.io/version"] = r.GroupVersionKind().Version
	r.Labels["app.kubernetes.io/component"] = "configuration"
	r.Labels["app.kubernetes.io/part-of"] = "configop-operator"
	r.Labels["app.kubernetes.io/managed-by"] = "configop-operator"
	/*
			To specify how an add-on should be managed, you can use the addonmanager.kubernetes.io/mode label. This label can have one of three values: Reconcile, EnsureExists, or Ignore.

		    Reconcile: Addon resources will be periodically reconciled with the expected state. If there are any differences, the add-on manager will recreate, reconfigure or delete the resources as needed. This is the default mode if no label is specified.
		    EnsureExists: Addon resources will be checked for existence only but will not be modified after creation. The add-on manager will create or re-create the resources when there is no instance of the resource with that name.
		    Ignore: Addon resources will be ignored. This mode is useful for add-ons that are not compatible with the add-on manager or that are managed by another controller.
	*/
	r.Labels["addonmanager.kubernetes.io/mode"] = "Reconcile"
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-config-configop-com-v1alpha1-configop,mutating=false,failurePolicy=fail,sideEffects=None,groups=config.configop.com,resources=configops,verbs=create;update,versions=v1alpha1,name=vconfigop.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &ConfigOp{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ConfigOp) ValidateCreate() error {
	configoplog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	err := r.ValidateConfigOp()

	if err != nil {
		return err
	}
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ConfigOp) ValidateUpdate(old runtime.Object) error {
	configoplog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.

	err := r.ValidateConfigOp()

	if err != nil {
		return err
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ConfigOp) ValidateDelete() error {
	configoplog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (r *ConfigOp) ValidateConfigOp() error {
	var allErrs field.ErrorList

	if err := r.validateConfigOpName(); err != nil {
		allErrs = append(allErrs, err)
	}

	if len(allErrs) != 0 {
		return apierrors.NewInvalid(
			schema.GroupKind{Group: "config", Kind: "ConfigOp"},
			r.Name, allErrs)
	}
	return nil
}

// validateConfigOpName checks the length of the custom resource configop not to be greater
// than 63 according to kubernetes naming conventions
// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#rfc-1035-label-names
func (r *ConfigOp) validateConfigOpName() *field.Error {
	if len(r.ObjectMeta.Name) > validationutils.DNS1035LabelMaxLength {
		// The configMap name length is 63 character like all Kubernetes objects
		// (which must fit in a DNS subdomain).
		return field.Invalid(field.NewPath("metadata").Child("name"), r.Name, "must be no more than 63 characters")
	}
	return nil
}
