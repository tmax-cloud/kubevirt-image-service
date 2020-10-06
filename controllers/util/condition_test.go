package util

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tmax-cloud/kubevirt-image-service/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

var _ = Describe("GetConditionByType", func() {
	Context("if conditions has no matching conditionType", func() {
		found, _ := GetConditionByType([]v1alpha1.Condition{
			{
				Type:               "type1",
				Status:             corev1.ConditionTrue,
				ObservedGeneration: 32,
				LastTransitionTime: v1.Time{
					Time: time.Time{},
				},
				Reason:  "TestReason",
				Message: "Message",
			},
		}, "NoMatchingType")

		It("Should return found=false", func() {
			Expect(found).Should(BeFalse())
		})
	})

	Context("if conditions has matching conditionType", func() {
		expectedCondition := v1alpha1.Condition{
			Type:               "MatchingType",
			Status:             corev1.ConditionFalse,
			ObservedGeneration: 17,
			LastTransitionTime: v1.Time{
				Time: time.Time{},
			},
			Reason:  "TestReason",
			Message: "Message",
		}
		found, cond := GetConditionByType([]v1alpha1.Condition{
			expectedCondition,
			{
				Type:               "type1",
				Status:             corev1.ConditionTrue,
				ObservedGeneration: 32,
				LastTransitionTime: v1.Time{
					Time: time.Time{},
				},
				Reason:  "TestReason",
				Message: "Message",
			},
			{
				Type:               "type2",
				Status:             corev1.ConditionFalse,
				ObservedGeneration: 33,
				LastTransitionTime: v1.Time{
					Time: time.Time{},
				},
				Reason:  "TestReason2",
				Message: "Message2",
			},
		}, expectedCondition.Type)

		It("Should return found=true", func() {
			Expect(found).Should(BeTrue())
		})

		It("Should return that condition", func() {
			Expect(cond).Should(Equal(expectedCondition))
		})
	})
})

var _ = Describe("SetConditionByType", func() {
	Context("if conditions has no matching conditionType", func() {
		conditions := []v1alpha1.Condition{
			{
				Type:               "type1",
				Status:             corev1.ConditionTrue,
				ObservedGeneration: 32,
				LastTransitionTime: v1.Time{
					Time: time.Time{},
				},
				Reason:  "TestReason",
				Message: "Message",
			},
			{
				Type:               "type2",
				Status:             corev1.ConditionFalse,
				ObservedGeneration: 33,
				LastTransitionTime: v1.Time{
					Time: time.Time{},
				},
				Reason:  "TestReason2",
				Message: "Message2",
			},
		}
		conditionsAfterSet := SetConditionByType(conditions, "type3", corev1.ConditionFalse, "TestReason3", "Message3")

		It("should append it", func() {
			Expect(conditionsAfterSet[2].Type).Should(Equal("type3"))
			Expect(conditionsAfterSet[2].Status).Should(Equal(corev1.ConditionFalse))
			Expect(conditionsAfterSet[2].Reason).Should(Equal("TestReason3"))
			Expect(conditionsAfterSet[2].Message).Should(Equal("Message3"))
		})

		It("should not change or delete other conditions", func() {
			Expect(len(conditionsAfterSet)).Should(Equal(3))
			Expect(conditionsAfterSet[0]).Should(Equal(conditions[0]))
			Expect(conditionsAfterSet[1]).Should(Equal(conditions[1]))
		})
	})

	Context("if conditions has matching conditionType", func() {
		conditions := []v1alpha1.Condition{
			{
				Type:               "type1",
				Status:             corev1.ConditionTrue,
				ObservedGeneration: 32,
				LastTransitionTime: v1.Time{
					Time: time.Time{},
				},
				Reason:  "TestReason",
				Message: "Message",
			},
			{
				Type:               "type2",
				Status:             corev1.ConditionFalse,
				ObservedGeneration: 33,
				LastTransitionTime: v1.Time{
					Time: time.Time{},
				},
				Reason:  "TestReason2",
				Message: "Message2",
			},
		}
		conditionsAfterSet := SetConditionByType(conditions, "type2", corev1.ConditionTrue, "TestReasonNew", "MessageNew")

		It("should update it", func() {
			Expect(conditionsAfterSet[1].Type).Should(Equal("type2"))
			Expect(conditionsAfterSet[1].Status).Should(Equal(corev1.ConditionTrue))
			Expect(conditionsAfterSet[1].Reason).Should(Equal("TestReasonNew"))
			Expect(conditionsAfterSet[1].Message).Should(Equal("MessageNew"))
		})

		It("should not change or delete other conditions", func() {
			Expect(len(conditionsAfterSet)).Should(Equal(2))
			Expect(conditionsAfterSet[0]).Should(Equal(conditions[0]))
		})
	})
})
