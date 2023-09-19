package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/armosec/armoapi-go/apis"
	"github.com/armosec/utils-go/httputils"
	logger "github.com/kubescape/go-logger"
	"github.com/kubescape/k8s-interface/k8sinterface"
	"github.com/kubescape/kubescape/v2/core/cautils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	operatorServicePort string = "4002"
	operatorTriggerPath string = "v1/triggerAction"
	kubescapeNamespace  string = "kubescape"
)

type OperatorAdapter struct {
	httpPostFunc func(httputils.IHttpClient, string, map[string]string, []byte) (*http.Response, error)
	cautils.OperatorScanInfo
	cautils.OperatorConnector
}

func getOperatorPod(k8sClient *k8sinterface.KubernetesApi) (*v1.Pod, error) {
	listOptions := metav1.ListOptions{
		LabelSelector: "app=operator",
	}
	pods, err := k8sClient.KubernetesClient.CoreV1().Pods(kubescapeNamespace).List(k8sClient.Context, listOptions)
	if err != nil {
		return nil, err
	}
	if len(pods.Items) != 1 {
		return nil, errors.New("in 'getOperatorPod' can't find specific operator pod")
	}

	return &pods.Items[0], nil
}

func NewOperatorAdapter(scanInfo cautils.OperatorScanInfo) (*OperatorAdapter, error) {
	k8sClient := getKubernetesApi()
	pod, err := getOperatorPod(k8sClient)
	if err != nil {
		return nil, err
	}
	operatorConnector, err := cautils.CreatePortForwarder(k8sClient, pod, operatorServicePort, kubescapeNamespace)
	if err != nil {
		return nil, err
	}

	return &OperatorAdapter{
		httpPostFunc:      httputils.HttpPost,
		OperatorScanInfo:  scanInfo,
		OperatorConnector: operatorConnector,
	}, nil
}

func (a *OperatorAdapter) httpPostOperatorScanRequest(body apis.Commands) (string, error) {
	reqBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("in 'httpPostOperatorScanRequest' failed to json.Marshal, reason: %v", err)
	}

	err = a.StartPortForwarder()
	if err != nil {
		return "", err
	}
	defer a.StopPortForwarder()

	urlQuery := url.URL{
		Scheme: "http",
		Host:   a.GetPortForwardLocalhost(),
		Path:   operatorTriggerPath,
	}

	resp, err := a.httpPostFunc(http.DefaultClient, urlQuery.String(), map[string]string{"Content-Type": "application/json"}, reqBody)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	return httputils.HttpRespToString(resp)
}

func (a *OperatorAdapter) OperatorScan() (string, error) {
	logger.L().Info("Triggering scan initiated, Scan results will be available at the kubescape backend config in your cluster")
	res, err := a.httpPostOperatorScanRequest(*a.OperatorScanInfo.GetRequestPayload())
	if err != nil {
		return "", err
	}
	logger.L().Info("Triggering scan finished successfully")
	return res, nil
}
