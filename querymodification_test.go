package traefik_plugin_query_modification_test

import (
	"context"
	traefik_plugin_query_modification "github.com/kingjan1999/traefik-plugin-query-modification"
	"net/http"
	"net/http/httptest"
	"testing"
)

// region Test Add

func TestAddQueryParam_NoPrevious(t *testing.T) {
	cfg := traefik_plugin_query_modification.CreateConfig()
	cfg.Type = "add"
	cfg.ParamName = "newparam"
	cfg.NewValue = "newvalue"
	expected := "newparam=newvalue"

	assertQueryModification(t, cfg, "", expected)
}

func TestAddQueryParam_OtherPrevious(t *testing.T) {
	cfg := traefik_plugin_query_modification.CreateConfig()
	cfg.Type = "add"
	cfg.ParamName = "newparam"
	cfg.NewValue = "newvalue"
	expected := "a=b&newparam=newvalue"
	previous := "a=b"

	assertQueryModification(t, cfg, previous, expected)
}

func TestAddQueryParam_AddPrevious(t *testing.T) {
	cfg := traefik_plugin_query_modification.CreateConfig()
	cfg.Type = "add"
	cfg.ParamName = "newparam"
	cfg.NewValue = "newvalue"
	expected := "newparam=oldvalue&newparam=newvalue"
	previous := "newparam=oldvalue"

	assertQueryModification(t, cfg, previous, expected)
}

func TestAddQueryParam_Previous(t *testing.T) {
	cfg := traefik_plugin_query_modification.CreateConfig()
	cfg.Type = "add"
	cfg.ParamName = "newparam"
	cfg.NewValue = "newvalue"
	expected := "a=b&newparam=newvalue"
	previous := "a=b"

	assertQueryModification(t, cfg, previous, expected)
}

// endregion

//region Delete
func TestDeleteQueryParam(t *testing.T) {
	cfg := traefik_plugin_query_modification.CreateConfig()
	cfg.Type = "delete"
	cfg.ParamName = "paramtodelete"
	expected := ""
	previous := "paramtodelete=anything"

	assertQueryModification(t, cfg, previous, expected)
}

func TestDeleteQueryParam_Multiple(t *testing.T) {
	cfg := traefik_plugin_query_modification.CreateConfig()
	cfg.Type = "delete"
	cfg.ParamName = "paramtodelete"
	expected := ""
	previous := "paramtodelete=anything&paramtodelete=somethingelse"

	assertQueryModification(t, cfg, previous, expected)
}

func TestDeleteQueryParam_NotFound(t *testing.T) {
	cfg := traefik_plugin_query_modification.CreateConfig()
	cfg.Type = "delete"
	cfg.ParamName = "paramtodelete"
	expected := "some=thing"
	previous := "some=thing"

	assertQueryModification(t, cfg, previous, expected)
}

func TestDeleteQueryParam_Others(t *testing.T) {
	cfg := traefik_plugin_query_modification.CreateConfig()
	cfg.Type = "delete"
	cfg.ParamName = "paramtodelete"
	expected := "otherparam=stillhere"
	previous := "otherparam=stillhere&paramtodelete=away"

	assertQueryModification(t, cfg, previous, expected)
}

//endregion

// region Modify
func TestModifyQueryParam_Simple(t *testing.T) {
	cfg := traefik_plugin_query_modification.CreateConfig()
	cfg.Type = "modify"
	cfg.ParamName = "a"
	cfg.NewValue = "c"
	previous := "a=b"
	expected := "a=c"

	assertQueryModification(t, cfg, previous, expected)
}

func TestModifyQueryParam_NotFound(t *testing.T) {
	cfg := traefik_plugin_query_modification.CreateConfig()
	cfg.Type = "modify"
	cfg.ParamName = "a"
	cfg.NewValue = "c"
	previous := "d=b"
	expected := "d=b"

	assertQueryModification(t, cfg, previous, expected)
}

func TestModifyQueryParam_SimpleMultiple(t *testing.T) {
	cfg := traefik_plugin_query_modification.CreateConfig()
	cfg.Type = "modify"
	cfg.ParamName = "a"
	cfg.NewValue = "c"
	previous := "a=b&a=d"
	expected := "a=c&a=c"

	assertQueryModification(t, cfg, previous, expected)
}

