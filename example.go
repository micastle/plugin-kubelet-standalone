// Package example is a CoreDNS plugin that prints "example" to stdout on every packet received.
//
// It serves as an example CoreDNS plugin with numerous code comments.
package example

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/coredns/coredns/request"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	clog "github.com/coredns/coredns/plugin/pkg/log"

	"github.com/miekg/dns"
)

// Define log to be a logger with the plugin name in it. This way we can just use log.Info and
// friends to log.
var log = clog.NewWithPlugin("example")

type MyError struct {
	When time.Time
	What string
}

func (e *MyError) Error() string {
	return e.What
}

func (e *MyError) Errors() string {
	return fmt.Sprintf("at %v, %s",
		e.When, e.What)
}

func (e *MyError) Strings() string {
	return fmt.Sprintf("%s, at %v",
		e.What, e.When)
}

// Example is an example plugin to show how to write a plugin.
type Example struct {
	Next       plugin.Handler
	KubeClient *Client
	Logger     *zap.SugaredLogger

	recordLock sync.Mutex
	Records    []*PodRecord
}

type PodRecord struct {
	Name string
	Ip   string
	Port int32
}

func (e *Example) GetEnvConfig(envVar string, default_val int) (int, error) {
	valueStr, ok := os.LookupEnv(envVar)
	if !ok {
		return default_val, nil
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		e.Logger.Warnw("configured environment variable value is not integer, take default value", "EnvVar", envVar, "Error", err)
		return default_val, err
	}

	return value, nil
}

func (e *Example) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	return e.ServeDNS_WhoAmI(ctx, w, r)
	// return e.ServeDNS_Example(ctx, w, r)
}

func MakeCopy(records []*PodRecord) []*PodRecord {
	dest := make([]*PodRecord, 0, len(records))

	for _, rc := range records {
		dest = append(dest, rc)
	}

	return dest
}

func (e *Example) GetRecords() []*PodRecord {
	e.recordLock.Lock()
	defer e.recordLock.Unlock()

	e.Logger.Infow("GetRecords", "record count", len(e.Records))
	dest := make([]*PodRecord, 0, len(e.Records))

	for _, rc := range e.Records {
		dest = append(dest, rc)
	}

	e.Logger.Infow("GetRecords", "record count", len(dest))
	return dest
}
func (e *Example) UpdateRecords(records []*PodRecord) {
	e.Logger.Infow("UpdateRecords", "record count", len(records))
	e.recordLock.Lock()
	defer e.recordLock.Unlock()

	e.Records = records
}
func (e *Example) BackgroundLoop() {
	seconds, _ := e.GetEnvConfig("KUBELET_STATUS_SYNC_INTERVAL", 10)
	for {
		// get pod info
		pods, err := e.KubeClient.GetPodsInfo()
		if err != nil {
			e.Logger.Warnw("Getting pods info failed!", "Error", err)
		} else {
			records := e.GetUserPodRecords(pods)
			e.UpdateRecords(records)
		}

		time.Sleep(time.Duration(seconds) * time.Second)
	}

}
func (e *Example) GetUserPodRecords(pods []v1.Pod) []*PodRecord {
	records := make([]*PodRecord, 0, 10)
	for idx := range pods {
		if pods[idx].Labels["userPod"] == "true" {
			name := pods[idx].Name
			last_index := strings.LastIndex(name, "-")
			if last_index > 0 {
				name = string([]rune(name)[:last_index])
			}
			ip := pods[idx].Status.PodIP
			port := pods[idx].Spec.Containers[0].Ports[0].ContainerPort
			e.Logger.Infow("Pod Info", "Info", pods[idx])
			e.Logger.Infow("Pod Info", "Name", name, "IP", ip, "port", port)
			rc := &PodRecord{Name: name, Ip: ip, Port: port}
			records = append(records, rc)
		}
	}

	return records
}

func (e *Example) printRecords(records []*PodRecord) {
	for idx := range records {
		e.Logger.Infow("Record Info", "Name", records[idx].Name, "IP", records[idx].Ip, "port", records[idx].Port)
	}
}

