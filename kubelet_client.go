package example

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	//"goms.io/azureml/mir/mir-vmagent/pkg/common"
	//"goms.io/azureml/mir/mir-vmagent/pkg/config"
	//"goms.io/azureml/mir/mir-vmagent/pkg/log"

	//"goms.io/azureml/mir/mir-vmagent/pkg/metrics"

	"go.uber.org/zap"

	v1 "k8s.io/api/core/v1"
	//statsapi "k8s.io/kubernetes/pkg/kubelet/apis/stats/v1alpha1"
)

const (
	// INSTANCE_ID key for hostname in os env
	INSTANCE_ID = "INSTANCE_ID"
)

// PodInfoGetter is a interface that help to get kubelet pods info
type PodInfoGetter interface {
	GetPodsInfo() ([]v1.Pod, error)
}

// RunningPodsGetter interface for getting all running pods
type RunningPodsGetter interface {
	GetRunningPods() (*v1.PodList, error)
}

// HealthStatusGetter is a interface that help to get kubelet healthz status
type HealthStatusGetter interface {
	GetHealthStatus() (bool, error)
}

// StatsSummaryGetter is a interface that help to get kubelet node stats summary
// type StatsSummaryGetter interface {
// 	GetStatsSummary() (*statsapi.Summary, error)
// }

// Client kubelet client interface
type Client struct {
	logger               *zap.SugaredLogger
	config               *KubeletConfig
	tlsBypassHttpsClient TlsBypassHttpsClient
	instanceId           string
}

// NewClient create a kubelet client
func NewClient(kubeletConfig *KubeletConfig, tlsBypassHttpsClient TlsBypassHttpsClient) *Client {
	logger, _ := GetLogger("KubeletClient")
	instanceId := strings.ToLower(os.Getenv(INSTANCE_ID))
	return &Client{
		logger:               logger,
		config:               kubeletConfig,
		tlsBypassHttpsClient: tlsBypassHttpsClient,
		instanceId:           instanceId,
	}
}

// GetPodsInfo get kubelet pods info from kubelet "/pods" api
func (kc *Client) GetPodsInfo() ([]v1.Pod, error) {
	var kubePods v1.PodList
	data, _, err := kc.tlsBypassHttpsClient.HttpGet(kc.config.ServiceAddr + kc.config.PodsAPI)
	if err != nil {
		//metrics.VMAgentAPIRequestFailure.WithLabelValues("/kubelet/pods", reflect.TypeOf(err).Name()).Inc()
		kc.logger.Warnw("Failed to get Kubelet pods info.", "Error", err.Error())
		return nil, err
	}
	_ = json.Unmarshal(data, &kubePods)
	kc.removeHostName(kubePods.Items)
	return kubePods.Items, nil
}

// GetRunningPods get all running pods in kubelet via the "runningpods" api
func (kc *Client) GetRunningPods() (*v1.PodList, error) {
	var kubePods v1.PodList
	data, _, err := kc.tlsBypassHttpsClient.HttpGet(kc.config.ServiceAddr + kc.config.RunningPodsAPI)
	if err != nil {
		kc.logger.Warnw("Failed to get Kubelet running pods info.", "Error", err.Error())
		return nil, err
	}
	err = json.Unmarshal(data, &kubePods)
	if err != nil {
		return nil, err
	}
	return &kubePods, nil
}

func (kc *Client) removeHostName(pods []v1.Pod) {
	if len(kc.instanceId) > 0 {
		suffix := fmt.Sprintf("-%s", kc.instanceId)
		start := 0
		for start < len(pods) {
			pod := &pods[start]
			start++
			pod.Name = strings.TrimSuffix(pod.Name, suffix)
		}
	} else {
		kc.logger.Warn("Can't find INSTANCE_ID from env")
	}
}

// GetHealthStatus get kubelet healthz status from kubelet "/healthz" api return true if kubelet is in good health
func (kc *Client) GetHealthStatus() (bool, error) {
	_, statusCode, err := kc.tlsBypassHttpsClient.HttpGet(kc.config.ServiceAddr + kc.config.HealthzAPI)
	if err != nil {
		//metrics.VMAgentAPIRequestFailure.WithLabelValues("/kubelet/healthz", reflect.TypeOf(err).Name()).Inc()
		kc.logger.Warnw("Failed to get Kubelet healthz info.", "Error", err.Error())
		return false, err
	}
	return statusCode == http.StatusOK, nil
}

// GetStatsSummary get kubelet stats summary from kubelet "/stats/summary" api
// func (kc *Client) GetStatsSummary() (*statsapi.Summary, error) {
// 	var stats statsapi.Summary
// 	data, _, err := kc.tlsBypassHttpsClient.HttpGet(kc.config.ServiceAddr + kc.config.StatsSummaryAPI)
// 	if err != nil {
// 		kc.logger.Warnw("Failed to get Kubelet node stats summary.", "Error", err.Error())
// 		return nil, err
// 	}
// 	err = json.Unmarshal(data, &stats)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &stats, nil
// }
