package webhook

import (
	corev1 "k8s.io/api/core/v1"
	"storageclass-accessor/client/apis/accessor/v1alpha1"
)

func matchLabel(info map[string]string, expressions []v1alpha1.MatchExpressions) bool {
	if len(expressions) == 0 {
		return true
	}

	for _, rule := range expressions {
		rulePass := true
		for _, item := range rule.MatchExpressions {
			switch item.Operator {
			case v1alpha1.In:
				rulePass = rulePass && inList(info[item.Key], item.Values)
			case v1alpha1.NotIn:
				rulePass = rulePass && !inList(info[item.Key], item.Values)
			}
		}
		if rulePass {
			return rulePass
		}
	}
	return false
}

func matchField(ns *corev1.Namespace, expressions []v1alpha1.FieldExpressions) bool {
	//If not set limit, default pass
	if len(expressions) == 0 {
		return true
	}

	for _, rule := range expressions {
		rulePass := true
		for _, item := range rule.FieldExpressions {
			var val string
			switch item.Field {
			case v1alpha1.Name:
				val = ns.Name
			case v1alpha1.Phase:
				val = string(ns.Status.Phase)
			}
			switch item.Operator {
			case v1alpha1.In:
				rulePass = rulePass && inList(val, item.Values)
			case v1alpha1.NotIn:
				rulePass = rulePass && !inList(val, item.Values)
			}
		}
		if rulePass {
			return rulePass
		}
	}
	return false
}

func inList(val string, list []string) bool {
	for _, elements := range list {
		if val == elements {
			return true
		}
	}
	return false
}
