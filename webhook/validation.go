package webhook

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"storageclass-accessor/client/apis/accessor/v1alpha1"
)

func validateNameSpace(reqResource reqInfo, accessor *v1alpha1.Accessor) error {
	klog.Info("start validate namespace")
	//accessor, err := getAccessor()
	ns, err := getNameSpace(reqResource.namespace)
	if err != nil {
		klog.Error(err)
		return err
	}
	var fieldPass, labelPass bool
	fieldPass = matchField(ns, accessor.Spec.NameSpaceSelector.FieldSelector)
	labelPass = matchLabel(ns.Labels, accessor.Spec.NameSpaceSelector.LabelSelector)
	if fieldPass && labelPass {
		return nil
	}

	klog.Error(fmt.Sprintf("%s %s does not allowed %s in the namespace: %s", reqResource.resource, reqResource.name, reqResource.operator, reqResource.namespace))
	return fmt.Errorf("The storageClass: %s does not allowed %s %s %s in the namespace: %s ", reqResource.storageClassName, reqResource.operator, reqResource.resource, reqResource.name, reqResource.namespace)
}

func getNameSpace(nameSpaceName string) (*corev1.Namespace, error) {
	nsClient, err := client.New(config.GetConfigOrDie(), client.Options{})
	if err != nil {
		return nil, err
	}
	ns := &corev1.Namespace{}
	err = nsClient.Get(context.Background(), types.NamespacedName{Namespace: "", Name: nameSpaceName}, ns)
	if err != nil {
		klog.Error("client get namespace failed, err:", err)
		return nil, err
	}
	return ns, nil
}

func getAccessor(storageClassName string) ([]*v1alpha1.Accessor, error) {
	// get config
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	var cli client.Client
	opts := client.Options{}
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)
	opts.Scheme = scheme
	cli, err = client.New(cfg, opts)
	if err != nil {
		return nil, err
	}
	accessorList := &v1alpha1.AccessorList{}

	var listOpt []client.ListOption
	err = cli.List(context.Background(), accessorList, listOpt...)
	if err != nil {
		// TODO If not found , pass or not?
		return nil, err
	}
	list := make([]*v1alpha1.Accessor, 0)
	for _, accessor := range accessorList.Items {
		if accessor.Spec.StorageClassName == storageClassName {
			list = append(list, &accessor)
		}
	}
	return list, nil
}
