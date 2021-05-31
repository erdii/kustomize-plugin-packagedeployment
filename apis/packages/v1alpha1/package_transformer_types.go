package v1alpha1

import (
	packagesv1alpha1 "github.com/thetechnick/package-operator/apis/packages/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	SchemeBuilder.Register(&PackageTransformer{})
}

// PackageTransformer
// +kubebuilder:object:root=true
type PackageTransformer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	ReadinessProbes   []packagesv1alpha1.PackageProbe `json:"readinessProbes"`
}