func (e *Example) printPods(pods []v1.Pod) {
	e.Logger.Infow("printPods", "msg", "start")

	for idx := range pods {
		if pods[idx].Labels["userPod"] == "true" {
			name := pods[idx].Name
			last_index := strings.LastIndex(name, "-")
			if last_index > 0 {
				name = string([]rune(name)[:last_index])
			}
			ip := pods[idx].Status.PodIP
			port := pods[idx].Spec.Containers[0].Ports[0].ContainerPort
			e.Logger.Infow("Pod Info", "Info", pods[idx])
			e.Logger.Infow("Pod Info", "Name", name, "IP", ip, "port", port)
		}
	}
}

func (e *Example) QueryForPodRecord(name string, state request.Request, ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {

	e.Logger.Infow("QueryForPodRecord", "query for name", name)

	ttl, _ := e.GetEnvConfig("LOCAL_CLUSTER_DNS_RECORD_TTL", 30)

	msg := new(dns.Msg)
	msg.SetReply(r)

	// turn on recursion to make some client tools work like nslookup
	msg.RecursionAvailable = true
	//msg.RecursionDesired = true
	msg.Authoritative = true

	answers := make([]dns.RR, 0, 10)

	records := e.GetRecords()
	e.printRecords(records)
	for _, rc := range records {
		//e.Logger.Debugw("QueryForPodRecord", "record", rc)
		if rc.Name == name {
			var ra dns.RR
			ra = new(dns.A)

			ra.(*dns.A).Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeA, Class: state.QClass(), Ttl: uint32(ttl)}
			ra.(*dns.A).A = net.ParseIP(rc.Ip).To4()

			answers = append(answers, ra)
		}
	}

	if len(answers) == 0 {
		e.Logger.Errorw("QueryForPodRecord", "no matching pod record", len(answers))
		//return dns.RcodeNameError, &MyError{When: time.Now(), What: "it didn't work"}
	}

	msg.Answer = answers

	e.Logger.Infow("QueryForPodRecord", "response", msg)

	w.WriteMsg(msg)

	return dns.RcodeSuccess, nil
}
func (e *Example) ServeDNS_WhoAmI(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {

	state := request.Request{W: w, Req: r}
	e.Logger.Debugw("ServeDNS", "protocol", state.Proto(), "port", state.Port(), "Qname", state.QName(), "Qclass", state.QClass())

	qname := state.QName()
	if strings.HasSuffix(qname, "cluster.local.") {
		index := strings.LastIndex(qname, "cluster.local.")
		podName := string([]rune(qname)[:index])
		if strings.HasSuffix(podName, ".") {
			podName = string([]rune(podName)[:len(podName)-1])
		}
		code, err := e.QueryForPodRecord(podName, state, ctx, w, r)
		return code, err
	}

	// Call next plugin (if any).
	return plugin.NextOrFailure(e.Name(), e.Next, ctx, w, r)
}

// ServeDNS implements the plugin.Handler interface. This method gets called when example is used
// in a Server.
func (e *Example) ServeDNS_Example(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	// This function could be simpler. I.e. just fmt.Println("example") here, but we want to show
	// a slightly more complex example as to make this more interesting.
	// Here we wrap the dns.ResponseWriter in a new ResponseWriter and call the next plugin, when the
	// answer comes back, it will print "example".

	// Debug log that we've have seen the query. This will only be shown when the debug plugin is loaded.
	log.Debug("Received response")

	// Wrap.
	pw := NewResponsePrinter(w)

	// Export metric with the server label set to the current server handling the request.
	requestCount.WithLabelValues(metrics.WithServer(ctx)).Inc()

	// Call next plugin (if any).
	return plugin.NextOrFailure(e.Name(), e.Next, ctx, pw, r)
}

// Name implements the Handler interface.
func (e *Example) Name() string { return "example" }

// ResponsePrinter wrap a dns.ResponseWriter and will write example to standard output when WriteMsg is called.
type ResponsePrinter struct {
	dns.ResponseWriter
}

// NewResponsePrinter returns ResponseWriter.
func NewResponsePrinter(w dns.ResponseWriter) *ResponsePrinter {
	return &ResponsePrinter{ResponseWriter: w}
}

// WriteMsg calls the underlying ResponseWriter's WriteMsg method and prints "example" to standard output.
func (r *ResponsePrinter) WriteMsg(res *dns.Msg) error {
	log.Info("example")
	return r.ResponseWriter.WriteMsg(res)
}
