package example

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

// Configuration vmagent configuration definition
type Configuration struct {
	Server          ServerConfig
	PodSpecSetting  PodSpecConfig
	Log             LogConfig
	Kubelet         KubeletConfig
	ScheduledEvents ScheduledEventsConfig
	ManyModel       ManyModelConfig
	Xds             XdsConfig
	UpstreamToken   UpstreamTokenConfig
	DownstreamToken DownstreamTokenConfig
	Drain           DrainConfig
	Certificate     CertificateConfig
	Flight          FlightConfig
	DefaultEnvVars  map[string]string
}

// FlightConfig flight related config
type FlightConfig struct {
	// EnableParallelInitContainers whether to enable parallel mode for customer init container(image fetch and model mount)
	EnableParallelInitContainers bool
}

// CertificateConfig certificates related config
type CertificateConfig struct {
	EnvoyCertName string
	MdsdCertName  string
	LocalCertPath string
}

// ServerConfig vmagent server config section
type ServerConfig struct {
	// The port the server listens to
	Port                                                       int `json:",string"`
	DeploymentTimerInterval                                    int
	DeploymentHealthMaxRetries                                 int
	DeploymentTimeout                                          int
	DeploymentHealthSystemPath                                 string
	DeploymentHealthUserPath                                   string
	DeploymentDataPath                                         string
	DeploymentDefaultPodsHealthyThreshold                      int
	DeploymentCustomerPodsWithoutLivenessProbeHealthyThreshold int
	DeploymentHealthCheckIntervalInMilliSec                    int
	RecoveringStateCustomerContainerRestartLimit               int
	VMStatePersistPath                                         string
	VMUnusableFilePath                                         string
	VMRebootSignalFilePath                                     string
}

type PodRetryLimitConfig struct {
	// Max retry count for infra containers
	InfraDefault int32
	// Default max retry count for customer containers
	CustomerDefault int32
	// Limits override for specific containers, key is the container name, only for infra containers
	Limits map[string]int32
}

// PodSpecConfig vmagent podspec config
type PodSpecConfig struct {
	// CustomerPodKey customer pod key in pod annotations map
	CustomerPodKey string
	// CustomerPodValue customer pod value in pod annotations map to indicate a customer pod
	CustomerPodValue string
	// LabelMsiTokenKey is the key of MSI token of label.
	LabelMsiTokenKey string
	// LabelMsiTokenValue is the value of MSI token of label.
	LabelMsiTokenValue string
	// LabelMdsdCertKey Label key for msds cert
	LabelMdsdCertKey string
	// LabelMdsdCertValue label value for mdsd cert
	LabelMdsdCertValue string
	// LabelInitContainerAsPodKey Label key for init container pod
	LabelInitContainerAsPodKey string
	// LabelInitContainerAsPodValue Label value for init container pod
	LabelInitContainerAsPodValue string
	// LabelSkipQuotaValidation Label key for skip quota validation pod
	LabelSkipQuotaValidationKey string
	// LabelSkipQuotaValidation Label value for skip quota validation pod
	LabelSkipQuotaValidationValue string
	// The folder where built in agent podspecs locate.
	BuiltInPodSpecFolder string
	// Namespace for Infra pods
	InfraNamespace string
	// Namespace for Customer pods
	CustomerNamespace string
	// Name for VMAgent pod
	VMAgentPodName string
	// ModelMount model mount podspec config
	ModelMount PodSpecModelMountConfig
	// ImageFetcher image fetcher podspec config
	ImageFetcher PodSpecImageFetcherConfig
	// ManyModel many model podspec config
	ManyModel PodSpecManyModelConfig
	// QuotaSettings a map of container quota setting
	QuotaSettings map[string]QuotaSettingConfig
	// Pod retry limit config
	PodRetryLimit PodRetryLimitConfig
	// CuratedImageAcr list
	CuratedAcrList []string
}

