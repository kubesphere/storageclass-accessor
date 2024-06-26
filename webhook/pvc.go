package webhook

import (
	"context"
	"fmt"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type ReqInfo struct {
	resource         string
	name             string
	namespace        string
	operator         string
	storageClassName string
}

func NewPVCReqInfo(name string, namespace string, storageClassName string, operation admissionv1.Operation) ReqInfo {
	return ReqInfo{
		resource:         "persistentVolumeClaim",
		name:             name,
		namespace:        namespace,
		operator:         string(operation),
		storageClassName: storageClassName,
	}
}

var reviewResponse = &admissionv1.AdmissionResponse{
	Allowed: true,
	Result:  &metav1.Status{},
}

type PVCAdmitter interface {
	AdmitPVC(ar admissionv1.AdmissionReview) *admissionv1.AdmissionResponse
	DecidePVC(ctx context.Context, pvc ReqInfo) *admissionv1.AdmissionResponse
}

type Admitter struct {
	client client.Client
}

var _ PVCAdmitter = (*Admitter)(nil)

func NewAdmitter() (*Admitter, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	cli, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, err
	}
	a := &Admitter{
		client: cli,
	}
	return a, nil
}

func NewAdmitterWithClient(client client.Client) PVCAdmitter {
	return &Admitter{
		client: client,
	}
}

func (a *Admitter) serverPVCRequest(w http.ResponseWriter, r *http.Request) {
	server(w, r, newDelegateToV1AdmitHandler(a.AdmitPVC))
}

func (a *Admitter) AdmitPVC(ar admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	if ar.Request.Operation != admissionv1.Create {
		return reviewResponse
	}

	raw := ar.Request.Object.Raw
	deserializer := codecs.UniversalDeserializer()
	pvc := &corev1.PersistentVolumeClaim{}
	obj, _, err := deserializer.Decode(raw, nil, pvc)
	if err != nil {
		klog.ErrorS(err, "failed to decode raw object")
		return toV1AdmissionResponse(err)
	}

	newPVC, ok := obj.(*corev1.PersistentVolumeClaim)
	if !ok {
		err = fmt.Errorf("obj can't exchange to pvc object")
		klog.ErrorS(err, "failed to exchange to pvc object")
		return toV1AdmissionResponse(err)
	}

	if newPVC.Spec.StorageClassName == nil {
		klog.Infof("pvc %s in namespace %s has no StorageClassName, allow it.", newPVC.Name, newPVC.Namespace)
		return reviewResponse
	}

	reqPVC := NewPVCReqInfo(newPVC.Name, newPVC.Namespace, *newPVC.Spec.StorageClassName, ar.Request.Operation)

	klog.Infof("request pvc: %v", reqPVC)
	return a.DecidePVC(context.Background(), reqPVC)
}

func (a *Admitter) DecidePVC(ctx context.Context, pvc ReqInfo) *admissionv1.AdmissionResponse {
	accessors, err := a.getAccessors(ctx, pvc.storageClassName)

	if err != nil {
		klog.ErrorS(err, "get accessor failed")
		return toV1AdmissionResponse(err)
	} else if len(accessors) == 0 {
		klog.Infof("Not Found accessor for the storageClass: %s", pvc.storageClassName)
		return reviewResponse
	}

	for _, accessor := range accessors {
		klog.Infof("starting validating accessor: %s", accessor.Name)
		if err = a.validateNameSpace(ctx, pvc, &accessor); err != nil {
			return toV1AdmissionResponse(err)
		}
		if err = a.validateWorkSpace(ctx, pvc, &accessor); err != nil {
			return toV1AdmissionResponse(err)
		}
	}
	return reviewResponse
}
