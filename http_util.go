package example

import (
	"bytes"
	"crypto/tls"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"
	//"goms.io/azureml/mir/mir-vmagent/pkg/log"
)

const DefaultTimeoutInSec = 2

type HttpRequestSender interface {
	// SendRequestWithFullReturn encapsulate error handling of http request.
	// Return HTTP status code, response header, pointer of response body byte slice and error.
	// A non-2xx status code doesn't cause an error.
	SendRequestWithFullReturn(method string, uri string, paramMap map[string]string, body []byte, headerMap map[string]string, timeoutInSec int64, isTlsBypass bool) (int, http.Header, []byte, error)

	// SendRequest encapsulate error handling of http request.
	// Return HTTP status code, response header, pointer of response body byte slice and error.
	// A non-2xx status code doesn't cause an error.
	SendRequestWithTimeout(method string, uri string, paramMap map[string]string, body []byte, headerMap map[string]string, timeoutInSec int64) (int, []byte, error)

	// SendRequest send given http request with a DefaultTimeoutInSec timeout
	SendRequest(method string, uri string, paramMap map[string]string, body []byte, headerMap map[string]string) (int, []byte, error)

	// SendRequest send given http request with TLS bypass and with a DefaultTimeoutInSec timeout
	SendRequestWithTlsBypass(method string, uri string, paramMap map[string]string, body []byte, headerMap map[string]string) (int, []byte, error)
}

type DefaultHttpRequestSender struct {
	Logger *zap.SugaredLogger
}

func NewDefaultHttpRequestSender() *DefaultHttpRequestSender {
	logger, _ := GetLogger("DefaultHttpRequestSender")

	return &DefaultHttpRequestSender{
		Logger: logger,
	}
}

func (se *DefaultHttpRequestSender) SendRequestWithFullReturn(method string, uri string, paramMap map[string]string, body []byte, headerMap map[string]string, timeoutInSec int64, isTlsBypass bool) (int, http.Header, []byte, error) {
	return SendRequestWithFullReturn(method, uri, paramMap, body, headerMap, timeoutInSec, false)
}

func (se *DefaultHttpRequestSender) SendRequestWithTimeout(method string, uri string, paramMap map[string]string, body []byte, headerMap map[string]string, timeoutInSec int64) (int, []byte, error) {
	statusCode, _, responseBytes, err := se.SendRequestWithFullReturn(method, uri, paramMap, body, headerMap, timeoutInSec, false)
	return statusCode, responseBytes, err
}

func (se *DefaultHttpRequestSender) SendRequest(method string, uri string, paramMap map[string]string, body []byte, headerMap map[string]string) (int, []byte, error) {
	return se.SendRequestWithTimeout(method, uri, paramMap, body, headerMap, DefaultTimeoutInSec)
}

func (se *DefaultHttpRequestSender) SendRequestWithTlsBypass(method string, uri string, paramMap map[string]string, body []byte, headerMap map[string]string) (int, []byte, error) {
	statusCode, _, responseBytes, err := SendRequestWithFullReturn(method, uri, paramMap, body, headerMap, DefaultTimeoutInSec, true)
	return statusCode, responseBytes, err
}

type TlsBypassHttpsClient interface {
	HttpGet(url string) ([]byte, int, error)
}

type DefaultTlsBypassHttpsClient struct {
	Logger *zap.SugaredLogger
	Client *http.Client
}

func NewDefaultTlsBypassHttpsClient() *DefaultTlsBypassHttpsClient {
	logger, _ := GetLogger("DefaultTlsBypassHttpsClient")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		MaxConnsPerHost: 1,
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(10) * time.Second,
	}
	return &DefaultTlsBypassHttpsClient{
		Logger: logger,
		Client: client,
	}
}

// HttpGet get http GET response body pointer from given url
func (dg *DefaultTlsBypassHttpsClient) HttpGet(url string) ([]byte, int, error) {
	resp, err := dg.Client.Get(url)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	return data, resp.StatusCode, nil
}

// SendRequestWithFullReturn encapsulate error handling of http request.
// Return HTTP status code, response header, pointer of response body byte slice and error.
// A non-2xx status code doesn't cause an error.
func SendRequestWithFullReturn(method string, uri string, paramMap map[string]string, body []byte, headerMap map[string]string, timeoutInSec int64, isTlsBypass bool) (int, http.Header, []byte, error) {
	logger, _ := GetDefaultLogger()

	urlInstance, err := url.Parse(uri)
	if err != nil {
		logger.Error("Error creating URL: ", err)
		return 0, nil, nil, err
	}

	if paramMap != nil {
		params, err := url.ParseQuery(urlInstance.RawQuery)
		if err != nil {
			logger.Errorw("Error parse url query", "error", err, "uri", uri)
			return 0, nil, nil, err
		}
		for k, v := range paramMap {
			params.Add(k, v)
		}

		urlInstance.RawQuery = params.Encode()
	}

	var bodyIOReader io.Reader
	if body != nil {
		bodyIOReader = bytes.NewBuffer(body)
	}

	req, err := http.NewRequest(method, urlInstance.String(), bodyIOReader)
	if err != nil {
		logger.Error("Error creating HTTP request: ", err)
		return 0, nil, nil, err
	}

	if headerMap != nil {
		for k, v := range headerMap {
			req.Header.Add(k, v)
		}
	}
	var client http.Client
	if isTlsBypass {
		client = http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				MaxConnsPerHost: 1,
			},
			Timeout: time.Duration(timeoutInSec) * time.Second,
		}
	} else {
		client = http.Client{
			Timeout: time.Duration(timeoutInSec) * time.Second,
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Warn("Error calling token endpoint: ", err)
		return 0, nil, nil, err
	}

	defer resp.Body.Close()
	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error reading response body: ", err)
		return resp.StatusCode, resp.Header, nil, err
	}

	// If Http status code is not 2xx
	if !(200 <= resp.StatusCode && resp.StatusCode < 300) {
		logger.Warnw("Http status code of response is not 2xx", "HTTP status code", resp.StatusCode)
	}

	return resp.StatusCode, resp.Header, responseBytes, nil
}

type SendHttpRequestWithTimeoutFunc func(method string, uri string, paramMap map[string]string, body []byte, headerMap map[string]string, timeoutInSec int64) (int, []byte, error)

// SendRequest encapsulate error handling of http request.
// Return HTTP status code, response header, pointer of response body byte slice and error.
// A non-2xx status code doesn't cause an error.
func SendRequestWithTimeout(method string, uri string, paramMap map[string]string, body []byte, headerMap map[string]string, timeoutInSec int64) (int, []byte, error) {
	statusCode, _, responseBytes, err := SendRequestWithFullReturn(method, uri, paramMap, body, headerMap, timeoutInSec, false)
	return statusCode, responseBytes, err
}

type SendHttpRequestFunc func(method string, uri string, paramMap map[string]string, body []byte, headerMap map[string]string) (int, []byte, error)

// SendRequest send given http request with a 2s timeout
func SendRequest(method string, uri string, paramMap map[string]string, body []byte, headerMap map[string]string) (int, []byte, error) {
	return SendRequestWithTimeout(method, uri, paramMap, body, headerMap, DefaultTimeoutInSec)
}

// SendRequest send given http request with TLS bypass
func SendRequestWithTlsBypass(method string, uri string, paramMap map[string]string, body []byte, headerMap map[string]string) (int, []byte, error) {
	statusCode, _, responseBytes, err := SendRequestWithFullReturn(method, uri, paramMap, body, headerMap, DefaultTimeoutInSec, true)
	return statusCode, responseBytes, err
}