// PodSpecModelMountConfig model mount pod spec config
type PodSpecModelMountConfig struct {
	// InitContainerName name of model mount init container
	InitContainerName string
	// InitContainerImage image of model mount init container
	InitContainerImage string
	// PodName name for standalone storage initializer pod
	PodName string
	// HostModelDir dir on the vm to store the downloaded model files
	HostModelDir string
}

// PodSpecImageFetcherConfig image fetcher pod spec config
type PodSpecImageFetcherConfig struct {
	// VolumeName name of image fatcher volume name
	VolumeName string
	// VolumeMountpath mount path of image fetcher
	VolumeMountpath string
	// InitContainerName name of image fetcher init container
	InitContainerName string
	// InitContainerImage image of image fetcher init container
	InitContainerImage string
	// pod name for standalone image fetcher pod
	PodName string
	// host containerd grpc address
	ContainerdAddress string
}

// PodSpecManyModelConfig many model side car container config
type PodSpecManyModelConfig struct {
	// ManyModelSideCarImageUrl side car image url
	SideCarImageUrl string
	SideCarPort     int32
}

// QuotaSettingConfig podspec quota setting config
type QuotaSettingConfig struct {
	// CpuResourceInPercent container cpu resource limit in percentage
	CpuResourceInPercent float64
	// MemoryResourceInGb container memory resource limit
	MemoryResourceInGb float64
	// NeedResourceLimit if container need resource limit
	NeedResourceLimit bool
}

// LogConfig vmagent log config
type LogConfig struct {
	// The server log file path
	LogPath string
	// The server log level(debug", "info", "warn",
	// "error", "dpanic", "panic", and "fatal")
	LogLevel string
	// Enable writing to standard output
	EnableStdOut bool
}

// KubeletConfig vmagent kubelet config
type KubeletConfig struct {
	// ServiceAddr kubelet endpoint
	ServiceAddr string
	// PodsAPI api path for getting pod info
	PodsAPI string
	// HealthzAPI api path for getting kubelet health info
	HealthzAPI string
	// RunningPodsAPI api path for getting running pods
	RunningPodsAPI string
	// StatsSummaryAPI api path for getting states summary
	StatsSummaryAPI string
	// ManifestsFolderPath path for kubelet pod manifests folder
	ManifestsFolderPath string
}

// manyModelConfig many model config section
type ManyModelConfig struct {
	// ManyModelUpdateTimerInterval update timer interval
	ManyModelUpdateTimerInterval int
}

// ScheduledEventsConfig vmagent scheduled events config section
type ScheduledEventsConfig struct {
	// MetadataUrl Url of scheduled events Metadata Service
	MetadataUrl string
	// SchduleEventsUpdateTimerInterval update timer interval
	SchduleEventsUpdateTimerInterval int
	// FreezeAckMinWaitTimeInSec is the minimum wait second for acknowledging Freeze event in Pause State
	FreezeAckMinWaitTimeInSec int
}

