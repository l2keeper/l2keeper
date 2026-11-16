// Package v1alpha1 contains the L2Keeper API types.
package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is the group and version for L2Keeper resources.
	GroupVersion = schema.GroupVersion{Group: "l2keeper.github.io", Version: "v1alpha1"}

	// SchemeBuilder registers the L2Keeper API types with a scheme.
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the L2Keeper API types to a scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

func init() {
	SchemeBuilder.Register(&VIPBinding{}, &VIPBindingList{}, &VIPPool{}, &VIPPoolList{})
}
