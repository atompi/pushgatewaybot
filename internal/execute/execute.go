package execute

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/atompi/pushgatewaybot/internal/options"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func Execute(opts options.Options) {
	for _, exporter := range opts.Exporters {
		c := cron.New(cron.WithSeconds())
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		t := fmt.Sprintf("%s/%s * * * * *", strconv.Itoa(r.Intn(exporter.Interval+1)), strconv.Itoa(exporter.Interval))
		c.AddFunc(t, func() {
			metricsURL := exporter.URL
			metricsResp, err := http.Get(metricsURL)
			if err != nil {
				zap.L().Sugar().Errorf("get metrics error: %v", err)
				return
			}
			defer metricsResp.Body.Close()
			if metricsResp.StatusCode != http.StatusOK {
				zap.L().Sugar().Errorf("get metrics error: %v", err)
				return
			}

			pushClient := &http.Client{}
			hostname, err := os.Hostname()
			if err != nil {
				zap.L().Sugar().Errorf("get hostname error: %v", err)
				return
			}
			pushURL := opts.Pushgateway.URL + "/metrics/job/edge_node_exporter/instance/" + hostname
			if strings.HasPrefix(pushURL, "https://") {
				clientTLSCert, err := tls.LoadX509KeyPair(opts.Pushgateway.CertFile, opts.Pushgateway.KeyFile)
				if err != nil {
					zap.L().Sugar().Errorf("load x509 key pair error: %v", err)
					return
				}

				RootCAPool, err := x509.SystemCertPool()
				if err != nil {
					zap.L().Sugar().Errorf("load root ca pool error: %v", err)
					return
				}
				if caCert, err := os.ReadFile(opts.Pushgateway.CAFile); err != nil {
					zap.L().Sugar().Errorf("read ca cert file error: %v", err)
					return
				} else if ok := RootCAPool.AppendCertsFromPEM(caCert); !ok {
					zap.L().Sugar().Errorf("append ca cert to cert pool error: %v", err)
					return
				}

				tlsConfig := &tls.Config{
					RootCAs: RootCAPool,
					Certificates: []tls.Certificate{
						clientTLSCert,
					},
					InsecureSkipVerify: opts.Pushgateway.InsecureSkipVerify,
				}
				tr := &http.Transport{
					TLSClientConfig: tlsConfig,
				}
				pushClient = &http.Client{Transport: tr}
			}

			pushReq, err := http.NewRequest(http.MethodPost, pushURL, metricsResp.Body)
			if err != nil {
				zap.L().Sugar().Errorf("new push post request error: %v", err)
				return
			}
			auth := opts.Pushgateway.Auth.Username + ":" + opts.Pushgateway.Auth.Password
			basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
			pushReq.Header.Set("Authorization", basicAuth)

			pushResp, err := pushClient.Do(pushReq)
			if err != nil {
				zap.L().Sugar().Errorf("push metrics error: %v", err)
				return
			}
			defer pushResp.Body.Close()
			if pushResp.StatusCode != http.StatusOK {
				zap.L().Sugar().Errorf("push metrics error: %v", err)
				return
			}
		})
		go c.Start()
	}
}