// XdsConfig vmagent envoy xds server config section
type XdsConfig struct {
	// Port is xds server listening port
	Port int `json:",string"`
	// EnvoyMeshPort is the port of mesh listener
	EnvoyMeshPort int `json:",string"`
	// EnvoyNodeID is the envoy node_id of mesh first layer
	EnvoyNodeID string
	// EnvoyLayer2NodeID is the envoy node_id of mesh second layer
	EnvoyLayer2NodeID string
	// DefaultEnvoySDSTlsSecretName is the SDS config name defined in Envoy configuration file
	DefaultEnvoySDSTlsSecretName string
	// DefaultEnvoySDSHttpValidationSecretName is the SDS config name defined in Envoy configuration file
	DefaultEnvoySDSHttpValidationSecretName string
	// DefaultEnvoySDSDiagValidationSecretName is the SDS config name defined in Envoy configuration file
	DefaultEnvoySDSDiagValidationSecretName string
	// EnvoyHealthCheckPath is the health check path defined in Envoy configuration file
	EnvoyHealthCheckPath string
	// EnvoyHealthCheckClusterName is the CDS config name defined in Envoy configuration file
	EnvoyHealthCheckClusterName string
	// EnvoyHealthCheckEndpointName is the EDS config name defined in Envoy configuration file
	EnvoyHealthCheckEndpointName string
	// EnvoyHealthCheckPort is the port of health check listener
	EnvoyHealthCheckPort int
	// EnvoyFDMeshClusterName
	EnvoyFDMeshClusterName string
	// EnvoyFDMeshEndpointName is the EDS config name defined in Envoy configuration file
	EnvoyFDMeshEndpointName string
	// EnvoyMeshRouteConfigName is the RDS config name defined in Envoy configuration file
	EnvoyMeshRouteConfigName string
	// HealthCheckTimeoutMilliseconds defines health check route timeout in milliseconds
	HealthCheckTimeoutMilliseconds int
	// RouteTimeoutMilliseconds defines the route request timeout in milliseconds
	RouteTimeoutMilliseconds int
	// UpstreamClusterConnectTimeoutMilliseconds defines the connection timeout to an upstream cluster
	UpstreamClusterConnectTimeoutMilliseconds int
	// EnvoyMeshEndpointUpdateTimerInterval time interval config for update envoy mesh endpoints
	EnvoyMeshEndpointUpdateTimerInterval int
	// CACertificateFileName is the CA certificates file used by envoy mTLS validation
	CACertificateFileName string
	// CertificateChainSortingEnabled indicates whether to sort certificate chain on exporting public key from PEM or PFX
	CertificateChainSortingEnabled bool
	// RequestHeadersToRemove contains a list of header names that need to remove before sending to custom container
	RequestHeadersToRemove []string
	// RequestHeadersToPassBy contains a list of header names that need to pass by sending to custom container, add back to response.
	RequestHeadersToPassBy []string

	// EnvoyManyModelRouteConfigName used for triton manymodel flight.
	EnvoyManyModelRouteConfigName string
	// EnvoyManyModelDirectRouteHeaderName used for triton manymodel flight.
	EnvoyManyModelDirectRouteHeaderName string
	// EnvoyManyModelDirectRouteHeaderName used for triton manymodel flight.
	EnvoyManyModelDirectRouteHeaderValue string
	// EnvoyManyModelRequestPathPattern used for triton manymodel flight.
	EnvoyManyModelRequestPathPattern string

	// ModelConcurrencyThresholdToDisableLogs
	ModelConcurrencyThresholdToDisableLogs int
}

type UpstreamTokenConfig struct {
	// UpstreamMsiTokenEndpoint is endpoint of MSI of token server.
	UpstreamMsiTokenEndpoint string
	// UpstreamAcrTokenEndpoint is endpoint of ACR of token server.
	UpstreamAcrTokenEndpoint string
	// UpstreamAcrTokenEndpoint is endpoint of Blob key of token server
	UpstreamBlobKeyEndpoint string
	// upstreamCertEndpoint is endpoint of cert server
	UpstreamSecretArtifactEndpoint string
}

type DownstreamTokenConfig struct {
	// DownstreamMsiTokenEndpoint is endpoint of MSI of VMAgent.
	DownstreamMsiTokenEndpoint string
	// DownstreamAcrTokenEndpoint is endpoint of ACR of VMAgent.
	DownstreamAcrTokenEndpoint string
	// DownstreamBlobKeyEndpoint is endpoint of BLOB of VMAgent.
	DownstreamBlobKeyEndpoint        string
	DownstreamSecretArtifactEndpoint string
}

type DrainConfig struct {
	EnvoyDrainEndpoint          string
	EnvoyStatsEndpoint          string
	SLBMarkUnhealthyMaxDuration int
}

