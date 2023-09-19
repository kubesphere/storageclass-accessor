module github.com/kubesphere/storageclass-accessor

go 1.16

require (
	github.com/fsnotify/fsnotify v1.4.9
	github.com/spf13/cobra v1.2.1
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/klog/v2 v2.9.0
	k8s.io/utils v0.0.0-20210802155522-efc7438f0176
	kubesphere.io/api v0.0.0-20210917114432-19cb9aacd65f
	sigs.k8s.io/controller-runtime v0.10.0
)

replace (
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.1
	k8s.io/client-go => k8s.io/client-go v0.21.2
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20210305001622-591a79e4bda7
)
