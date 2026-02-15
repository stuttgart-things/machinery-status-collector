package informer

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ExtractClaimStatus navigates .status.conditions on an unstructured Crossplane
// claim and returns the message from the first condition whose type is "Ready".
// If no status or conditions are found, it returns a descriptive fallback.
func ExtractClaimStatus(obj *unstructured.Unstructured) (string, error) {
	conditions, found, err := unstructured.NestedSlice(obj.Object, "status", "conditions")
	if err != nil {
		return "", fmt.Errorf("read status.conditions: %w", err)
	}
	if !found || len(conditions) == 0 {
		return "no status conditions available", nil
	}

	for _, c := range conditions {
		cond, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		condType, _, _ := unstructured.NestedString(cond, "type")
		if condType != "Ready" {
			continue
		}
		msg, _, _ := unstructured.NestedString(cond, "message")
		status, _, _ := unstructured.NestedString(cond, "status")
		if msg != "" {
			return msg, nil
		}
		return fmt.Sprintf("Ready=%s", status), nil
	}

	return "no Ready condition found", nil
}
