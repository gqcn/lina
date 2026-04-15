// This file implements the host service demo business logic for the dynamic
// sample plugin.

package dynamicservice

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/pluginbridge"
)

const (
	hostCallDemoStateKey           = "host_call_demo_visit_count"
	hostCallDemoStoragePath        = "host-call-demo/"
	hostCallDemoStoragePrefix      = "host-call-demo"
	hostCallDemoStorageContentType = "application/json"
	hostCallDemoNetworkURL         = "https://example.com"
	hostCallDemoNetworkMethodGet   = "GET"
	hostCallDemoDataTable          = "sys_plugin_node_state"
	hostCallDemoDesiredState       = "running"
	hostCallDemoCurrentStateNew    = "pending"
	hostCallDemoCurrentStateReady  = "running"
	hostCallDemoAnonymousUser      = "anonymous"
	hostCallDemoSummaryMessage     = "Host service demo executed through runtime, storage, network, and data services."
	hostCallDemoNetworkPreview     = 120
)

// BuildHostCallDemoPayload executes the host service demo and returns the
// response payload.
func (s *Service) BuildHostCallDemoPayload(request *pluginbridge.BridgeRequestEnvelopeV1) (map[string]any, error) {
	username := hostCallDemoAnonymousUser
	if request.Identity != nil && request.Identity.Username != "" {
		username = request.Identity.Username
	}

	nowValue, err := s.runtimeSvc.Now()
	if err != nil {
		return nil, err
	}
	uuidValue, err := s.runtimeSvc.UUID()
	if err != nil {
		return nil, err
	}
	nodeValue, err := s.runtimeSvc.Node()
	if err != nil {
		return nil, err
	}
	if err = s.runtimeSvc.Log(int(pluginbridge.LogLevelInfo), "host service demo invoked", map[string]string{
		"username":  username,
		"requestId": request.RequestID,
		"route":     request.Route.InternalPath,
		"demoKey":   uuidValue,
	}); err != nil {
		return nil, err
	}

	visitCount, found, err := s.runtimeSvc.StateGetInt(hostCallDemoStateKey)
	if err != nil || !found {
		visitCount = 0
	}
	visitCount++
	_ = s.runtimeSvc.StateSetInt(hostCallDemoStateKey, visitCount)

	storageSummary, err := s.runHostCallDemoStorage(request.PluginID, uuidValue)
	if err != nil {
		return nil, err
	}
	dataSummary, err := s.runHostCallDemoData(request.PluginID, uuidValue)
	if err != nil {
		return nil, err
	}
	networkSummary := s.runHostCallDemoNetwork(request, uuidValue)

	return map[string]any{
		"visitCount": visitCount,
		"pluginId":   request.PluginID,
		"runtime": map[string]any{
			"now":  nowValue,
			"uuid": uuidValue,
			"node": nodeValue,
		},
		"storage": storageSummary,
		"network": networkSummary,
		"data":    dataSummary,
		"message": hostCallDemoSummaryMessage,
	}, nil
}

func (s *Service) runHostCallDemoStorage(pluginID string, demoKey string) (map[string]any, error) {
	objectPath := fmt.Sprintf("%s/%s.json", hostCallDemoStoragePrefix, demoKey)
	body, err := json.Marshal(map[string]string{
		"pluginId": pluginID,
		"demoKey":  demoKey,
	})
	if err != nil {
		return nil, gerror.Wrap(err, "marshal storage demo request body failed")
	}
	if _, err = s.storageSvc.Put(objectPath, body, hostCallDemoStorageContentType, true); err != nil {
		return nil, err
	}
	deleted := false
	defer func() {
		if !deleted {
			_ = s.storageSvc.Delete(objectPath)
		}
	}()

	readBody, _, found, err := s.storageSvc.Get(objectPath)
	if err != nil {
		return nil, err
	}
	if !found || string(readBody) != string(body) {
		return nil, gerror.New("storage demo object verification failed")
	}

	objects, err := s.storageSvc.List(hostCallDemoStoragePrefix, 10)
	if err != nil {
		return nil, err
	}
	if err = s.storageSvc.Delete(objectPath); err != nil {
		return nil, err
	}
	deleted = true

	_, statFound, err := s.storageSvc.Stat(objectPath)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"pathPrefix":  hostCallDemoStoragePath,
		"objectPath":  objectPath,
		"stored":      true,
		"listedCount": len(objects),
		"deleted":     !statFound,
	}, nil
}

