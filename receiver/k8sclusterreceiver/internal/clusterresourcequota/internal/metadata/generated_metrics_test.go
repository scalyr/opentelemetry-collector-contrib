// Code generated by mdatagen. DO NOT EDIT.

package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

type testConfigCollection int

const (
	testSetDefault testConfigCollection = iota
	testSetAll
	testSetNone
)

func TestMetricsBuilder(t *testing.T) {
	tests := []struct {
		name      string
		configSet testConfigCollection
	}{
		{
			name:      "default",
			configSet: testSetDefault,
		},
		{
			name:      "all_set",
			configSet: testSetAll,
		},
		{
			name:      "none_set",
			configSet: testSetNone,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			start := pcommon.Timestamp(1_000_000_000)
			ts := pcommon.Timestamp(1_000_001_000)
			observedZapCore, observedLogs := observer.New(zap.WarnLevel)
			settings := receivertest.NewNopCreateSettings()
			settings.Logger = zap.New(observedZapCore)
			mb := NewMetricsBuilder(loadMetricsBuilderConfig(t, test.name), settings, WithStartTime(start))

			expectedWarnings := 0
			assert.Equal(t, expectedWarnings, observedLogs.Len())

			defaultMetricsCount := 0
			allMetricsCount := 0

			defaultMetricsCount++
			allMetricsCount++
			mb.RecordOpenshiftAppliedclusterquotaLimitDataPoint(ts, 1, "k8s.namespace.name-val", "resource-val")

			defaultMetricsCount++
			allMetricsCount++
			mb.RecordOpenshiftAppliedclusterquotaUsedDataPoint(ts, 1, "k8s.namespace.name-val", "resource-val")

			defaultMetricsCount++
			allMetricsCount++
			mb.RecordOpenshiftClusterquotaLimitDataPoint(ts, 1, "resource-val")

			defaultMetricsCount++
			allMetricsCount++
			mb.RecordOpenshiftClusterquotaUsedDataPoint(ts, 1, "resource-val")

			metrics := mb.Emit(WithOpencensusResourcetype("opencensus.resourcetype-val"), WithOpenshiftClusterquotaName("openshift.clusterquota.name-val"), WithOpenshiftClusterquotaUID("openshift.clusterquota.uid-val"))

			if test.configSet == testSetNone {
				assert.Equal(t, 0, metrics.ResourceMetrics().Len())
				return
			}

			assert.Equal(t, 1, metrics.ResourceMetrics().Len())
			rm := metrics.ResourceMetrics().At(0)
			attrCount := 0
			enabledAttrCount := 0
			attrVal, ok := rm.Resource().Attributes().Get("opencensus.resourcetype")
			attrCount++
			assert.Equal(t, mb.resourceAttributesConfig.OpencensusResourcetype.Enabled, ok)
			if mb.resourceAttributesConfig.OpencensusResourcetype.Enabled {
				enabledAttrCount++
				assert.EqualValues(t, "opencensus.resourcetype-val", attrVal.Str())
			}
			attrVal, ok = rm.Resource().Attributes().Get("openshift.clusterquota.name")
			attrCount++
			assert.Equal(t, mb.resourceAttributesConfig.OpenshiftClusterquotaName.Enabled, ok)
			if mb.resourceAttributesConfig.OpenshiftClusterquotaName.Enabled {
				enabledAttrCount++
				assert.EqualValues(t, "openshift.clusterquota.name-val", attrVal.Str())
			}
			attrVal, ok = rm.Resource().Attributes().Get("openshift.clusterquota.uid")
			attrCount++
			assert.Equal(t, mb.resourceAttributesConfig.OpenshiftClusterquotaUID.Enabled, ok)
			if mb.resourceAttributesConfig.OpenshiftClusterquotaUID.Enabled {
				enabledAttrCount++
				assert.EqualValues(t, "openshift.clusterquota.uid-val", attrVal.Str())
			}
			assert.Equal(t, enabledAttrCount, rm.Resource().Attributes().Len())
			assert.Equal(t, attrCount, 3)

			assert.Equal(t, 1, rm.ScopeMetrics().Len())
			ms := rm.ScopeMetrics().At(0).Metrics()
			if test.configSet == testSetDefault {
				assert.Equal(t, defaultMetricsCount, ms.Len())
			}
			if test.configSet == testSetAll {
				assert.Equal(t, allMetricsCount, ms.Len())
			}
			validatedMetrics := make(map[string]bool)
			for i := 0; i < ms.Len(); i++ {
				switch ms.At(i).Name() {
				case "openshift.appliedclusterquota.limit":
					assert.False(t, validatedMetrics["openshift.appliedclusterquota.limit"], "Found a duplicate in the metrics slice: openshift.appliedclusterquota.limit")
					validatedMetrics["openshift.appliedclusterquota.limit"] = true
					assert.Equal(t, pmetric.MetricTypeGauge, ms.At(i).Type())
					assert.Equal(t, 1, ms.At(i).Gauge().DataPoints().Len())
					assert.Equal(t, "The upper limit for a particular resource in a specific namespace.", ms.At(i).Description())
					assert.Equal(t, "1", ms.At(i).Unit())
					dp := ms.At(i).Gauge().DataPoints().At(0)
					assert.Equal(t, start, dp.StartTimestamp())
					assert.Equal(t, ts, dp.Timestamp())
					assert.Equal(t, pmetric.NumberDataPointValueTypeInt, dp.ValueType())
					assert.Equal(t, int64(1), dp.IntValue())
					attrVal, ok := dp.Attributes().Get("k8s.namespace.name")
					assert.True(t, ok)
					assert.EqualValues(t, "k8s.namespace.name-val", attrVal.Str())
					attrVal, ok = dp.Attributes().Get("resource")
					assert.True(t, ok)
					assert.EqualValues(t, "resource-val", attrVal.Str())
				case "openshift.appliedclusterquota.used":
					assert.False(t, validatedMetrics["openshift.appliedclusterquota.used"], "Found a duplicate in the metrics slice: openshift.appliedclusterquota.used")
					validatedMetrics["openshift.appliedclusterquota.used"] = true
					assert.Equal(t, pmetric.MetricTypeGauge, ms.At(i).Type())
					assert.Equal(t, 1, ms.At(i).Gauge().DataPoints().Len())
					assert.Equal(t, "The usage for a particular resource in a specific namespace.", ms.At(i).Description())
					assert.Equal(t, "1", ms.At(i).Unit())
					dp := ms.At(i).Gauge().DataPoints().At(0)
					assert.Equal(t, start, dp.StartTimestamp())
					assert.Equal(t, ts, dp.Timestamp())
					assert.Equal(t, pmetric.NumberDataPointValueTypeInt, dp.ValueType())
					assert.Equal(t, int64(1), dp.IntValue())
					attrVal, ok := dp.Attributes().Get("k8s.namespace.name")
					assert.True(t, ok)
					assert.EqualValues(t, "k8s.namespace.name-val", attrVal.Str())
					attrVal, ok = dp.Attributes().Get("resource")
					assert.True(t, ok)
					assert.EqualValues(t, "resource-val", attrVal.Str())
				case "openshift.clusterquota.limit":
					assert.False(t, validatedMetrics["openshift.clusterquota.limit"], "Found a duplicate in the metrics slice: openshift.clusterquota.limit")
					validatedMetrics["openshift.clusterquota.limit"] = true
					assert.Equal(t, pmetric.MetricTypeGauge, ms.At(i).Type())
					assert.Equal(t, 1, ms.At(i).Gauge().DataPoints().Len())
					assert.Equal(t, "The configured upper limit for a particular resource.", ms.At(i).Description())
					assert.Equal(t, "1", ms.At(i).Unit())
					dp := ms.At(i).Gauge().DataPoints().At(0)
					assert.Equal(t, start, dp.StartTimestamp())
					assert.Equal(t, ts, dp.Timestamp())
					assert.Equal(t, pmetric.NumberDataPointValueTypeInt, dp.ValueType())
					assert.Equal(t, int64(1), dp.IntValue())
					attrVal, ok := dp.Attributes().Get("resource")
					assert.True(t, ok)
					assert.EqualValues(t, "resource-val", attrVal.Str())
				case "openshift.clusterquota.used":
					assert.False(t, validatedMetrics["openshift.clusterquota.used"], "Found a duplicate in the metrics slice: openshift.clusterquota.used")
					validatedMetrics["openshift.clusterquota.used"] = true
					assert.Equal(t, pmetric.MetricTypeGauge, ms.At(i).Type())
					assert.Equal(t, 1, ms.At(i).Gauge().DataPoints().Len())
					assert.Equal(t, "The usage for a particular resource with a configured limit.", ms.At(i).Description())
					assert.Equal(t, "1", ms.At(i).Unit())
					dp := ms.At(i).Gauge().DataPoints().At(0)
					assert.Equal(t, start, dp.StartTimestamp())
					assert.Equal(t, ts, dp.Timestamp())
					assert.Equal(t, pmetric.NumberDataPointValueTypeInt, dp.ValueType())
					assert.Equal(t, int64(1), dp.IntValue())
					attrVal, ok := dp.Attributes().Get("resource")
					assert.True(t, ok)
					assert.EqualValues(t, "resource-val", attrVal.Str())
				}
			}
		})
	}
}
