package blockuseofbindmounts

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/multi-tenancy/benchmarks/kubectl-mtb/bundle/box"
	"sigs.k8s.io/multi-tenancy/benchmarks/kubectl-mtb/pkg/benchmark"
	"sigs.k8s.io/multi-tenancy/benchmarks/kubectl-mtb/test"
	"sigs.k8s.io/multi-tenancy/benchmarks/kubectl-mtb/test/utils"
	"sigs.k8s.io/multi-tenancy/benchmarks/kubectl-mtb/test/utils/log"
	podutil "sigs.k8s.io/multi-tenancy/benchmarks/kubectl-mtb/test/utils/resources/pod"
	"sigs.k8s.io/multi-tenancy/benchmarks/kubectl-mtb/types"
)

var b = &benchmark.Benchmark{

	PreRun: func(options types.RunOptions) error {

		resource := utils.GroupResource{
			APIGroup: "",
			APIResource: metav1.APIResource{
				Name: "pods",
			},
		}

		access, msg, err := utils.RunAccessCheck(options.TClient, options.TenantNamespace, resource, "create")
		if err != nil {
			log.Logging.Debug(err.Error())
			return err
		}
		if !access {
			return fmt.Errorf(msg)
		}

		return nil
	},

	Run: func(options types.RunOptions) error {

		// Host path
		inlineVolumeSources := []*v1.VolumeSource{
			{
				HostPath: &v1.HostPathVolumeSource{
					Path: "/tmp/busybox",
				},
			},
		}

		podSpec := &podutil.PodSpec{NS: options.TenantNamespace, InlineVolumeSources: inlineVolumeSources}
		err := podSpec.SetDefaults()
		if err != nil {
			log.Logging.Debug(err.Error())
			return err
		}

		// Try to create a pod as tenant-admin impersonation
		pod := podSpec.MakeSecPod()
		_, err = options.TClient.CoreV1().Pods(options.TenantNamespace).Create(context.TODO(), pod, metav1.CreateOptions{DryRun: []string{metav1.DryRunAll}})
		if err == nil {
			return fmt.Errorf("Tenant must be unable to create pod with host-path volume")
		}
		log.Logging.Debug("Test passed: ", err.Error())
		return nil
	},
}

func init() {
	// Get the []byte representation of a file, or an error if it doesn't exist:
	err := b.ReadConfig(box.Get("block_use_of_bind_mounts/config.yaml"))
	if err != nil {
		log.Logging.Error(err.Error())
	}

	test.BenchmarkSuite.Add(b)
}
