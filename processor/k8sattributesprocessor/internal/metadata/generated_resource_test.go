// Code generated by mdatagen. DO NOT EDIT.

package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResourceBuilder(t *testing.T) {
	for _, test := range []string{"default", "all_set", "none_set"} {
		t.Run(test, func(t *testing.T) {
			cfg := loadResourceAttributesConfig(t, test)
			rb := NewResourceBuilder(cfg)
			rb.SetContainerID("container.id-val")
			rb.SetContainerImageName("container.image.name-val")
			rb.SetContainerImageTag("container.image.tag-val")
			rb.SetK8sContainerName("k8s.container.name-val")
			rb.SetK8sCronjobName("k8s.cronjob.name-val")
			rb.SetK8sDaemonsetName("k8s.daemonset.name-val")
			rb.SetK8sDaemonsetUID("k8s.daemonset.uid-val")
			rb.SetK8sDeploymentName("k8s.deployment.name-val")
			rb.SetK8sDeploymentUID("k8s.deployment.uid-val")
			rb.SetK8sJobName("k8s.job.name-val")
			rb.SetK8sJobUID("k8s.job.uid-val")
			rb.SetK8sNamespaceName("k8s.namespace.name-val")
			rb.SetK8sNodeName("k8s.node.name-val")
			rb.SetK8sPodHostname("k8s.pod.hostname-val")
			rb.SetK8sPodName("k8s.pod.name-val")
			rb.SetK8sPodStartTime("k8s.pod.start_time-val")
			rb.SetK8sPodUID("k8s.pod.uid-val")
			rb.SetK8sReplicasetName("k8s.replicaset.name-val")
			rb.SetK8sReplicasetUID("k8s.replicaset.uid-val")
			rb.SetK8sStatefulsetName("k8s.statefulset.name-val")
			rb.SetK8sStatefulsetUID("k8s.statefulset.uid-val")

			res := rb.Emit()
			assert.Equal(t, 0, rb.Emit().Attributes().Len()) // Second call should return 0

			switch test {
			case "default":
				assert.Equal(t, 8, res.Attributes().Len())
			case "all_set":
				assert.Equal(t, 21, res.Attributes().Len())
			case "none_set":
				assert.Equal(t, 0, res.Attributes().Len())
				return
			default:
				assert.Failf(t, "unexpected test case: %s", test)
			}

			val, ok := res.Attributes().Get("container.id")
			assert.Equal(t, test == "all_set", ok)
			if ok {
				assert.EqualValues(t, "container.id-val", val.Str())
			}
			val, ok = res.Attributes().Get("container.image.name")
			assert.True(t, ok)
			if ok {
				assert.EqualValues(t, "container.image.name-val", val.Str())
			}
			val, ok = res.Attributes().Get("container.image.tag")
			assert.True(t, ok)
			if ok {
				assert.EqualValues(t, "container.image.tag-val", val.Str())
			}
			val, ok = res.Attributes().Get("k8s.container.name")
			assert.Equal(t, test == "all_set", ok)
			if ok {
				assert.EqualValues(t, "k8s.container.name-val", val.Str())
			}
			val, ok = res.Attributes().Get("k8s.cronjob.name")
			assert.Equal(t, test == "all_set", ok)
			if ok {
				assert.EqualValues(t, "k8s.cronjob.name-val", val.Str())
			}
			val, ok = res.Attributes().Get("k8s.daemonset.name")
			assert.Equal(t, test == "all_set", ok)
			if ok {
				assert.EqualValues(t, "k8s.daemonset.name-val", val.Str())
			}
			val, ok = res.Attributes().Get("k8s.daemonset.uid")
			assert.Equal(t, test == "all_set", ok)
			if ok {
				assert.EqualValues(t, "k8s.daemonset.uid-val", val.Str())
			}
			val, ok = res.Attributes().Get("k8s.deployment.name")
			assert.True(t, ok)
			if ok {
				assert.EqualValues(t, "k8s.deployment.name-val", val.Str())
			}
			val, ok = res.Attributes().Get("k8s.deployment.uid")
			assert.Equal(t, test == "all_set", ok)
			if ok {
				assert.EqualValues(t, "k8s.deployment.uid-val", val.Str())
			}
			val, ok = res.Attributes().Get("k8s.job.name")
			assert.Equal(t, test == "all_set", ok)
			if ok {
				assert.EqualValues(t, "k8s.job.name-val", val.Str())
			}
			val, ok = res.Attributes().Get("k8s.job.uid")
			assert.Equal(t, test == "all_set", ok)
			if ok {
				assert.EqualValues(t, "k8s.job.uid-val", val.Str())
			}
			val, ok = res.Attributes().Get("k8s.namespace.name")
			assert.True(t, ok)
			if ok {
				assert.EqualValues(t, "k8s.namespace.name-val", val.Str())
			}
			val, ok = res.Attributes().Get("k8s.node.name")
			assert.True(t, ok)
			if ok {
				assert.EqualValues(t, "k8s.node.name-val", val.Str())
			}
			val, ok = res.Attributes().Get("k8s.pod.hostname")
			assert.Equal(t, test == "all_set", ok)
			if ok {
				assert.EqualValues(t, "k8s.pod.hostname-val", val.Str())
			}
			val, ok = res.Attributes().Get("k8s.pod.name")
			assert.True(t, ok)
			if ok {
				assert.EqualValues(t, "k8s.pod.name-val", val.Str())
			}
			val, ok = res.Attributes().Get("k8s.pod.start_time")
			assert.True(t, ok)
			if ok {
				assert.EqualValues(t, "k8s.pod.start_time-val", val.Str())
			}
			val, ok = res.Attributes().Get("k8s.pod.uid")
			assert.True(t, ok)
			if ok {
				assert.EqualValues(t, "k8s.pod.uid-val", val.Str())
			}
			val, ok = res.Attributes().Get("k8s.replicaset.name")
			assert.Equal(t, test == "all_set", ok)
			if ok {
				assert.EqualValues(t, "k8s.replicaset.name-val", val.Str())
			}
			val, ok = res.Attributes().Get("k8s.replicaset.uid")
			assert.Equal(t, test == "all_set", ok)
			if ok {
				assert.EqualValues(t, "k8s.replicaset.uid-val", val.Str())
			}
			val, ok = res.Attributes().Get("k8s.statefulset.name")
			assert.Equal(t, test == "all_set", ok)
			if ok {
				assert.EqualValues(t, "k8s.statefulset.name-val", val.Str())
			}
			val, ok = res.Attributes().Get("k8s.statefulset.uid")
			assert.Equal(t, test == "all_set", ok)
			if ok {
				assert.EqualValues(t, "k8s.statefulset.uid-val", val.Str())
			}
		})
	}
}
