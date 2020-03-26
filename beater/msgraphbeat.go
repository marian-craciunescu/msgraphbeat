package beater

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"

	"github.com/marian-craciunescu/msgraphbeat/config"
)

var (
	errFirstStart   = errors.New("ErrFirstStart.Registry  File was not yet created")
	errNon2xxStatus = errors.New("errNon2xxStatus .HTTP response error")
)

// msgraphbeat configuration.
type msgraphbeat struct {
	done   chan struct{}
	config config.Config
	client beat.Client
	logger *logp.Logger
	auth   *authInfo
}

// New creates an instance of msgraphbeat.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	logger := logp.NewLogger("msgraphbeat")

	//logger.Info(b.Info)
	c := config.DefaultConfig
	if err := cfg.Unpack(&c); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}
	ai := authInfo{}
	bt := &msgraphbeat{
		done:   make(chan struct{}),
		config: c,
		logger: logger,
		auth:   &ai,
	}
	return bt, nil
}

// Run starts msgraphbeat.
func (msgb *msgraphbeat) Run(b *beat.Beat) error {
	msgb.logger.Info("msgraphbeat is running! Hit CTRL-C to stop it.")

	var err error
	msgb.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	msgb.retrieveSecurityAlerts()

	ticker := time.NewTicker(msgb.config.Period)
	for {
		select {
		case <-msgb.done:
			return nil
		case <-ticker.C:
			msgb.retrieveSecurityAlerts()
		}
	}
}

// Stop stops msgraphbeat.
func (msgb *msgraphbeat) Stop() {
	_ = msgb.client.Close()
	close(msgb.done)
}

func (msgb *msgraphbeat) authenticate() error {

	loginURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", msgb.config.TenantID)
	reqBody := url.Values{}
	reqBody.Set("grant_type", "client_credentials")
	reqBody.Set("scope", "https://graph.microsoft.com/.default")
	reqBody.Set("client_id", msgb.config.ClientID)
	reqBody.Set("client_secret", msgb.config.Secret)
	req, err := http.NewRequest("POST", loginURL, strings.NewReader(reqBody.Encode()))
	if err != nil {
		msgb.logger.Error(err)
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	msgb.logger.Debugf("sending auth req: %v", req)

	httpClient := http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		msgb.logger.Errorf("Error doing request to server err=%v", err)
		return err
	}
	msgb.logger.Debugf("Server response=%i", resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msgb.logger.Errorf("Error doing request to server err=%v", err)
		return err
	}
	err = resp.Body.Close()

	if err != nil {
		msgb.logger.Errorf("Error closing Body", err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("HTTP non2xx.StatusCode=%d,status=%s body=%s",
			resp.StatusCode, resp.Status, string(body))

		msgb.logger.Infof("HTTP non2xx.StatusCode=%d,status=%s body=%s",
			resp.StatusCode, resp.Status, string(body))
	}
	var ai authInfo
	err = json.Unmarshal(body, &ai)
	msgb.logger.Debugf("auth", "got auth info: %v", ai)
	ai.setTime = time.Now()
	msgb.auth = &ai
	return nil
}

func (msgb *msgraphbeat) getRegistry() (time.Time, error) {
	logp.Debug("beat", "getting registry info from %v", msgb.config.RegistryFilePath)
	reg, err := ioutil.ReadFile(msgb.config.RegistryFilePath)
	if err != nil {
		msgb.logger.Info("could not read registry file, may not exist (this is normal on first run). returning earliest possible time.")
		return time.Time{}, errFirstStart
	}
	lastProcessed, err := time.Parse(time.RFC3339, string(reg))
	if err != nil {
		// handle corrupted state file the same way we handle missing state file
		// (alternative: error out and let user try to fix state file)
		msgb.logger.Errorf("error parsing timestamp in registry file (%v): %v; returning earliest possible time.", msgb.config.RegistryFilePath, string(reg))
		return time.Time{}, nil
	}
	return lastProcessed, nil
}

func (msgb *msgraphbeat) putRegistry(lastProcessed time.Time) error {
	msgb.logger.Debugf("putting registry info (%v) to %v", lastProcessed, msgb.config.RegistryFilePath)
	ts := []byte(lastProcessed.Format(time.RFC3339))
	err := ioutil.WriteFile(msgb.config.RegistryFilePath, ts, 0644)
	if err != nil {
		msgb.logger.Error(err)
		return err
	}
	return nil
}

func (msgb *msgraphbeat) doApiRequest(lastProcessed time.Time) ([]byte, error) {
	url := "https://graph.microsoft.com/v1.0/security/alerts"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		logp.Error(err)
		return nil, err
	}
	reqQuery := req.URL.Query()

	//t := lastProcessed.Format("YYYY-MM-DDThh:mm:ss.sssZ")
	t := lastProcessed.UTC().Format(time.RFC3339)
	reqQuery.Add("$filter",
		fmt.Sprintf("lastModifiedDateTime ge %s", t))

	reqQuery.Add("$count", "true")
	req.URL.RawQuery = reqQuery.Encode()

	fmt.Println(req.URL.RawQuery)
	if msgb.auth == nil || msgb.auth.expired() {
		err := msgb.authenticate()
		if err != nil {
			logp.Error(err)
			return nil, err
		}
	}
	req.Header.Set("Authorization", msgb.auth.header())
	httpClient := http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		msgb.logger.Errorf("Error doing request to server err=%v", err)
		return nil, err
	}
	msgb.logger.Debugf("Server response=%i", resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msgb.logger.Errorf("Error doing request to server err=%v", err)
		return nil, err
	}
	err = resp.Body.Close()

	if err != nil {
		msgb.logger.Errorf("Error closing Body", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("HTTP non2xx.StatusCode=%d,status=%s body=%s",
			resp.StatusCode, resp.Status, string(body))

		msgb.logger.Infof("HTTP non2xx.StatusCode=%d,status=%s body=%s",
			resp.StatusCode, resp.Status, string(body))
		return nil, errNon2xxStatus
	}
	return body, nil
}

