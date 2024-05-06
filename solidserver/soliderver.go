package solidserver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/parnurzeal/gorequest"
)

type HttpRequestFunc func(*gorequest.SuperAgent, string) *gorequest.SuperAgent

var httpRequestMethods = map[string]HttpRequestFunc{
	"post":   (*gorequest.SuperAgent).Post,
	"put":    (*gorequest.SuperAgent).Put,
	"delete": (*gorequest.SuperAgent).Delete,
	"get":    (*gorequest.SuperAgent).Get,
}

const regexpIPPort = `^!?(([0-9]{1,3})\.){3}[0-9]{1,3}:[0-9]{1,5}$`
const regexpHostname = `^(([a-z0-9]|[a-z0-9][a-z0-9\-]*[a-z0-9])\.)*([a-z0-9]|[a-z0-9][a-z0-9\-]*[a-z0-9])$`
const regexpNetworkAcl = `^!?(([0-9]{1,3})\.){3}[0-9]{1,3}/[0-9]{1,2}$`

type SOLIDserver struct {
	Ctx                      context.Context
	Host                     string
	Username                 string
	Password                 string
	BaseUrl                  string
	SSLVerify                bool
	AdditionalTrustCertsFile string
	Timeout                  int
	Version                  int
	Authenticated            bool
	ProxyURL                 string
}

func NewSOLIDserver(ctx context.Context, host string, username string, password string, sslverify bool, certsfile string, timeout int, version string, proxyURL string) (*SOLIDserver, diag.Diagnostics) {
	s := &SOLIDserver{
		Ctx:                      ctx,
		Host:                     host,
		Username:                 username,
		Password:                 password,
		BaseUrl:                  "https://" + host,
		SSLVerify:                sslverify,
		AdditionalTrustCertsFile: certsfile,
		Timeout:                  timeout,
		Version:                  0,
		Authenticated:            false,
		ProxyURL:                 proxyURL,
	}

	if err := s.GetVersion(version); err != nil {
		return nil, err
	}

	return s, nil
}

func SubmitRequest(s *SOLIDserver, apiclient *gorequest.SuperAgent, method string, service string, parameters string) (*http.Response, string, error) {
	var resp *http.Response = nil
	var body string = ""
	var errs []error = nil
	var requestUrl string = ""

	var httpRequestTimings = map[string]struct {
		msSweep  int
		sTimeout int
		maxTry   int
	}{
		"post":   {msSweep: 16, sTimeout: s.Timeout, maxTry: 1},
		"put":    {msSweep: 16, sTimeout: s.Timeout, maxTry: 1},
		"delete": {msSweep: 16, sTimeout: s.Timeout, maxTry: 1},
		"get":    {msSweep: 16, sTimeout: s.Timeout, maxTry: 6},
	}

	// Get the SystemCertPool, continue with an empty pool on error
	rootCAs, x509err := x509.SystemCertPool()

	if rootCAs == nil || x509err != nil {
		rootCAs = x509.NewCertPool()
	}

	if s.AdditionalTrustCertsFile != "" {
		certs, readErr := ioutil.ReadFile(s.AdditionalTrustCertsFile)
		tflog.Debug(s.Ctx, fmt.Sprintf("Certificates = %s\n", certs))

		if readErr != nil {
			tflog.Error(s.Ctx, fmt.Sprintf("Failed to append %q to RootCAs: %v\n", s.AdditionalTrustCertsFile, readErr))
			os.Exit(1)
		}

		tflog.Debug(s.Ctx, fmt.Sprintf("Cert Subjects Before Append = %d\n", len(rootCAs.Subjects())))

		if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
			tflog.Debug(s.Ctx, fmt.Sprintf("No certs appended, using system certs only\n"))
		}

		tflog.Debug(s.Ctx, fmt.Sprintf("Cert Subjects After Append = %d\n", len(rootCAs.Subjects())))
	}

	t := httpRequestTimings[method]

	tflog.Debug(s.Ctx, fmt.Sprintf("Timings for method '%s' : {%v}\n", method, t))

	apiclient.Timeout(time.Duration(t.sTimeout) * time.Second)

	retryCount := 0