func TestModifyQueryParam_SimpleReplace(t *testing.T) {
	cfg := traefik_plugin_query_modification.CreateConfig()
	cfg.Type = "modify"
	cfg.ParamName = "a"
	cfg.NewValue = "c$1"
	previous := "a=b"
	expected := "a=cb"

	assertQueryModification(t, cfg, previous, expected)
}

func TestModifyQueryParam_SimpleReplaceMultiple(t *testing.T) {
	cfg := traefik_plugin_query_modification.CreateConfig()
	cfg.Type = "modify"
	cfg.ParamName = "a"
	cfg.NewValue = "c$1"
	previous := "a=b&a=d"
	expected := "a=cb&a=cd"

	assertQueryModification(t, cfg, previous, expected)
}

func TestModifyQueryParam_RegexKey(t *testing.T) {
	cfg := traefik_plugin_query_modification.CreateConfig()
	cfg.Type = "modify"
	cfg.ParamNameRegex = "^[abc]$"
	cfg.NewValue = "d"
	previous := "a=b&c=e&e=f"
	expected := "a=d&c=d&e=f"

	assertQueryModification(t, cfg, previous, expected)
}

func TestModifyQueryParam_RegexValue(t *testing.T) {
	cfg := traefik_plugin_query_modification.CreateConfig()
	cfg.Type = "modify"
	cfg.ParamValueRegex = "^secretpassword$"
	cfg.NewValue = "censored"
	previous := "a=secretpassword&b=secretpassword&c=somethingelse"
	expected := "a=censored&b=censored&c=somethingelse"

	assertQueryModification(t, cfg, previous, expected)
}

func TestModifyQueryParam_RegexValueRegexReplace(t *testing.T) {
	cfg := traefik_plugin_query_modification.CreateConfig()
	cfg.Type = "modify"
	cfg.ParamValueRegex = "^.*(p..sword)$"
	cfg.NewValueRegex = "no-$1"
	previous := "a=secretpassword&b=secretpassword&c=somethingelse"
	expected := "a=no-password&b=no-password&c=somethingelse"

	assertQueryModification(t, cfg, previous, expected)
}

func TestModifyQueryParam_RegexValueMultiple(t *testing.T) {
	cfg := traefik_plugin_query_modification.CreateConfig()
	cfg.Type = "modify"
	cfg.ParamValueRegex = "^secretpassword$"
	cfg.NewValue = "censored"
	previous := "a=secretpassword&a=somethingelse"
	expected := "a=censored&a=somethingelse"

	assertQueryModification(t, cfg, previous, expected)
}

// endregion

func TestErrorInvalidType(t *testing.T) {
	cfg := traefik_plugin_query_modification.CreateConfig()
	cfg.Type = "bla"
	cfg.ParamName = "blub"
	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
	_, err := traefik_plugin_query_modification.New(ctx, next, cfg, "query-modification-plugin")

	if err == nil {
		t.Error("expected error but err is nil")
	}
}

func TestErrorNoParam(t *testing.T) {
	cfg := traefik_plugin_query_modification.CreateConfig()
	cfg.Type = "delete"
	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
	_, err := traefik_plugin_query_modification.New(ctx, next, cfg, "query-modification-plugin")

	if err == nil {
		t.Error("expected error but err is nil")
	}
}

func createReqAndRecorder(cfg *traefik_plugin_query_modification.Config) (http.Handler, error, *httptest.ResponseRecorder, *http.Request) {
	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
	handler, err := traefik_plugin_query_modification.New(ctx, next, cfg, "query-modification-plugin")
	if err != nil {
		return nil, err, nil, nil
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	return handler, err, recorder, req
}

func assertQueryModification(t *testing.T, cfg *traefik_plugin_query_modification.Config, previous, expected string) {
	handler, err, recorder, req := createReqAndRecorder(cfg)
	if err != nil {
		t.Fatal(err)
		return
	}
	req.URL.RawQuery = previous
	handler.ServeHTTP(recorder, req)

	if req.URL.Query().Encode() != expected {
		t.Errorf("Expected %s, got %s", expected, req.URL.Query().Encode())
	}
}
