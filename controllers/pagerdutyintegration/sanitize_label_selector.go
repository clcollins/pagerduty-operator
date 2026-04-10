package pagerdutyintegration

import (
	"github.com/openshift/pagerduty-operator/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func clusterDeploymentSelector(labelSelector *metav1.LabelSelector) (labels.Selector, error) {
	return metav1.LabelSelectorAsSelector(sanitizeLabelSelector(labelSelector))
}

func sanitizeLabelSelector(labelSelector *metav1.LabelSelector) *metav1.LabelSelector {
	if labelSelector == nil {
		labelSelector = &metav1.LabelSelector{}
	}

	sanitizedSelector := labelSelector.DeepCopy()

	delete(sanitizedSelector.MatchLabels, config.ClusterDeploymentManagedLabel)
	if len(sanitizedSelector.MatchLabels) == 0 {
		sanitizedSelector.MatchLabels = nil
	}

	matchExpressions := make([]metav1.LabelSelectorRequirement, 0, len(sanitizedSelector.MatchExpressions)+1)
	for _, matchExpression := range sanitizedSelector.MatchExpressions {
		if matchExpression.Key == config.ClusterDeploymentManagedLabel {
			continue
		}
		matchExpressions = append(matchExpressions, matchExpression)
	}

	matchExpressions = append(matchExpressions, metav1.LabelSelectorRequirement{
		Key:      config.ClusterDeploymentManagedLabel,
		Operator: metav1.LabelSelectorOpIn,
		Values:   []string{"true"},
	})
	sanitizedSelector.MatchExpressions = matchExpressions

	return sanitizedSelector
}
