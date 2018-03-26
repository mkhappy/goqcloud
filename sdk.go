package qcloud

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/lauyoume/gohttp"
)

var (
	default_server_uri = "/v2/index.php"
)

type Qcloud struct {
	config     *Config
	serverHost string
	serverUri  string
	actionName string
}

func New(cfg *Config) *Qcloud {
	if cfg.DefaultRegion == "" {
		cfg.DefaultRegion = "gz"
	}

	if cfg.RequestMethod == "" {
		cfg.RequestMethod = "GET"
	}

	cfg.RequestMethod = strings.ToUpper(cfg.RequestMethod)

	return &Qcloud{
		config:    cfg,
		serverUri: default_server_uri,
	}
}

func (q *Qcloud) Module(name string) *Qcloud {
	q.serverHost = get_know_module_host(name)
	return q
}

func (q *Qcloud) Action(action string) *Qcloud {
	if action != "" {
		q.actionName = strings.ToUpper(action[:1]) + action[1:]
	}
	return q
}

func (q *Qcloud) Send(params map[string]interface{}) (string, error) {

	if q.actionName == "" {
		return "", errors.New("empty action")
	}

	if params == nil {
		params = make(map[string]interface{})
	}
	params["Action"] = q.actionName
	if _, ok := params["SecretId"]; !ok {
		params["SecretId"] = q.config.SecretId
	}
	if _, ok := params["Timestamp"]; !ok {
		params["Timestamp"] = time.Now().Unix()
	}
	if _, ok := params["Nonce"]; !ok {
		params["Nonce"] = 1 + rand.Intn(65534)
	}
	if _, ok := params["Region"]; !ok {
		params["Region"] = q.config.DefaultRegion
	}

	params["Signature"] = sign(make_sign_plain_text(params, q.config.RequestMethod, q.serverHost, q.serverUri), q.config.SecretKey)

	req := gohttp.New()
	uri := fmt.Sprintf("https://%s%s", q.serverHost, q.serverUri)

	var (
		resp *http.Response
		errs []error
	)

	params_str := http_query_params(params)
	if q.config.RequestMethod == "GET" {
		resp, errs = req.Get(uri).Query(params_str).Timeout(60 * time.Second).End()
	} else {
		resp, errs = req.Post(uri).Type("form").Send(params_str).Timeout(60 * time.Second).End()
	}

	if errs != nil {
		return "", errs[0]
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func sign(str, secretkey string) string {
	mac := hmac.New(sha1.New, []byte(secretkey))
	mac.Write([]byte(str))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func make_sign_plain_text(params map[string]interface{}, method, host, path string) string {
	return fmt.Sprintf("%s%s%s?%s", method, host, path, build_query_params(params, method))
}

func build_query_params(params map[string]interface{}, method string) string {
	keys := make([]string, 0)
	for k, _ := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	str := ""
	for _, k := range keys {
		if str != "" {
			str += "&"
		}
		if val, ok := params[k]; ok {
			val_str := fmt.Sprintf("%v", val)
			if method == "POST" && strings.HasPrefix(val_str, "@") {
				continue
			}
			k = strings.Replace(k, "_", ".", -1)
			str += fmt.Sprintf("%s=%v", k, val)
		}
	}

	return str
}

func http_query_params(params map[string]interface{}) string {
	str := ""
	for k, v := range params {
		if str != "" {
			str += "&"
		}
		k = strings.Replace(k, "_", ".", -1)
		val := fmt.Sprintf("%v", v)
		str += fmt.Sprintf("%s=%s", k, url.QueryEscape(val))
	}

	return str
}
