package datahub

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type DatahubConfig struct {
	BaseUrl string
}

type DatahubAuth struct {
	ClientID     string
	ClientSecret string
	Token        string
	ExpiresAt    time.Time
}

type DatahubClient struct {
	config     DatahubConfig
	httpClient *http.Client
	auth       DatahubAuth
}

func NewDatahubClient(baseUrl, clientID, clientSecret string) *DatahubClient {

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
		
	}
	
	if strings.Contains(baseUrl, "localhost") || strings.Contains(baseUrl, "127.0.0.1"){
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client := &http.Client{Transport: tr}

	return &DatahubClient{
		httpClient: client,
		config: DatahubConfig{
			BaseUrl: baseUrl,
		},
		auth: DatahubAuth{
			ClientID:     clientID,
			ClientSecret: clientSecret,
		},
	}
}

func (dc *DatahubClient) authenticate(ctx context.Context) error {
	if dc.auth.Token != "" && dc.auth.ExpiresAt.After(time.Now().Add(-10*time.Second)) {
		return nil
	}

	endpoint := "/api/v1/auth/token"

	URL := dc.config.BaseUrl + endpoint

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("scope", "client_scope")

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, URL, strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		return err
	}

	request.SetBasicAuth(dc.auth.ClientID, dc.auth.ClientSecret)

	response, err := dc.httpClient.Do(request)
	if err != nil {
		return err
	}

	if response.Body != nil {
		defer response.Body.Close()
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	tokenDetails := TokenResponse{}
	err = json.Unmarshal(body, &tokenDetails)
	if err != nil {
		log.Fatal(err)
		return err
	}

	dc.auth.Token = tokenDetails.AccessToken
	dc.auth.ExpiresAt = time.Now().Add(time.Duration(tokenDetails.ExpiresIn) * time.Second)
	return nil
}

func (dc *DatahubClient) do(ctx context.Context, method string, endpoint string, jsonObject interface{}) (*http.Response, error) {
	dc.authenticate(ctx)
	tflog.Debug(ctx, "Executed autenticate")

	URL := dc.config.BaseUrl + endpoint
	tflog.Debug(ctx, "URL = " + URL)

	var body *bytes.Reader
	if jsonObject != nil {
		bodyObj, err := json.Marshal(jsonObject)
		if err != nil {
			tflog.Debug(ctx, "Could not marshal JSON Body")
			return nil, err
		}

		body = bytes.NewReader(bodyObj)
		
	}else{
		
		body = bytes.NewReader([]byte{})
	}

	req, err := http.NewRequestWithContext(ctx, method, URL, body)
	if err != nil {
		tflog.Debug(ctx, "Failed DO request create" + err.Error())
		return nil, err
	}


	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+dc.auth.Token)

	return dc.httpClient.Do(req)
}

func (dc *DatahubClient) CreateJob(ctx context.Context, job CreateJobRequest) (JobResponse, error) {
	tflog.Debug(ctx, "Starting datahub CREATE JOB")
	endpoint := "/api/v1/job"
	method := http.MethodPost

	resp, err := dc.do(ctx, method, endpoint, job)
	if err != nil {
		tflog.Debug(ctx, "Error response in CREATE JOB")
		return JobResponse{}, err
	}

	
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return JobResponse{}, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
        return JobResponse{}, errors.New("Datahub Engine API returned error response: " +string(body))
	}


	jobInfo := JobResponse{}
	err = json.Unmarshal(body, &jobInfo)
	if err != nil {
		tflog.Debug(ctx, "Cannot Unmarshal response body = " + string(body))
		return JobResponse{}, errors.New("Cannot decode body into response:" + string(body))
	}

	return jobInfo, nil
}

func (dc *DatahubClient) GetJob(ctx context.Context, jobID string) (JobResponse, error) {
	endpoint := "/api/v1/job/" + jobID
	method := http.MethodGet

	resp, err := dc.do(ctx, method, endpoint, nil)
	if err != nil {
		return JobResponse{}, err
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return JobResponse{}, err
	}

	jobInfo := JobResponse{}
	err = json.Unmarshal(body, &jobInfo)
	if err != nil {
		return JobResponse{}, err
	}

	return jobInfo, nil
}

func (dc *DatahubClient) UpdateJob(ctx context.Context, jobID string, jobUpdate UpdateJobRequest) (JobResponse, error) {
	endpoint := "/api/v1/job/" + jobID
	method := http.MethodPatch

	resp, err := dc.do(ctx, method, endpoint, jobUpdate)
	if err != nil {
		return JobResponse{}, err
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return JobResponse{}, err
	}

	jobInfo := JobResponse{}
	err = json.Unmarshal(body, &jobInfo)
	if err != nil {
		return JobResponse{}, err
	}

	return jobInfo, nil
}
