package function

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func Handle(req []byte) string {

	grafanaUrl := os.Getenv("grafana_url")
	if len(grafanaUrl) == 0 {
		grafanaUrl = "http://grafana:3000"
	}

	ignoreCert := false
	envSkipVerify := os.Getenv("skip_tls_verify")
	if len(envSkipVerify) == 0 {
		ignoreCert = false
	} else {
		v, err := strconv.ParseBool(envSkipVerify)
		if err != nil {
			ignoreCert = false
		} else {
			ignoreCert = v
		}
	}

	payload := createPayload(string(req), grafanaUrl)

	payloadJson, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/api/annotations", grafanaUrl)

	scheme := "http"
	if strings.Contains(grafanaUrl, "https") {
		scheme = "https"
	}
	trIgnore := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{}
	if scheme == "https" && ignoreCert {
		client = &http.Client{Transport: trIgnore}
	}

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadJson))

	token, tokenErr := getSecret("grafana_api_token")

	var contextMessage string

	// try to use grafana_api_token secret first, then basic auth
	if tokenErr != nil {
		contextMessage = fmt.Sprintf("Grafana credential error: API token not found. Using basic auth: %v", tokenErr)
		username, userErr := getSecret("grafana_username")
		password, pwdErr := getSecret("grafana_password")

		if userErr != nil {
			contextMessage += fmt.Sprintf("Grafana credential error: username not found. Using default... %v", userErr)
			username = []byte(`admin`)
		}

		if pwdErr != nil {
			contextMessage += fmt.Sprintf("Grafana credential error: password not found. Using default... %v", pwdErr)
			password = []byte(`admin`)
		}
		request.SetBasicAuth(string(username), string(password))
	} else {
		contextMessage = "Grafana API token enabled. Using API token..."
		request.Header.Add("Authorization", "Bearer "+string(token))
	}

	request.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(request)

	if err != nil {
		return fmt.Sprintf("Failed to create annotation %v", err)
	} else {
		responseBody, responseErr := ioutil.ReadAll(resp.Body)
		if responseErr != nil {
			contextMessage += fmt.Sprintf("%v", responseErr)
		}
		if resp.StatusCode != 200 {
			errorResponse := map[string]interface{}{
				"statusCode": resp.StatusCode,
				"message":    fmt.Sprintf("Failed to post annotation with status: %v", resp.Status),
				"context":    contextMessage,
			}
			errorJson, _ := json.Marshal(errorResponse)
			return string(errorJson)
		} else {
			return string(responseBody)
		}
	}
}

func getSecret(secretName string) (secretBytes []byte, err error) {
	// read from the openfaas secrets folder
	secretBytes, err = ioutil.ReadFile("/var/openfaas/secrets/" + secretName)
	return secretBytes, err
}

func createPayload(input string, queryString string) map[string]interface{} {

	now := time.Now()
	nanos := now.UnixNano()
	millis := nanos / 1000000

	payload := map[string]interface{}{}

	params, paramsError := url.ParseQuery(os.Getenv("Http_Query"))
	if paramsError != nil {
		payload = map[string]interface{}{
			"time":     millis,
			"isRegion": false,
			"timeEnd":  0,
			"text":     input,
			"tags":     []string{"global"},
		}
	} else {
		payload = map[string]interface{}{
			"time":     millis,
			"isRegion": false,
			"timeEnd":  0,
			"text":     input,
		}
		if len(params["tag"]) > 0 {
			newTags := []string{}
			for _, tag := range params["tag"] {
				newTags = append(newTags, tag)
			}
			payload["tags"] = newTags
		} else {
			payload["tags"] = []string{"global"}
		}
		if len(params["dashboardId"]) > 0 {
			dashId, dashErr := strconv.Atoi(params["dashboardId"][0])
			if dashErr == nil {
				payload["dashboardId"] = dashId
			}
		}
		if len(params["panelId"]) > 0 {
			panelId, panelErr := strconv.Atoi(params["panelId"][0])
			if panelErr == nil {
				payload["panelId"] = panelId
			}
		}
	}
	return payload
}