func InitConfig(configPath string) (*Configuration, error) {
	var configuration Configuration
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		fmt.Println("Error getting config file absolute path: ", err)
		return nil, err
	}
	data, err := ioutil.ReadFile(absPath)
	if err != nil {
		fmt.Println("Error reading config file: ", err)
		return nil, err
	}
	err = json.Unmarshal(data, &configuration)
	if err != nil {
		fmt.Println("Error Unmarshal config file", "data", string(data), "Error", err.Error())
		return nil, err
	}
	return &configuration, nil
}

func GetDefaultConfig() *Configuration {
	return &Configuration{
		Server: ServerConfig{
			Port:                                  8911,
			DeploymentTimerInterval:               5,
			DeploymentHealthMaxRetries:            300,
			DeploymentTimeout:                     -1,
			DeploymentHealthSystemPath:            "/var/log/mir/system.err",
			DeploymentHealthUserPath:              "/var/log/mir/user.err",
			DeploymentDataPath:                    "/etc/deployment/deployment.data",
			DeploymentDefaultPodsHealthyThreshold: 3,
			DeploymentCustomerPodsWithoutLivenessProbeHealthyThreshold: 16,
			DeploymentHealthCheckIntervalInMilliSec:                    500,
			RecoveringStateCustomerContainerRestartLimit:               5,
			VMUnusableFilePath:     "/var/log/unusable.txt",
			VMRebootSignalFilePath: "/var/log/reboot.txt",
		},
		PodSpecSetting: PodSpecConfig{
			InfraNamespace:                "mirinfra",
			CustomerNamespace:             "miruser",
			BuiltInPodSpecFolder:          "/etc/podspecs",
			CustomerPodKey:                "userPod",
			CustomerPodValue:              "true",
			LabelMsiTokenKey:              "msitoken",
			LabelMsiTokenValue:            "true",
			LabelMdsdCertKey:              "mdsdcert",
			LabelMdsdCertValue:            "true",
			LabelInitContainerAsPodKey:    "initpod",
			LabelInitContainerAsPodValue:  "true",
			LabelSkipQuotaValidationKey:   "skipquotavalidation",
			LabelSkipQuotaValidationValue: "true",
			ModelMount: PodSpecModelMountConfig{
				InitContainerName:  "model-mount",
				InitContainerImage: "model-mount-sample/image",
				PodName:            "model-mount",
			},
			ImageFetcher: PodSpecImageFetcherConfig{
				VolumeName:         "image-fetcher-volume",
				VolumeMountpath:    "image-fetcher/volume/mountpath",
				InitContainerName:  "image-fetcher",
				InitContainerImage: "image-fetcher-sample/image",
				PodName:            "image-fetcher",
				ContainerdAddress:  "/run/containerd/containerd.sock",
			},
			QuotaSettings: map[string]QuotaSettingConfig{
				"envoy": {
					CpuResourceInPercent: 10,
					MemoryResourceInGb:   0.5,
					NeedResourceLimit:    false,
				},
			},
			PodRetryLimit: PodRetryLimitConfig{
				InfraDefault:    4,
				CustomerDefault: 3,
			},
			ManyModel: PodSpecManyModelConfig{
				SideCarImageUrl: "mirmasteracr.azurecr.io/many-models-sidecar-proxy:v1",
				SideCarPort:     8111,
			},
		},
		Log: LogConfig{
			LogPath:      "/var/log/mir-vmagent.log",
			LogLevel:     "DEBUG",
			EnableStdOut: true,
		},
		Kubelet: KubeletConfig{
			ServiceAddr:         "https://localhost:10250",
			PodsAPI:             "/pods",
			HealthzAPI:          "/healthz",
			RunningPodsAPI:      "/runningpods",
			StatsSummaryAPI:     "/stats/summary",
			ManifestsFolderPath: "/etc/kubelet/manifests",
		},
		ScheduledEvents: ScheduledEventsConfig{
			MetadataUrl:                      "http://169.254.169.254/metadata/scheduledevents?api-version=2019-08-01",
			SchduleEventsUpdateTimerInterval: 20,
			FreezeAckMinWaitTimeInSec:        30,
		},
		ManyModel: ManyModelConfig{
			ManyModelUpdateTimerInterval: 5,
		},
		Xds: XdsConfig{
			Port:                                      12345,
			EnvoyMeshPort:                             10001,
			EnvoyNodeID:                               "node_01",
			EnvoyLayer2NodeID:                         "node_02",
			DefaultEnvoySDSTlsSecretName:              "server-cert",
			DefaultEnvoySDSHttpValidationSecretName:   "validation-context-http",
			DefaultEnvoySDSDiagValidationSecretName:   "validation-context-diag",
			EnvoyHealthCheckPath:                      "/healthz",
			EnvoyHealthCheckClusterName:               "fd_health_cluster",
			EnvoyHealthCheckEndpointName:              "fd_health_endpoint",
			EnvoyHealthCheckPort:                      10002,
			EnvoyFDMeshClusterName:                    "fd_mesh_cluster",
			EnvoyFDMeshEndpointName:                   "fd_mesh_endpoint",
			EnvoyMeshRouteConfigName:                  "ingress-route-config-mesh",
			HealthCheckTimeoutMilliseconds:            250,
			RouteTimeoutMilliseconds:                  600000,
			UpstreamClusterConnectTimeoutMilliseconds: 250,
			EnvoyMeshEndpointUpdateTimerInterval:      10,
			CACertificateFileName:                     "/etc/ssl/certs/ca-certificates.crt",
			CertificateChainSortingEnabled:            false,
			RequestHeadersToRemove:                    []string{"x-envoy-expected-rq-timeout-ms"},
			RequestHeadersToPassBy:                    []string{"x-request-id"},
			EnvoyManyModelRouteConfigName:             "ingress-route-config-manymodel",
			EnvoyManyModelDirectRouteHeaderName:       "x-mir-route-all",
			EnvoyManyModelDirectRouteHeaderValue:      "true",
			EnvoyManyModelRequestPathPattern:          "/v2/models/%s/",
			ModelConcurrencyThresholdToDisableLogs:    32,
		},
		UpstreamToken: UpstreamTokenConfig{
			UpstreamMsiTokenEndpoint: "http://127.0.0.1:8081/token/MSI",
			UpstreamAcrTokenEndpoint: "http://127.0.0.1:8081/token/ACR",
			UpstreamBlobKeyEndpoint:  "http://127.0.0.1:8081/secret/storage",
		},
		DownstreamToken: DownstreamTokenConfig{
			DownstreamMsiTokenEndpoint: "http://127.0.0.1:8080/v1/token/msi",
			DownstreamAcrTokenEndpoint: "http://127.0.0.1:8080/v1/token/acr",
			DownstreamBlobKeyEndpoint:  "http://127.0.0.1:8080/v1/token/blob",
		},
		Drain: DrainConfig{
			EnvoyDrainEndpoint:          "http://127.0.0.1:9901/drain_listeners?graceful",
			EnvoyStatsEndpoint:          "http://127.0.0.1:9901/stats",
			SLBMarkUnhealthyMaxDuration: 15,
		},
		Flight: FlightConfig{
			EnableParallelInitContainers: true,
		},
		Certificate: CertificateConfig{
			EnvoyCertName: "MIR_ENVOY_CERT_NAME",
			MdsdCertName:  "MIR_MDSD_CERT_NAME",
			LocalCertPath: "",
		},
		DefaultEnvVars: map[string]string{
			"MODEL_REQUEST_TIMEOUT":      "600",
			"MIR_MTLS_DISABLE":           "false",
			"MIR_SCHEDULED_EVENT_ENABLE": "false",
		},
	}
}
