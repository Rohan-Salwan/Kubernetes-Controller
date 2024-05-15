/*
Copyright 2024 Rohan-Salwan.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	apiv1alpha1 "github.com/clusterscan/kubernetes-controller.git/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ClusterScanReconciler reconciles a ClusterScan object
type ClusterScanReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=api.core.scan.io,resources=clusterscans,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=api.core.scan.io,resources=clusterscans/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=api.core.scan.io,resources=clusterscans/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ClusterScan object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *ClusterScanReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	clusterscans := &apiv1alpha1.ClusterScan{}
	err := r.Get(ctx, client.ObjectKey{
		Namespace: req.Namespace,
		Name:      "clusterscan-sample",
	}, clusterscans)
	if err != nil {
		return ctrl.Result{}, err
	}

	for index, job := range clusterscans.Spec.Jobs {
		if (clusterscans.Spec.Results[index].Status == "pending") && (clusterscans.Spec.Results[index].Name == job.Name) {
			cronjob := CreateCronJobTemplate(job.Name, job.Parameters["image"], job.Parameters["cmd"], job.Schedule)
			if err := r.Create(context.Background(), cronjob); err != nil {
				return ctrl.Result{}, err
			}
			clusterscans.Spec.Results[index].Status = "created"
		}

	}

	if err := r.Update(ctx, clusterscans); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func CreateCronJobTemplate(name string, image string, cmd string, schedule string) *batchv1.CronJob {
	return &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: batchv1.CronJobSpec{
			Schedule: schedule,
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:    name,
									Image:   image,
									Command: []string{"/bin/sh", "-c", "date; echo Hello " + cmd},
								},
							},
							RestartPolicy: "OnFailure",
						},
					},
				},
			},
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterScanReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1alpha1.ClusterScan{}).
		Complete(r)
}