func (s *Service) runHostCallDemoData(pluginID string, demoKey string) (map[string]any, error) {
	createResult, err := s.dataSvc.Table(hostCallDemoDataTable).Insert(map[string]any{
		"pluginId":     pluginID,
		"releaseId":    0,
		"nodeKey":      "host-call-demo-" + demoKey,
		"desiredState": hostCallDemoDesiredState,
		"currentState": hostCallDemoCurrentStateNew,
		"generation":   1,
		"errorMessage": "",
	})
	if err != nil {
		return nil, err
	}
	if createResult == nil || createResult.Key == nil {
		return nil, gerror.New("data demo create did not return a record key")
	}

	recordKey := createResult.Key
	deleted := false
	defer func() {
		if !deleted {
			_, _ = s.dataSvc.Table(hostCallDemoDataTable).WhereKey(recordKey).Delete()
		}
	}()

	listRecords, listTotal, err := s.dataSvc.Table(hostCallDemoDataTable).
		Fields("id", "nodeKey", "currentState").
		WhereEq("pluginId", pluginID).
		WhereLike("nodeKey", demoKey).
		WhereIn("currentState", []string{hostCallDemoCurrentStateNew, hostCallDemoCurrentStateReady}).
		OrderDesc("id").
		Page(1, 10).
		All()
	if err != nil {
		return nil, err
	}
	if listTotal < 1 || len(listRecords) == 0 {
		return nil, gerror.New("data demo list did not find the created record")
	}
	countTotal, err := s.dataSvc.Table(hostCallDemoDataTable).
		WhereEq("pluginId", pluginID).
		WhereLike("nodeKey", demoKey).
		Count()
	if err != nil {
		return nil, err
	}
	recordKey = listRecords[0]["id"]

	if _, err = s.dataSvc.Table(hostCallDemoDataTable).WhereKey(recordKey).Update(map[string]any{
		"currentState": hostCallDemoCurrentStateReady,
		"errorMessage": "",
	}); err != nil {
		return nil, err
	}

	record, found, err := s.dataSvc.Table(hostCallDemoDataTable).Fields("currentState").WhereKey(recordKey).One()
	if err != nil {
		return nil, err
	}
	if !found || record == nil || fmt.Sprint(record["currentState"]) != hostCallDemoCurrentStateReady {
		return nil, gerror.New("data demo get did not return the updated record")
	}

	if _, err = s.dataSvc.Table(hostCallDemoDataTable).WhereKey(recordKey).Delete(); err != nil {
		return nil, err
	}
	deleted = true

	return map[string]any{
		"table":      hostCallDemoDataTable,
		"recordKey":  fmt.Sprint(recordKey),
		"listTotal":  int(listTotal),
		"countTotal": int(countTotal),
		"updated":    true,
		"deleted":    true,
	}, nil
}

func (s *Service) runHostCallDemoNetwork(request *pluginbridge.BridgeRequestEnvelopeV1, demoKey string) map[string]any {
	result := map[string]any{
		"url":         hostCallDemoNetworkURL,
		"skipped":     false,
		"statusCode":  0,
		"contentType": "",
		"bodyPreview": "",
		"error":       "",
	}
	if hasHostCallDemoFlag(request, "skipNetwork") {
		result["skipped"] = true
		return result
	}

	response, err := s.httpSvc.Request(hostCallDemoNetworkURL, &pluginbridge.HostServiceNetworkRequest{
		Method: hostCallDemoNetworkMethodGet,
		Headers: map[string]string{
			"x-request-id": request.RequestID + "-" + demoKey,
		},
	})
	if err != nil {
		result["error"] = err.Error()
		return result
	}
	result["statusCode"] = int(response.StatusCode)
	result["contentType"] = response.ContentType
	result["bodyPreview"] = buildHostCallDemoBodyPreview(response.Body)
	return result
}

func hasHostCallDemoFlag(request *pluginbridge.BridgeRequestEnvelopeV1, key string) bool {
	if request == nil || request.Route == nil || len(request.Route.QueryValues) == 0 {
		return false
	}
	values := request.Route.QueryValues[key]
	for _, value := range values {
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "1", "true", "yes", "on":
			return true
		}
	}
	return false
}

func buildHostCallDemoBodyPreview(body []byte) string {
	preview := strings.TrimSpace(string(body))
	if preview == "" {
		return ""
	}
	if len(preview) <= hostCallDemoNetworkPreview {
		return preview
	}
	return preview[:hostCallDemoNetworkPreview]
}
