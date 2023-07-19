// Code generated by mdatagen. DO NOT EDIT.

package metadata

import "go.opentelemetry.io/collector/confmap"

// MetricConfig provides common config for a particular metric.
type MetricConfig struct {
	Enabled bool `mapstructure:"enabled"`

	enabledSetByUser bool
}

func (ms *MetricConfig) Unmarshal(parser *confmap.Conf) error {
	if parser == nil {
		return nil
	}
	err := parser.Unmarshal(ms, confmap.WithErrorUnused())
	if err != nil {
		return err
	}
	ms.enabledSetByUser = parser.IsSet("enabled")
	return nil
}

// MetricsConfig provides config for k8s/replicationcontroller metrics.
type MetricsConfig struct {
	K8sReplicationControllerAvailable MetricConfig `mapstructure:"k8s.replication_controller.available"`
	K8sReplicationControllerDesired   MetricConfig `mapstructure:"k8s.replication_controller.desired"`
}

func DefaultMetricsConfig() MetricsConfig {
	return MetricsConfig{
		K8sReplicationControllerAvailable: MetricConfig{
			Enabled: true,
		},
		K8sReplicationControllerDesired: MetricConfig{
			Enabled: true,
		},
	}
}

// ResourceAttributeConfig provides common config for a particular resource attribute.
type ResourceAttributeConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

// ResourceAttributesConfig provides config for k8s/replicationcontroller resource attributes.
type ResourceAttributesConfig struct {
	K8sNamespaceName             ResourceAttributeConfig `mapstructure:"k8s.namespace.name"`
	K8sReplicationcontrollerName ResourceAttributeConfig `mapstructure:"k8s.replicationcontroller.name"`
	K8sReplicationcontrollerUID  ResourceAttributeConfig `mapstructure:"k8s.replicationcontroller.uid"`
	OpencensusResourcetype       ResourceAttributeConfig `mapstructure:"opencensus.resourcetype"`
}

func DefaultResourceAttributesConfig() ResourceAttributesConfig {
	return ResourceAttributesConfig{
		K8sNamespaceName: ResourceAttributeConfig{
			Enabled: true,
		},
		K8sReplicationcontrollerName: ResourceAttributeConfig{
			Enabled: true,
		},
		K8sReplicationcontrollerUID: ResourceAttributeConfig{
			Enabled: true,
		},
		OpencensusResourcetype: ResourceAttributeConfig{
			Enabled: true,
		},
	}
}

// MetricsBuilderConfig is a configuration for k8s/replicationcontroller metrics builder.
type MetricsBuilderConfig struct {
	Metrics            MetricsConfig            `mapstructure:"metrics"`
	ResourceAttributes ResourceAttributesConfig `mapstructure:"resource_attributes"`
}

func DefaultMetricsBuilderConfig() MetricsBuilderConfig {
	return MetricsBuilderConfig{
		Metrics:            DefaultMetricsConfig(),
		ResourceAttributes: DefaultResourceAttributesConfig(),
	}
}
