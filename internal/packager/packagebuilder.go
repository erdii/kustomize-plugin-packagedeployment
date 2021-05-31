package packager

import (
	"fmt"

	packagesv1alpha1 "github.com/thetechnick/package-operator/apis/packages/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

const (
	phaseAnnotation      = "packages.thetechnick.ninja/phase"
	packageSelectorLabel = "packagedeployment"
)

type PackageBuilder struct {
	name, namespace string
	phases          phaseMap
	probes          []packagesv1alpha1.PackageProbe
}

func NewPackageBuilder(name, namespace string, probes []packagesv1alpha1.PackageProbe) *PackageBuilder {
	return &PackageBuilder{
		name:      name,
		namespace: namespace,
		phases:    make(phaseMap),
		probes:    probes,
	}
}

func (p *PackageBuilder) Build() *packagesv1alpha1.PackageDeployment {
	packageDeployment := &packagesv1alpha1.PackageDeployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      p.name,
			Namespace: p.namespace,
		},
		Spec: packagesv1alpha1.PackageDeploymentSpec{
			Selector: v1.LabelSelector{
				MatchLabels: map[string]string{
					packageSelectorLabel: fmt.Sprintf(
						"%s_%s", p.namespace, p.name),
				},
			},
			Template: packagesv1alpha1.PackageSetTemplate{
				Metadata: v1.ObjectMeta{
					Labels: map[string]string{
						packageSelectorLabel: fmt.Sprintf(
							"%s_%s", p.namespace, p.name),
					},
				},
				Spec: packagesv1alpha1.PackageSetTemplateSpec{
					Phases:          p.phases.slice(),
					ReadinessProbes: p.probes,
				},
			},
		},
	}

	packageDeployment.Kind = "PackageDeployment"
	packageDeployment.APIVersion = "packages.thetechnick.ninja/v1alpha1"

	return packageDeployment
}

func (p *PackageBuilder) YAML() ([]byte, error) {
	return yaml.Marshal(p.Build())
}

func (p *PackageBuilder) ensurePhase(name string) *packagesv1alpha1.PackagePhase {
	if phase, ok := p.phases[name]; ok {
		return phase
	}

	phase := &packagesv1alpha1.PackagePhase{
		Name: name,
	}
	p.phases[name] = phase
	return phase
}

func (p *PackageBuilder) AddManifest(bytes []byte) error {
	obj := &unstructured.Unstructured{}
	if err := yaml.Unmarshal(bytes, obj); err != nil {
		return fmt.Errorf("could not parse yaml: %w", err)
	}
	return p.AddObject(obj)
}

func (p *PackageBuilder) AddObject(obj *unstructured.Unstructured) error {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	phaseName, ok := annotations[phaseAnnotation]
	if !ok {
		phaseName = "deploy"
		annotations[phaseAnnotation] = phaseName
		obj.SetAnnotations(annotations)
	}

	objBytes, err := yaml.Marshal(obj)
	if err != nil {
		return fmt.Errorf("could not marshal object: %w", err)
	}
	raw := &runtime.RawExtension{}
	if err := yaml.Unmarshal(objBytes, raw); err != nil {
		return fmt.Errorf("could not remarshal object: %w", err)
	}
	packageObject := packagesv1alpha1.PackageObject{
		Object: *raw,
	}

	phase := p.ensurePhase(phaseName)
	phase.Objects = append(phase.Objects, packageObject)
	return nil
}
