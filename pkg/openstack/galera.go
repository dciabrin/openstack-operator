package openstack

import (
	"context"
	"fmt"

	"github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	mariadbv1 "github.com/openstack-k8s-operators/mariadb-operator/api/v1beta1"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	corev1beta1 "github.com/openstack-k8s-operators/openstack-operator/apis/core/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// ReconcileGalera -
func ReconcileGalera(ctx context.Context, instance *corev1beta1.OpenStackControlPlane, helper *helper.Helper) (ctrl.Result, error) {
	if !instance.Spec.Galera.Enabled {
		return ctrl.Result{}, nil
	}

	galera := &mariadbv1.Galera{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "openstack", //FIXME
			Namespace: instance.Namespace,
		},
	}

	helper.GetLogger().Info("Reconciling Galera", "Galera.Namespace", instance.Namespace, "Galera.Name", "openstack")
	op, err := controllerutil.CreateOrPatch(ctx, helper.GetClient(), galera, func() error {
		instance.Spec.Galera.Template.DeepCopyInto(&galera.Spec)
		if galera.Spec.Secret == "" {
			galera.Spec.Secret = instance.Spec.Secret
		}
		if galera.Spec.StorageClass == "" {
			galera.Spec.StorageClass = instance.Spec.StorageClass
		}
		err := controllerutil.SetControllerReference(helper.GetBeforeObject(), galera, helper.GetScheme())
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		instance.Status.Conditions.Set(condition.FalseCondition(
			corev1beta1.OpenStackControlPlaneMariaDBReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			corev1beta1.OpenStackControlPlaneMariaDBReadyErrorMessage,
			err.Error()))
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		helper.GetLogger().Info(fmt.Sprintf("Galera %s - %s", galera.Name, op))
	}

	if galera.IsReady() {
		instance.Status.Conditions.MarkTrue(corev1beta1.OpenStackControlPlaneMariaDBReadyCondition, corev1beta1.OpenStackControlPlaneMariaDBReadyMessage)
	} else {
		instance.Status.Conditions.Set(condition.FalseCondition(
			corev1beta1.OpenStackControlPlaneMariaDBReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			corev1beta1.OpenStackControlPlaneMariaDBReadyRunningMessage))
	}

	return ctrl.Result{}, nil

}
