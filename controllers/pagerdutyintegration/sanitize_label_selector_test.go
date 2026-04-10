package pagerdutyintegration

import (
	"testing"

	"github.com/openshift/pagerduty-operator/config"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func TestSanitizeLabelSelector(t *testing.T) {
	tests := []struct {
		name        string
		input       *metav1.LabelSelector
		matchingSet labels.Set
		nonMatchSet labels.Set
		expected    *metav1.LabelSelector
	}{
		{
			name: "adds managed selector when missing",
			input: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"team": "sre",
				},
			},
			matchingSet: labels.Set{
				config.ClusterDeploymentManagedLabel: "true",
				"team":                               "sre",
			},
			nonMatchSet: labels.Set{
				config.ClusterDeploymentManagedLabel: "false",
				"team":                               "sre",
			},
			expected: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"team": "sre",
				},
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      config.ClusterDeploymentManagedLabel,
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{"true"},
					},
				},
			},
		},
		{
			name: "replaces conflicting managed selectors",
			input: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					config.ClusterDeploymentManagedLabel: "false",
					"team":                               "sre",
				},
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      config.ClusterDeploymentManagedLabel,
						Operator: metav1.LabelSelectorOpNotIn,
						Values:   []string{"true"},
					},
					{
						Key:      "region",
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{"us-east-1"},
					},
				},
			},
			matchingSet: labels.Set{
				config.ClusterDeploymentManagedLabel: "true",
				"team":                               "sre",
				"region":                             "us-east-1",
			},
			nonMatchSet: labels.Set{
				config.ClusterDeploymentManagedLabel: "false",
				"team":                               "sre",
				"region":                             "us-east-1",
			},
			expected: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"team": "sre",
				},
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "region",
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{"us-east-1"},
					},
					{
						Key:      config.ClusterDeploymentManagedLabel,
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{"true"},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			original := test.input.DeepCopy()

			sanitized := sanitizeLabelSelector(test.input)
			assert.Equal(t, test.expected, sanitized)
			assert.Equal(t, original, test.input)

			selector, err := clusterDeploymentSelector(test.input)
			assert.NoError(t, err)
			assert.True(t, selector.Matches(test.matchingSet))
			assert.False(t, selector.Matches(test.nonMatchSet))
		})
	}
}

func TestClusterDeploymentSelectorReturnsErrors(t *testing.T) {
	_, err := clusterDeploymentSelector(&metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      "region",
				Operator: metav1.LabelSelectorOpIn,
			},
		},
	})

	assert.Error(t, err)
}
