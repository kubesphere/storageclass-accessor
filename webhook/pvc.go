package webhook

import (
	"context"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type reqInfo struct {
	resource         string
	name             string
	namespace        string
	operator         string
	storageClassName string
}

var reviewResponse = &admissionv1.AdmissionResponse{
	Allowed: true,
	Result:  &metav1.Status{},
}

func admitPVC(ar admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	klog.Info("admitting pvc")

	if !(ar.Request.Operation == admissionv1.Delete || ar.Request.Operation == admissionv1.Create) {
		return reviewResponse
	}

	raw := ar.Request.Object.Raw

	var newPVC *corev1.PersistentVolumeClaim

	switch ar.Request.Operation {
	case admissionv1.Create:
		deserializer := codecs.UniversalDeserializer()
		pvc := &corev1.PersistentVolumeClaim{}
		obj, _, err := deserializer.Decode(raw, nil, pvc)
		if err != nil {
			klog.Error(err)
			return toV1AdmissionResponse(err)
		}
		var ok bool
		newPVC, ok = obj.(*corev1.PersistentVolumeClaim)
		if !ok {
			klog.Error("obj can't exchange to pvc object")
			return toV1AdmissionResponse(err)
		}
	case admissionv1.Delete:
		pvcInfo := types.NamespacedName{
			Namespace: ar.Request.Namespace,
			Name:      ar.Request.Name,
		}
		cli, err := client.New(config.GetConfigOrDie(), client.Options{})
		if err != nil {
			return toV1AdmissionResponse(err)
		}
		targetPVC := &corev1.PersistentVolumeClaim{}
		err = cli.Get(context.Background(), pvcInfo, targetPVC)
		if err != nil {
			klog.Error("get target Delete PVC from client failed, err:", err)
			return toV1AdmissionResponse(err)
		}
		newPVC = targetPVC
	}

	reqPVC := reqInfo{
		resource:         "persistentVolumeClaim",
		name:             newPVC.Name,
		namespace:        newPVC.Namespace,
		operator:         string(ar.Request.Operation),
		storageClassName: *newPVC.Spec.StorageClassName,
	}
	return decidePVCV1(reqPVC)
}

func decidePVCV1(pvc reqInfo) *admissionv1.AdmissionResponse {

	accessors, err := getAccessor(pvc.storageClassName)

	if err != nil {
		klog.Error("get accessor failed, err:", err)
		return toV1AdmissionResponse(err)
	} else if len(accessors) == 0 {
		klog.Info("Not Found accessor for the storageClass:", pvc.storageClassName)
		return reviewResponse
	}

	for _, accessor := range accessors {
		if err = validateNameSpace(pvc, accessor); err != nil {
			return toV1AdmissionResponse(err)
		}
	}
	return reviewResponse
}
