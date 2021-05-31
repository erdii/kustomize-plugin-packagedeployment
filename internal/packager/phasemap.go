package packager

import packagesv1alpha1 "github.com/thetechnick/package-operator/apis/packages/v1alpha1"

type phaseMap map[string]*packagesv1alpha1.PackagePhase

func (pm phaseMap) slice() (slice []packagesv1alpha1.PackagePhase) {
	for _, phase := range pm {
		slice = append(slice, *phase)
	}
	return slice
}
