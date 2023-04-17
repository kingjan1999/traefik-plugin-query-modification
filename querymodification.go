package traefik_plugin_query_modification

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/url"
	"regexp"
)

type modificationType string

const (
	addType    modificationType = "add"
	modifyType modificationType = "modify"
	deleteType modificationType = "delete"
)

// Config is the configuration for this plugin
type Config struct {
	Type            modificationType `json:"type"`
	ParamName       string           `json:"paramName"`
	ParamNameRegex  string           `json:"paramNameRegex"`
	ParamValueRegex string           `json:"paramValueRegex"`
	NewValue        string           `json:"newValue"`
	NewValueRegex   string           `json:"newValueRegex"`
}

// CreateConfig creates a new configuration for this plugin
func CreateConfig() *Config {
	return &Config{}
}

// QueryModification represents the basic properties of this plugin
type QueryModification struct {
	next                    http.Handler
	name                    string
	config                  *Config
	paramNameRegexCompiled  *regexp.Regexp
	paramValueRegexCompiled *regexp.Regexp
}

// New creates a new instance of this plugin
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if !config.Type.isValid() {
		return nil, errors.New("invalid modification type, expected add / modify / delete")
	}

	if config.ParamNameRegex == "" && config.ParamName == "" && config.ParamValueRegex == "" {
		return nil, errors.New("either paramNameRegex or paramName or paramValueRegex must be set")
	}

	if config.ParamNameRegex != "" && containsNonEmpty(config.ParamName, config.ParamValueRegex) ||
		config.ParamName != "" && containsNonEmpty(config.ParamNameRegex, config.ParamValueRegex) ||
		config.ParamValueRegex != "" && containsNonEmpty(config.ParamName, config.ParamNameRegex) {
		log.Println("[Plugin Query Modification] It is discouraged to use multiple param matchers at once. Please proceed with caution")
	}

	if config.NewValueRegex != "" && config.ParamValueRegex == "" {
		return nil, errors.New("newValueRegex can only be used together with paramValueRegex")
	}

	var paramNameRegexCompiled *regexp.Regexp = nil
	if config.ParamNameRegex != "" {
		var err error
		paramNameRegexCompiled, err = regexp.Compile(config.ParamNameRegex)
		if err != nil {
			return nil, err
		}
	}

	var paramValueRegexCompiled *regexp.Regexp = nil
	if config.ParamValueRegex != "" {
		var err error
		paramValueRegexCompiled, err = regexp.Compile(config.ParamValueRegex)
		if err != nil {
			return nil, err
		}
	}

	return &QueryModification{
		next:                    next,
		name:                    name,
		config:                  config,
		paramNameRegexCompiled:  paramNameRegexCompiled,
		paramValueRegexCompiled: paramValueRegexCompiled,
	}, nil
}

func (q *QueryModification) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	v := url.Values{}
	v.Add(q.config.ParamName, q.config.NewValue)

	req.URL.RawQuery = v.Encode()
	req.RequestURI = req.URL.RequestURI()

	q.next.ServeHTTP(rw, req)
}

func determineAffectedParams(req *http.Request, q *QueryModification) []string {
	var result []string
	for key, values := range req.URL.Query() {
		if q.config.ParamName == key ||
			(q.paramNameRegexCompiled != nil && q.paramNameRegexCompiled.MatchString(key)) ||
			(q.paramValueRegexCompiled != nil && anyMatch(values, q.paramValueRegexCompiled)) {
			result = append(result, key)
		}
	}

	return result
}

func anyMatch(values []string, regex *regexp.Regexp) bool {
	for _, value := range values {
		if regex.MatchString(value) {
			return true
		}
	}
	return false
}

func (mt modificationType) isValid() bool {
	switch mt {
	case addType, modifyType, deleteType, "":
		return true
	}

	return false
}

func containsNonEmpty(ss ...string) bool {
	for _, s := range ss {
		if s != "" {
			return true
		}
	}
	return false
}
