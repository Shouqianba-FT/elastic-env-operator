package handler

import (
	"context"

	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	qav1alpha1 "github.com/wosai/elastic-env-operator/api/v1alpha1"
	"github.com/wosai/elastic-env-operator/domain/common"
	"github.com/wosai/elastic-env-operator/domain/entity"
)

type sqbDeploymentHandler struct {
	req ctrl.Request
	ctx context.Context
}

func NewSqbDeploymentHandler(req ctrl.Request, ctx context.Context, indexer cache.Indexer) SQBReconciler {
	return &sqbDeploymentHandler{
		req: req,
		ctx: context.WithValue(ctx, common.ContextKeyCronHPAIndexer, indexer),
	}
}

func (h *sqbDeploymentHandler) GetInstance() (runtimeObj, error) {
	in := &qav1alpha1.SQBDeployment{}
	err := k8sclient.Get(h.ctx, h.req.NamespacedName, in)
	return in, err
}

// 初始化逻辑
func (h *sqbDeploymentHandler) IsInitialized(obj runtimeObj) (bool, error) {
	in := obj.(*qav1alpha1.SQBDeployment)
	if in.Annotations[entity.InitializeAnnotationKey] == "true" {
		return true, nil
	}

	// 补充默认值
	sqbapplication := &qav1alpha1.SQBApplication{}
	if err := k8sclient.Get(h.ctx, client.ObjectKey{Namespace: in.Namespace, Name: in.Spec.Selector.App},
		sqbapplication); err != nil {
		return false, err
	}
	if len(in.Annotations) == 0 {
		in.Annotations = make(map[string]string)
	}
	in.Annotations[entity.InitializeAnnotationKey] = "true"
	if IsIstioInject(sqbapplication) {
		in.Annotations[entity.IstioInjectAnnotationKey] = "true"
	} else {
		in.Annotations[entity.IstioInjectAnnotationKey] = "false"
	}
	return false, CreateOrUpdate(h.ctx, in)
}

// 正常处理逻辑
func (h *sqbDeploymentHandler) Operate(obj runtimeObj) error {
	in := obj.(*qav1alpha1.SQBDeployment)
	deleted, err := IsDeleted(in)
	if err != nil {
		return err
	}

	handlers := []SQBHandler{
		NewPVCHandler(in, h.ctx),
		NewDeploymentHandler(in, h.ctx),
		//NewGrayServiceHandler(in, h.ctx),
		//NewGrayVMServiceScrapeHandler(in, h.ctx),
		NewSqbdeploymentIngressHandler(in, h.ctx),
	}

	for _, handler := range handlers {
		if err = handler.Handle(); err != nil {
			return err
		}
	}

	if deleted {
		return Delete(h.ctx, in)
	} else if in.Status.ErrorInfo != "" {
		in.Status.ErrorInfo = ""
		return UpdateStatus(h.ctx, in)
	}
	return nil
}

// 处理失败后逻辑
func (h *sqbDeploymentHandler) ReconcileFail(obj runtimeObj, err error) {
	in := obj.(*qav1alpha1.SQBDeployment)
	in.Status.ErrorInfo = err.Error()
	_ = UpdateStatus(h.ctx, in)
}

func HasPublicEntry(sqbdeployment *qav1alpha1.SQBDeployment) bool {
	publicEntry, ok := sqbdeployment.Annotations[entity.PublicEntryAnnotationKey]
	if ok {
		return publicEntry == "true"
	}
	return false
}

// 返回special virtualservice的入口对应在哪个ingress上
func SpecialVirtualServiceIngress(sqbdeployment *qav1alpha1.SQBDeployment) string {
	if specialVirtualServiceIngress, ok := sqbdeployment.Annotations[entity.SpecialVirtualServiceIngress]; ok {
		return specialVirtualServiceIngress
	}
	return entity.ConfigMapData.SpecialVirtualServiceIngress()
}
