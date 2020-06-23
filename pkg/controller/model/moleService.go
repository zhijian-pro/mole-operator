package model

import (
	molev1 "dtstack.com/dtstack/mole-operator/pkg/apis/mole/v1"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"strings"
)

func getServiceLabels(cr *molev1.Mole, name string) map[string]string {
	var labels = map[string]string{}

	labels["pid"] = strconv.Itoa(cr.Spec.Product.Pid)
	labels["deploy_uuid"] = cr.Spec.Product.DeployUUid
	labels["cluster_id"] = strconv.Itoa(cr.Spec.Product.ClusterId)
	labels["product_name"] = cr.Spec.Product.ProductName
	labels["product_version"] = cr.Spec.Product.ProductVersion
	labels["parent_product_name"] = cr.Spec.Product.ParentProductName
	labels["service_name"] = name
	labels["service_version"] = cr.Spec.Product.Service[name].Version

	return labels
}

func getServiceAnnotations(cr *molev1.Mole, existing map[string]string, name string) map[string]string {
	if cr.Spec.Product.Service[name].Instance.Service == nil || cr.Spec.Product.Service[name].Instance.Service.Annotations == nil {
		return existing
	}
	return MergeAnnotations(cr.Spec.Product.Service[name].Instance.Service.Annotations, existing)
}

func getServiceType(cr *molev1.Mole, name string) v1.ServiceType {
	if cr.Spec.Product.Service[name].Instance.Service == nil {
		return v1.ServiceTypeClusterIP
	}
	if cr.Spec.Product.Service[name].Instance.Service.Type == "" {
		return v1.ServiceTypeClusterIP
	}
	return cr.Spec.Product.Service[name].Instance.Service.Type
}

//func GetMolePort(cr *molev1.Mole, name string, index int) int {
//	return cr.Spec.Product.Service[name].Instance.Deployment.Ports[index]
//}

func getServicePorts(cr *molev1.Mole, name string) []v1.ServicePort {
	//portName := BuildPortName(name, MoleHttpPortName)
	defaultPorts := make([]v1.ServicePort, 0)
	for index, port := range cr.Spec.Product.Service[name].Instance.Deployment.Ports {
		defaultPorts = append(defaultPorts, v1.ServicePort{
			Name:       BuildPortName(name, index),
			Protocol:   "TCP",
			Port:       int32(port),
			TargetPort: intstr.FromString(BuildPortName(name, index)),
		})
	}
	if cr.Spec.Product.Service[name].Instance.Service == nil {
		return defaultPorts
	}
	return defaultPorts
}

func MoleService(cr *molev1.Mole, name string) *v1.Service {
	productVersion := strings.ReplaceAll(cr.Spec.Product.ProductVersion, ".", "")
	return &v1.Service{
		ObjectMeta: v12.ObjectMeta{
			Name:        BuildResourceName(MoleServiceName, cr.Spec.Product.ParentProductName, cr.Spec.Product.ProductName, productVersion, name),
			Namespace:   cr.Namespace,
			Labels:      getServiceLabels(cr, name),
			Annotations: getServiceAnnotations(cr, nil, name),
		},
		Spec: v1.ServiceSpec{
			Ports: getServicePorts(cr, name),
			Selector: map[string]string{
				"app": BuildResourceLabel(cr.Spec.Product.ParentProductName, cr.Spec.Product.ProductName, name),
			},
			ClusterIP: "",
			Type:      getServiceType(cr, name),
		},
	}
}

func MoleServiceReconciled(cr *molev1.Mole, currentState *v1.Service, name string) *v1.Service {
	reconciled := currentState.DeepCopy()
	reconciled.Labels = getServiceLabels(cr, name)
	reconciled.Annotations = getServiceAnnotations(cr, currentState.Annotations, name)
	reconciled.Spec.Ports = getServicePorts(cr, name)
	reconciled.Spec.Type = getServiceType(cr, name)
	return reconciled
}

func MoleServiceSelector(cr *molev1.Mole, name string) client.ObjectKey {
	productVersion := strings.ReplaceAll(cr.Spec.Product.ProductVersion, ".", "")
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      BuildResourceName(MoleServiceName, cr.Spec.Product.ParentProductName, cr.Spec.Product.ProductName, productVersion, name),
	}
}
