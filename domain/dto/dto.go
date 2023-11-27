package dto

type Alertmanager struct {
	Version           string                 `json:"version"`
	GroupKey          string                 `json:"groupKey"`
	TruncatedAlerts   int                    `json:"truncatedAlerts"`
	Status            string                 `json:"status"`
	Receiver          string                 `json:"receiver"`
	GroupLabels       map[string]interface{} `json:"groupLabels"`
	CommonLabels      map[string]interface{} `json:"commonLabels"`
	CommonAnnotations map[string]interface{} `json:"commonAnnotations"`
	ExternalURL       string                 `json:"externalURL"`
	Alerts            []AMAlert              `json:"alerts"`
}

type AMAlert struct {
	Status       string                 `json:"status"`
	Labels       map[string]interface{} `json:"labels"`
	Annotations  map[string]interface{} `json:"annotations"`
	StartsAt     string                 `json:"startsAt"`
	EndsAt       string                 `json:"endsAt"`
	GeneratorURL string                 `json:"generatorURL"`
}

type GrafanaAlert struct {
	DashboardId int                    `json:"dashboardId"`
	EvalMatches []GrafanaMatches       `json:"evalMatches"`
	ImageUrl    string                 `json:"imageUrl"`
	Message     string                 `json:"message"`
	OrgId       int                    `json:"orgId"`
	PanelId     int                    `json:"panelId"`
	RuleId      int                    `json:"ruleId"`
	RuleName    string                 `json:"ruleName"`
	RuleUrl     string                 `json:"ruleUrl"`
	State       string                 `json:"state"`
	Tags        map[string]interface{} `json:"tags"`
	Title       string                 `json:"title"`
}

type GrafanaMatches struct {
	Value  int                    `json:"value"`
	Metric string                 `json:"metric"`
	Tags   map[string]interface{} `json:"tags"`
}

type SignalMessage struct {
	Attachments []string `json:"base64_attachments"`
	Message     string   `json:"message"`
	Number      string   `json:"number"`
	Recipients  []string `json:"recipients"`
}
