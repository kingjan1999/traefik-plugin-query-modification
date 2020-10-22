package traefik_plugin_query_modification

import (
	"context"
	"errors"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type ModificationType string

const (
	Add    ModificationType = "add"
	Modify ModificationType = "modify"
	Delete ModificationType = "delete"
)

type Config struct {
	Type            ModificationType `json:"type"`
	ParamName       string           `json:"paramName"`
	ParamNameRegex  string           `json:"paramNameRegex"`
	ParamValueRegex string           `json:"paramValueRegex"`
	NewValue        string           `json:"newValue"`
	NewValueRegex   string           `json:"newValueRegex"`
}

func CreateConfig() *Config {
	return &Config{}
}

type QueryModification struct {
	next                    http.Handler
	name                    string
	config                  *Config
	paramNameRegexCompiled  *regexp.Regexp
	paramValueRegexCompiled *regexp.Regexp
	newValueRegexCompiled   *regexp.Regexp
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if !config.Type.isValid() {
		return nil, errors.New("invalid modification type, expected add / modify / delete")
	}

	if config.ParamNameRegex == "" && config.ParamName == "" && config.ParamValueRegex == "" {
		return nil, errors.New("either paramNameRegex or paramName or paramValueRegex must be set")
	}

	if config.ParamNameRegex != "" && !containsEmpty(config.ParamName, config.ParamValueRegex) ||
		config.ParamName != "" && !containsEmpty(config.ParamNameRegex, config.ParamValueRegex) ||
		config.ParamValueRegex != "" && !containsEmpty(config.ParamName, config.ParamNameRegex) {
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

	var newValueRegexCompiled *regexp.Regexp = nil
	if config.NewValueRegex != "" {
		var err error
		newValueRegexCompiled, err = regexp.Compile(config.NewValueRegex)
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
		newValueRegexCompiled:   newValueRegexCompiled,
	}, nil
}

func (q *QueryModification) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	qry := req.URL.Query()
	switch q.config.Type {
	case Add:
		qry.Add(q.config.ParamName, q.config.NewValue)
		break
	case Delete:
		paramsToDelete := determineAffectedParams(req, q)
		for _, paramToDelete := range paramsToDelete {
			qry.Del(paramToDelete)
		}
		break
	case Modify:
		paramsToModify := determineAffectedParams(req, q)
		for _, paramToModify := range paramsToModify {
			// use "old" query to prevent unwanted side effects
			oldValues := req.URL.Query()[paramToModify]
			var newValues []string
			for _, oldValue := range oldValues {
				var newValue string
				if q.paramValueRegexCompiled == nil || q.paramValueRegexCompiled.MatchString(oldValue) {
					if q.newValueRegexCompiled != nil {
						q.newValueRegexCompiled.
					} else {
						newValue = strings.ReplaceAll(q.config.NewValue, "$1", oldValue)
						newValues = append(newValues, newValue)
					}
				} else {
					//
					newValue = oldValue
				}
				newValues = append(newValues, newValue)
			}
			qry[paramToModify] = newValues
		}
	}

	req.URL.RawQuery = qry.Encode()

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

func (mt ModificationType) isValid() bool {
	switch mt {
	case Add, Modify, Delete, "":
		return true
	}

	return false
}

func containsEmpty(ss ...string) bool {
	for _, s := range ss {
		if s == "" {
			return true
		}
	}
	return false
}