KeepTrying:
	for retryCount < t.maxTry {

		httpFunc, ok := httpRequestMethods[method]

		if !ok {
			return nil, "", fmt.Errorf("Unsupported HTTP request '%s'\n", method)
		}

		// Random Delay for write operation to distribute the load
		time.Sleep(time.Duration(rand.Intn(t.msSweep)) * time.Millisecond)

		requestUrl = fmt.Sprintf("%s/%s?%s", s.BaseUrl, service, parameters)

		resp, body, errs = httpFunc(apiclient, requestUrl).
			TLSClientConfig(&tls.Config{InsecureSkipVerify: !s.SSLVerify, RootCAs: rootCAs}).
			Set("X-IPM-Username", base64.StdEncoding.EncodeToString([]byte(s.Username))).
			Set("X-IPM-Password", base64.StdEncoding.EncodeToString([]byte(s.Password))).
			End()

		if errs == nil {
			return resp, body, nil
		}

		tflog.Debug(s.Ctx, fmt.Sprintf("'%s' API request '%s' failed with errors.\n", method, requestUrl))

		for _, err := range errs {
			if err, ok := err.(net.Error); ok && err.Timeout() {
				tflog.Debug(s.Ctx, fmt.Sprintf("Timeout Retry (%d/%d)\n", retryCount+1, t.maxTry))
				retryCount++
				continue KeepTrying
			}

			return nil, "", fmt.Errorf("Non-Retryable error (%q): Bailing out\n", err)
		}
	}

	return nil, "", fmt.Errorf("Error '%s' API request '%s' : timeout retry count exceeded (maxTry = %d) !\n", method, requestUrl, t.maxTry)
}

func (s *SOLIDserver) GetVersion(version string) diag.Diagnostics {

	apiclient := gorequest.New()
	apiclient.Proxy(s.ProxyURL)

	parameters := url.Values{}
	parameters.Add("WHERE", "member_is_me='1'")

	resp, body, err := SubmitRequest(s, apiclient, "get", "rest/member_list", parameters.Encode())

	if err == nil && resp.StatusCode == 200 {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		if rversion, rversionExist := buf[0]["member_version"].(string); rversionExist {
			//tflog.Debug(s.Ctx, fmt.Sprintf("Version: %s\n", rversion))
			StrVersion := strings.Split(rversion, ".")

			// Computing version number
			for i := 0; i < len(StrVersion) && i < 3; i++ {
				num, numErr := strconv.Atoi(StrVersion[i])
				if numErr == nil {
					s.Version = s.Version*10 + num
				} else {
					s.Version = s.Version*10 + 0
				}
			}

			// Handling new branch version
			if s.Version < 100 {
				s.Version = s.Version * 10
			}

			tflog.Debug(s.Ctx, fmt.Sprintf("SOLIDserver version retrieved from remote SOLIDserver: %d\n", s.Version))

			return nil
		}
	}

	if err == nil && (resp.StatusCode <= 400 && resp.StatusCode < 500) {
		if version != "" {
			StrVersion := strings.Split(version, ".")

			for i := 0; i < len(StrVersion) && i < 3; i++ {
				num, numErr := strconv.Atoi(StrVersion[i])
				if numErr == nil {
					s.Version = s.Version*10 + num
				} else {
					s.Version = s.Version*10 + 0
				}
			}
			tflog.Debug(s.Ctx, fmt.Sprintf("Error retrieving SOLIDserver Version (Insufficient Permissions)."))
			tflog.Debug(s.Ctx, fmt.Sprintf("SOLIDserver version retrived from local provider parameter: %d\n", s.Version))

			return nil
		} else {
			return diag.Errorf("Error retrieving SOLIDserver Version (Insufficient Permissions). Consider setting the SOLIDserver's version using the provider options.\n")
		}
	}

	if err != nil {
		return diag.Errorf("Error retrieving SOLIDserver Version (%s)\n", err)
	}

	return diag.Errorf("Error retrieving SOLIDserver Version (No Answer)\n")
}

func (s *SOLIDserver) Request(method string, service string, parameters *url.Values) (*http.Response, string, error) {
	var resp *http.Response = nil
	var body string = ""
	var err error = nil

	apiclient := gorequest.New()
	apiclient.Proxy(s.ProxyURL)

	if s.Authenticated == false {
		apiclient.Retry(3, time.Duration(rand.Intn(15)+1)*time.Second, http.StatusTooManyRequests, http.StatusInternalServerError)
	} else {
		apiclient.Retry(3, time.Duration(rand.Intn(15)+1)*time.Second, http.StatusRequestTimeout, http.StatusTooManyRequests, http.StatusInternalServerError, http.StatusUnauthorized)
	}

	resp, body, err = SubmitRequest(s, apiclient, method, service, parameters.Encode())

	if err != nil {
		return nil, "", fmt.Errorf("SOLIDServer - Error initiating API call (%q)\n", err)
	}

	if len(body) > 0 && body[0] == '{' && body[len(body)-1] == '}' {
		tflog.Debug(s.Ctx, fmt.Sprintf("Repacking HTTP JSON Body\n"))
		body = "[" + body + "]"
	}

	if s.Authenticated == false && (200 <= resp.StatusCode && resp.StatusCode <= 204) {
		s.Authenticated = true
	}

	return resp, body, nil
}
