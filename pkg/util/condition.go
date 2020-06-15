package util

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubevirt-image-service/pkg/apis/hypercloud/v1alpha1"
	"time"
)

// GetConditionByType returns a condition with conditionType. The condition returned is a clone, so changing it does not change the original conditions[]
func GetConditionByType(conditions []v1alpha1.Condition, conditionType string) (found bool, condition v1alpha1.Condition) {
	for _, c := range conditions {
		if c.Type == conditionType {
			return true, c
		}
	}
	return false, v1alpha1.Condition{}
}

// SetConditionByType sets condition to conditions. If there is a matching condition.Type, update it, if not, add it. It Returns the new slice
func SetConditionByType(conditions []v1alpha1.Condition, conditionType string, status corev1.ConditionStatus, reason, message string) []v1alpha1.Condition {
	for i := range conditions {
		if conditions[i].Type != conditionType {
			continue
		}
		conditions[i].Status = status
		conditions[i].Reason = reason
		conditions[i].Message = message
		conditions[i].LastTransitionTime = metav1.NewTime(time.Now())
		return conditions
	}
	return append(conditions, v1alpha1.Condition{
		Type:               conditionType,
		Status:             status,
		LastTransitionTime: metav1.NewTime(time.Now()),
		Reason:             reason,
		Message:            message,
	})
}
