package common

import (
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cronhpav1beta1 "github.com/wosai/elastic-env-operator/api/cronhpa/v1beta1"
)

const CronHPAIndexByRef = "byRefWorkload"

var CronHPARefIndexFunc = func(obj interface{}) ([]string, error) {
	cronHPA := obj.(*cronhpav1beta1.CronHorizontalPodAutoscaler)

	return []string{fmt.Sprintf("%s.%s.%s", cronHPA.Spec.ScaleTargetRef.ApiVersion, cronHPA.Spec.ScaleTargetRef.Kind, cronHPA.Spec.ScaleTargetRef.Name)}, nil
}

func BuildRefKey(obj client.Object) string {
	return fmt.Sprintf("%s/%s.%s.%s", obj.GetObjectKind().GroupVersionKind().Group,
		obj.GetObjectKind().GroupVersionKind().Version,
		obj.GetObjectKind().GroupVersionKind().Kind,
		obj.GetName())
}