func (msgb *msgraphbeat) securityAlerts() ([]common.MapStr, error) {
	mapStrArr := make([]common.MapStr, 0)
	lastProcessed, err := msgb.getRegistry()
	if err == errFirstStart {
		lastProcessed, err = time.Parse(time.RFC3339, msgb.config.StartDate)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	body, err := msgb.doApiRequest(lastProcessed)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(body))
	m, err := msgb.jsonBodyToMapArray(body)
	if err != nil {
		return nil, err
	}
	fmt.Println(m)
	mapStrArr, err = msgb.toMapStrArr(m)
	if err != nil {
		return nil, err
	}

	err = msgb.putRegistry(time.Now())
	if err != nil {
		msgb.logger.Info("Error writing registry file")
	}
	return mapStrArr, nil
}

// converts the json response to an array of map of string and generic (int/string,submap,etc)
func (msgb *msgraphbeat) jsonBodyToMapArray(response []byte) ([]map[string]interface{}, error) {
	reader := bytes.NewReader(response)
	dec := json.NewDecoder(reader)

	var m AlertResponse

	if err := dec.Decode(&m); err == io.EOF {
		return m.Value, nil
	} else if err != nil {
		msgb.logger.Errorf("error decoding json response err=%s", err.Error())
		fmt.Println(err.Error())
		return nil, err
	}
	return m.Value, nil
}

//converts a array of map of string and generic to the mapStr structure used by elastic beats framework
func (msgb *msgraphbeat) toMapStrArr(m []map[string]interface{}) (mapStrArr []common.MapStr, err error) {
	for i := range m {
		newMap := m[i]
		mapStr, err := msgb.toMapStr(newMap)
		if err != nil {
			return nil, err
		} else {
			mapStrArr = append(mapStrArr, mapStr)
		}
	}
	return mapStrArr, nil
}

func (msgb *msgraphbeat) toMapStr(initialMap map[string]interface{}) (common.MapStr, error) {
	mapStr := common.MapStr{}
	msgb.recurseAndNormalizeMap("", mapStr, initialMap)
	return mapStr, nil
}

func (msgb *msgraphbeat) recurseAndNormalizeMap(parentKey string, result common.MapStr, initialMap map[string]interface{}) {

	for k := range initialMap {
		switch innerType := initialMap[k].(type) {
		case map[string]interface{}:
			{
				parentKey := fmt.Sprintf("%s.", k)
				msgb.recurseAndNormalizeMap(parentKey, result, innerType)
			}
		case float32, float64, int, int8, int16, int32, int64, string, bool:
			actualKey := parentKey + k
			ecsField := actualKey //msgb.mapper.EcsField(actualKey)

			_, err := result.Put(ecsField, initialMap[k])
			if err != nil {
				msgb.logger.Errorf("Error putting  field in map err=%s", err.Error())
			}
		case nil:
			_, _ = result.Put(parentKey+k, "")
		case interface{}:
			_, _ = result.Put(parentKey+k, innerType)
		default:
			fmt.Println(innerType)
			break
		}
	}

}

func (msgb *msgraphbeat) publishEvents(mapStrArr []common.MapStr) {
	ts := time.Now()
	for _, mapStr := range mapStrArr {
		event := beat.Event{
			Timestamp: ts,
			Fields:    mapStr,
		}
		msgb.client.Publish(event)
	}

}

func (msgb *msgraphbeat) retrieveSecurityAlerts() {
	initialLocationEvents, err := msgb.securityAlerts()
	if err != nil {
		msgb.logger.Error("Error polling security alerts %s", err.Error())
	} else {
		msgb.publishEvents(initialLocationEvents)
	}
}
