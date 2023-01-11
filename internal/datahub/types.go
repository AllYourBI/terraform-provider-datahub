package datahub

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type CreateJobRequest struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Image       string            `json:"image"`
	Environment map[string]string `json:"environment,omitempty"`
	Secrets     map[string]string `json:"secrets,omitempty"`
	Command     []string          `json:"command,omitempty"`
	Oauth       *Oauth            `json:"oauth,omitempty"`
}

type JobResponse struct {
	ClientID    string            `json:"client_id"`
	JobID       string            `json:"job_id"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Image       string            `json:"image"`
	Environment map[string]string `json:"environment,omitempty"`
	Secrets     map[string]string `json:"secrets,omitempty"`
	Command     []string          `json:"command,omitempty"`
	Oauth       *Oauth            `json:"oauth,omitempty"`
}

type Oauth struct {
	Application      string `json:"application,omitempty"`
	Flow             string `json:"flow,omitempty"`
	AuthorizationURL string `json:"authorization_url,omitempty"`
	TokenURL         string `json:"token_url,omitempty"`
	Scope            string `json:"scope,omitempty"`
	ConfigPrefix     string `json:"config_prefix,omitempty"`
}

type UpdateJobRequest struct {
	Name        string            `json:"name,omitempty"`
	Type        string            `json:"type,omitempty"`
	Image       string            `json:"image,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Secrets     map[string]string `json:"secrets,omitempty"`
	Command     []string          `json:"command,omitempty"`
	Deletes     *Deletes          `json:"deletes,omitempty"`
	Oauth		*Oauth			  `json:"oauth,omitempty"`
}

type Deletes struct {
	Environment []string `json:"environment,omitempty"`
	Secrets     []string `json:"secrets,omitempty"`
}
