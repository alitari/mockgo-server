package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/alitari/mockgo-server/internal/kvstore"
	"github.com/alitari/mockgo-server/internal/mock"
	"github.com/alitari/mockgo-server/internal/model"
	"github.com/alitari/mockgo-server/internal/utils"
	"github.com/google/uuid"

	"github.com/go-http-utils/headers"
	"github.com/gorilla/mux"
)

const NoAdvertiseHeader = "No-advertise"

type ClusterSetup int64

const (
	Disabled ClusterSetup = iota
	Static
	Dynamic
)

func NewClusterSetup(clusterUrls []string) ClusterSetup {
	if len(clusterUrls) == 0 {
		return Disabled
	} else {
		if len(clusterUrls) == 1 && clusterUrls[0] == "dynamic" {
			return Dynamic
		} else {
			return Static
		}
	}
}

func (c ClusterSetup) String() string {
	switch c {
	case Disabled:
		return "disabled"
	case Static:
		return "static"
	case Dynamic:
		return "dynamic"
	}
	return "unknown"
}

type ConfigRouter struct {
	mockRouter           *mock.MockRouter
	router               *mux.Router
	server               *http.Server
	port                 int
	id                   string
	clusterSetup         ClusterSetup
	clusterUrls          []string
	clusterPodLabelValue string
	logger               *utils.Logger
	kvstore              *kvstore.KVStore
	basicAuthUsername    string
	basicAuthPassword    string
	transferringMatches  bool
	httpClient           http.Client
}

type EndpointsResponse struct {
	Endpoints []*model.MockEndpoint
}

type AddKVStoreRequest struct {
	Path  string `json:"path"`
	Value string `json:"value"`
}

type RemoveKVStoreRequest struct {
	Path string `json:"path"`
}

func NewConfigRouter(username, password string, mockRouter *mock.MockRouter, port int, clusterUrls []string, clusterPodLabelValue string, kvstore *kvstore.KVStore, httpClientTimeout time.Duration, logger *utils.Logger) *ConfigRouter {
	configRouter := &ConfigRouter{
		mockRouter:           mockRouter,
		port:                 port,
		clusterSetup:         NewClusterSetup(clusterUrls),
		clusterUrls:          clusterUrls,
		clusterPodLabelValue: clusterPodLabelValue,
		id:                   uuid.New().String(),
		logger:               logger,
		kvstore:              kvstore,
		httpClient:           utils.CreateHttpClient(httpClientTimeout),
		basicAuthUsername:    username,
		basicAuthPassword:    password,
		transferringMatches:  false,
	}

	//time.Sleep(300 * time.Millisecond)
	configRouter.newRouter()
	return configRouter
}

func (r *ConfigRouter) Name() string {
	return "Configrouter"
}

func (r *ConfigRouter) Router() *mux.Router {
	return r.router
}

func (r *ConfigRouter) Server() *http.Server {
	return r.server
}

func (r *ConfigRouter) Port() int {
	return r.port
}

func (r *ConfigRouter) Logger() *utils.Logger {
	return r.logger
}

func (r *ConfigRouter) newRouter() {
	router := mux.NewRouter()
	router.NewRoute().Name("health").Path("/health").Methods(http.MethodGet).HandlerFunc(utils.RequestMustHave("", "", http.MethodGet, "", "", nil, r.health))
	router.NewRoute().Name("serverId").Path("/id").Methods(http.MethodGet).HandlerFunc(utils.RequestMustHave(r.basicAuthUsername, r.basicAuthPassword, http.MethodGet, "", "application/text", nil, r.serverId))
	router.NewRoute().Name("setKVStore").Path("/kvstore/{key}").Methods(http.MethodPut).HandlerFunc(utils.RequestMustHave(r.basicAuthUsername, r.basicAuthPassword, http.MethodPut, "application/json", "", []string{"key"}, r.setKVStore))
	router.NewRoute().Name("getKVStore").Path("/kvstore/{key}").Methods(http.MethodGet).HandlerFunc(utils.RequestMustHave(r.basicAuthUsername, r.basicAuthPassword, http.MethodGet, "", "application/json", []string{"key"}, r.getKVStore))
	router.NewRoute().Name("addKVStore").Path("/kvstore/{key}/add").Methods(http.MethodPost).HandlerFunc(utils.RequestMustHave(r.basicAuthUsername, r.basicAuthPassword, http.MethodPost, "application/json", "", []string{"key"}, r.addKVStore))
	router.NewRoute().Name("removeKVStore").Path("/kvstore/{key}/remove").Methods(http.MethodPost).HandlerFunc(utils.RequestMustHave(r.basicAuthUsername, r.basicAuthPassword, http.MethodPost, "application/json", "", []string{"key"}, r.removeKVStore))
	router.NewRoute().Name("uploadKVStore").Path("/kvstore").Methods(http.MethodPut).HandlerFunc(utils.RequestMustHave(r.basicAuthUsername, r.basicAuthPassword, http.MethodPut, "application/json", "", nil, r.uploadKVStore))
	router.NewRoute().Name("downloadKVStore").Path("/kvstore").Methods(http.MethodGet).HandlerFunc(utils.RequestMustHave(r.basicAuthUsername, r.basicAuthPassword, http.MethodGet, "", "application/json", nil, r.downloadKVStore))
	router.NewRoute().Name("getMatches").Path("/matches").Methods(http.MethodGet).HandlerFunc(utils.RequestMustHave(r.basicAuthUsername, r.basicAuthPassword, http.MethodGet, "", "application/json", nil, r.getMatchesFromAll))
	router.NewRoute().Name("getMismatches").Path("/mismatches").Methods(http.MethodGet).HandlerFunc(utils.RequestMustHave(r.basicAuthUsername, r.basicAuthPassword, http.MethodGet, "", "application/json", nil, r.getMismatchesFromAll))
	router.NewRoute().Name("addMatches").Path("/addmatches").Methods(http.MethodPost).HandlerFunc(utils.RequestMustHave(r.basicAuthUsername, r.basicAuthPassword, http.MethodPost, "application/json", "", nil, r.addMatches))
	router.NewRoute().Name("addMismatches").Path("/addmismatches").Methods(http.MethodPost).HandlerFunc(utils.RequestMustHave(r.basicAuthUsername, r.basicAuthPassword, http.MethodPost, "application/json", "", nil, r.addMismatches))
	router.NewRoute().Name("deleteMatches").Path("/matches").Methods(http.MethodDelete).HandlerFunc(utils.RequestMustHave(r.basicAuthUsername, r.basicAuthPassword, http.MethodDelete, "", "", nil, r.deleteMatchesFromAll))
	router.NewRoute().Name("deleteMismatches").Path("/mismatches").Methods(http.MethodDelete).HandlerFunc(utils.RequestMustHave(r.basicAuthUsername, r.basicAuthPassword, http.MethodDelete, "", "", nil, r.deleteMismatchesFromAll))
	router.NewRoute().Name("transferMatches").Path("/transfermatches").Methods(http.MethodGet).HandlerFunc(utils.RequestMustHave("", "", http.MethodGet, "", "", nil, r.transferMatches))
	r.router = router
	r.server = &http.Server{Addr: ":" + strconv.Itoa(r.port), Handler: router}
}

func (r *ConfigRouter) health(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
}

func (r *ConfigRouter) getClusterUrls() []string {
	if r.clusterSetup == Dynamic {
		kupeProxyPort, err := strconv.Atoi(os.Getenv("KUBEPROXY_PORT"))
		if err != nil {
			log.Fatalf("Can't find KUBEPROXY_PORT : %v", err)
		}
		podIps, err := utils.K8sGetPodsIps(r.httpClient, kupeProxyPort, r.clusterPodLabelValue)
		if err != nil {
			log.Fatalf("Can't get k8s pods : %v", err)
		}
		var clusterUrls []string
		for _, ip := range podIps {
			clusterUrl := fmt.Sprintf("http://%s:%d%s", ip, r.mockRouter.Port(), r.mockRouter.ProxyConfigRouterPath)
			clusterUrls = append(clusterUrls, clusterUrl)
		}
		return clusterUrls
	} else {
		return r.clusterUrls
	}
}

func (r *ConfigRouter) clusterRequest(method, path string, header map[string]string, body string, expectedStatusCode int, continueOnError bool, responseHandler func(clusterUrl, responseBody string) (bool, error)) error {
	clusterUrls := r.getClusterUrls()
	for _, clusterUrl := range clusterUrls {
		clusterRequest, err := http.NewRequest(method, clusterUrl+path, bytes.NewBufferString(body))
		if err != nil {
			r.logger.LogWhenVerbose(fmt.Sprintf("can't create request, error: %v", err))
			if continueOnError {
				continue
			} else {
				return err
			}
		}
		for k, v := range header {
			clusterRequest.Header.Add(k, v)
		}
		clusterRequest.Header.Add(headers.Authorization, utils.BasicAuth(r.basicAuthUsername, r.basicAuthPassword))
		r.logger.LogWhenVerbose(fmt.Sprintf("Requesting %s %s to url %s ...", clusterRequest.Method, clusterRequest.URL.String(), clusterUrl))
		clusterResponse, err := r.httpClient.Do(clusterRequest)
		if err != nil {
			r.logger.LogWhenVerbose(fmt.Sprintf("cluster node '%s' can't process request, answered with error: %v", clusterUrl, err))
			if continueOnError {
				continue
			} else {
				return err
			}
		}
		bodyBytes, err := ioutil.ReadAll(clusterResponse.Body)
		if clusterResponse.StatusCode != expectedStatusCode {
			mess := fmt.Sprintf("cluster node '%s' can't process request, answered with status: %v body: '%s'", clusterUrl, clusterResponse.StatusCode, bodyBytes)
			r.logger.LogWhenVerbose(mess)
			if continueOnError {
				continue
			} else {
				return errors.New(mess)
			}
		}
		if err != nil {
			r.logger.LogAlways(fmt.Sprintf("(ERROR) reading response from cluster node '%s' failed: %v ", clusterUrl, err))
			return err
		}
		stop, err := responseHandler(clusterUrl, string(bodyBytes))
		clusterResponse.Body.Close()
		if stop {
			return err
		}
	}
	return nil
}

func (r *ConfigRouter) DownloadKVStoreFromCluster() error {
	return r.clusterRequest(http.MethodGet, "/kvstore", map[string]string{headers.Accept: `application/json`}, "", http.StatusOK, true,
		func(clusterUrl, responseBody string) (bool, error) {
			err := r.kvstore.PutAllJson(responseBody)
			if err != nil {
				r.logger.LogAlways(fmt.Sprintf("(ERROR) creating new kvstore downloaded from clusterurl '%s' failed: %v ", clusterUrl, err))
				return true, err
			}
			r.logger.LogWhenVerbose(fmt.Sprintf("syncing completed, kvstore successfully downloaded from clusterurl '%s' ", clusterUrl))
			return true, nil
		})
}

func (r *ConfigRouter) setKVStoreToCluster(key, value string) error {
	return r.clusterRequest(http.MethodPut, "/kvstore/"+key, map[string]string{NoAdvertiseHeader: "true", headers.ContentType: "application/json"}, value, http.StatusNoContent, false,
		func(clusterUrl, responseBody string) (bool, error) {
			r.logger.LogWhenVerbose(fmt.Sprintf("successfully set kvstore '%s' to cluster url '%s' !", key, clusterUrl))
			return false, nil
		})
}

func (r *ConfigRouter) addKVStoreToCluster(key, path, value string) error {
	body := fmt.Sprintf(`{ "path":"%s", "value":"%s" }`, path, value)
	return r.clusterRequest(http.MethodPost, "/kvstore/"+key+"/add", map[string]string{NoAdvertiseHeader: "true", headers.ContentType: "application/json"}, body, http.StatusNoContent, false,
		func(clusterUrl, responseBody string) (bool, error) {
			r.logger.LogWhenVerbose(fmt.Sprintf("kvstore '%s': successfully added path: '%s' and value: '%s'  to  cluster url '%s' !", key, path, value, clusterUrl))
			return false, nil
		})
}

func (r *ConfigRouter) removeKVStoreToCluster(key, path string) error {
	body := fmt.Sprintf(`{ "path":"%s" }`, path)
	return r.clusterRequest(http.MethodPost, "/kvstore/"+key+"/add", map[string]string{NoAdvertiseHeader: "true", headers.ContentType: "application/json"}, body, http.StatusNoContent, false,
		func(clusterUrl, responseBody string) (bool, error) {
			r.logger.LogWhenVerbose(fmt.Sprintf("kvstore '%s': successfully removed path: '%s' to cluster url '%s' !", key, path, clusterUrl))
			return false, nil
		})
}

func (r *ConfigRouter) serverId(writer http.ResponseWriter, request *http.Request) {
	_, err := io.WriteString(writer, r.id)
	if err != nil {
		http.Error(writer, fmt.Sprintf("Cannot write response: %v", err), http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusOK)
}

func (r *ConfigRouter) deleteMatches() {
	r.mockRouter.Matches = make(map[string][]*model.Match)
	r.mockRouter.MatchesCount = make(map[string]int64)
}

func (r *ConfigRouter) deleteMismatches() {
	r.mockRouter.Mismatches = []*model.Mismatch{}
	r.mockRouter.MismatchesCount = 0
}

func (r *ConfigRouter) downloadKVStore(writer http.ResponseWriter, request *http.Request) {
	utils.WriteEntity(writer, r.kvstore.GetAll())
}

func (r *ConfigRouter) uploadKVStore(writer http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "Problem reading request body: "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = r.kvstore.PutAllJson(string(body))
	if err != nil {
		http.Error(writer, "Problem creating kvstore: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (r *ConfigRouter) getKVStore(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	key := vars["key"]
	val := r.kvstore.Get(key)
	utils.WriteEntity(writer, val)
}

func (r *ConfigRouter) getMatchesFromAll(writer http.ResponseWriter, request *http.Request) {
	if r.clusterSetup == Disabled || request.Header.Get(NoAdvertiseHeader) == "true" {
		if r.mockRouter.MatchesCountOnly {
			utils.WriteEntity(writer, r.mockRouter.MatchesCount)
		} else {
			utils.WriteEntity(writer, r.mockRouter.Matches)
		}
	} else {
		allMatches := make(map[string][]*model.Match)
		allMatchesCount := make(map[string]int64)
		err := r.clusterRequest(http.MethodGet, "/matches", map[string]string{NoAdvertiseHeader: "true", headers.Accept: "application/json"}, "", http.StatusOK, false,
			func(clusterUrl, responseBody string) (bool, error) {
				if r.mockRouter.MatchesCountOnly {
					var bodyData map[string]int64
					err := json.Unmarshal([]byte(responseBody), &bodyData)
					if err != nil {
						return true, err
					}
					for k, v := range bodyData {
						allMatchesCount[k] = allMatchesCount[k] + int64(v)
					}
					return false, nil
				} else {
					var bodyData map[string][]*model.Match
					err := json.Unmarshal([]byte(responseBody), &bodyData)
					if err != nil {
						return true, err
					}
					for k, v := range bodyData {
						allMatches[k] = append(allMatches[k], v...)
					}
					return false, nil
				}
			})
		if err != nil {
			http.Error(writer, "Problem getting matches: "+err.Error(), http.StatusInternalServerError)
		} else {
			if r.mockRouter.MatchesCountOnly {
				utils.WriteEntity(writer, allMatchesCount)
			} else {
				utils.WriteEntity(writer, allMatches)
			}
		}
	}
}

func (r *ConfigRouter) getMismatchesFromAll(writer http.ResponseWriter, request *http.Request) {
	if r.clusterSetup == Disabled || request.Header.Get(NoAdvertiseHeader) == "true" {
		if r.mockRouter.MismatchesCountOnly {
			utils.WriteEntity(writer, r.mockRouter.MismatchesCount)
		} else {
			utils.WriteEntity(writer, r.mockRouter.Mismatches)
		}
	} else {
		allMismatches := []*model.Mismatch{}
		allMismatchesCount := int64(0)
		err := r.clusterRequest(http.MethodGet, "/mismatches", map[string]string{NoAdvertiseHeader: "true", headers.Accept: "application/json"}, "", http.StatusOK, false,
			func(clusterUrl, responseBody string) (bool, error) {
				if r.mockRouter.MismatchesCountOnly {
					var bodyData int64
					err := json.Unmarshal([]byte(responseBody), &bodyData)
					if err != nil {
						return true, err
					}
					allMismatchesCount = allMismatchesCount + bodyData
					return false, nil
				} else {
					var bodyData []*model.Mismatch
					err := json.Unmarshal([]byte(responseBody), &bodyData)
					if err != nil {
						return true, err
					}
					allMismatches = append(allMismatches, bodyData...)
					return false, nil
				}
			})
		if err != nil {
			http.Error(writer, "Problem getting mismatches: "+err.Error(), http.StatusInternalServerError)
		} else {
			if r.mockRouter.MismatchesCountOnly {
				utils.WriteEntity(writer, allMismatchesCount)
			} else {
				utils.WriteEntity(writer, allMismatches)
			}
		}
	}
}

func (r *ConfigRouter) deleteMatchesFromAll(writer http.ResponseWriter, request *http.Request) {
	if r.clusterSetup == Disabled || request.Header.Get(NoAdvertiseHeader) == "true" {
		r.deleteMatches()
	} else {
		err := r.clusterRequest(http.MethodDelete, "/matches", map[string]string{NoAdvertiseHeader: "true"}, "", http.StatusOK, false,
			func(clusterUrl, responseBody string) (bool, error) {
				return false, nil
			})
		if err != nil {
			http.Error(writer, "Problem deleting matches: "+err.Error(), http.StatusInternalServerError)
		} else {
			writer.WriteHeader(http.StatusOK)
		}
	}
}

func (r *ConfigRouter) deleteMismatchesFromAll(writer http.ResponseWriter, request *http.Request) {
	if r.clusterSetup == Disabled || request.Header.Get(NoAdvertiseHeader) == "true" {
		r.deleteMismatches()
	} else {
		err := r.clusterRequest(http.MethodDelete, "/mismatches", map[string]string{NoAdvertiseHeader: "true"}, "", http.StatusOK, false,
			func(clusterUrl, responseBody string) (bool, error) {
				return false, nil
			})
		if err != nil {
			http.Error(writer, "Problem deleting mismatches: "+err.Error(), http.StatusInternalServerError)
		} else {
			writer.WriteHeader(http.StatusOK)
		}
	}
}

func (r *ConfigRouter) addMatches(writer http.ResponseWriter, request *http.Request) {
	r.logger.LogWhenVerbose("adding matches...")
	if r.transferringMatches {
		http.Error(writer, "Already transferring", http.StatusLocked)
		return
	}
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "Problem reading request body: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if r.mockRouter.MatchesCountOnly {
		var matchData map[string]int64
		err = json.Unmarshal(body, &matchData)
		if err != nil {
			http.Error(writer, "Problem marshalling matches response body: "+err.Error(), http.StatusInternalServerError)
			return
		}
		for k, v := range matchData {
			r.mockRouter.MatchesCount[k] = r.mockRouter.MatchesCount[k] + v
		}
		r.logger.LogWhenVerbose(fmt.Sprintf("added matchesCount from %d endpoints sucessfully", len(matchData)))

	} else {
		var matchData map[string][]*model.Match
		err = json.Unmarshal(body, &matchData)
		if err != nil {
			http.Error(writer, "Problem marshalling matches response body: "+err.Error(), http.StatusInternalServerError)
			return
		}
		for k, v := range matchData {
			r.mockRouter.Matches[k] = append(r.mockRouter.Matches[k], v...)
			r.mockRouter.MatchesCount[k] = r.mockRouter.MatchesCount[k] + int64(len(v))
		}
		r.logger.LogWhenVerbose(fmt.Sprintf("added matches from %d endpoints sucessfully", len(matchData)))
	}
	writer.WriteHeader(http.StatusOK)
}

func (r *ConfigRouter) addMismatches(writer http.ResponseWriter, request *http.Request) {
	r.logger.LogWhenVerbose("adding mismatches...")
	if r.transferringMatches {
		http.Error(writer, "Already transferring", http.StatusLocked)
		return
	}
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "Problem reading request body: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if r.mockRouter.MismatchesCountOnly {
		var matchData int64
		err = json.Unmarshal(body, &matchData)
		if err != nil {
			http.Error(writer, "Problem marshalling mismatches response body: "+err.Error(), http.StatusInternalServerError)
			return
		}
		r.mockRouter.MismatchesCount = r.mockRouter.MismatchesCount + matchData
		r.logger.LogWhenVerbose(fmt.Sprintf("added %d mismatchesCount sucessfully", matchData))
	} else {
		var matchData []*model.Mismatch
		err = json.Unmarshal(body, &matchData)
		if err != nil {
			http.Error(writer, "Problem marshalling mismatches response body: "+err.Error(), http.StatusInternalServerError)
			return
		}
		r.mockRouter.Mismatches = append(r.mockRouter.Mismatches, matchData...)
		r.mockRouter.MismatchesCount = r.mockRouter.MismatchesCount + int64(len(matchData))

		r.logger.LogWhenVerbose(fmt.Sprintf("added %d mismatches sucessfully", len(matchData)))
	}
	writer.WriteHeader(http.StatusOK)
}

func (r *ConfigRouter) transferMatches(writer http.ResponseWriter, request *http.Request) {
	r.transferringMatches = true
	defer func() {
		r.transferringMatches = false
	}()
	var matches []byte
	var err error
	if r.mockRouter.MatchesCountOnly {
		matches, err = json.Marshal(r.mockRouter.MatchesCount)
	} else {
		matches, err = json.Marshal(r.mockRouter.Matches)
	}
	if err != nil {
		http.Error(writer, "Problem marshalling matches: "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = r.clusterRequest(http.MethodPost, "/addmatches", map[string]string{headers.ContentType: `application/json`}, string(matches), http.StatusOK, true,
		func(clusterUrl, responseBody string) (bool, error) {
			r.logger.LogWhenVerbose(fmt.Sprintf("matches of %d endpoints successfully transferred to: %s", len(r.mockRouter.MatchesCount), clusterUrl))
			r.deleteMatches()
			return true, nil
		})
	if err != nil {
		http.Error(writer, "Problem transferring matches: "+err.Error(), http.StatusInternalServerError)
	} else {
		var mismatches []byte
		var err error
		if r.mockRouter.MismatchesCountOnly {
			mismatches, err = json.Marshal(r.mockRouter.MismatchesCount)
		} else {
			mismatches, err = json.Marshal(r.mockRouter.Mismatches)
		}
		if err != nil {
			http.Error(writer, "Problem marshalling mismatches: "+err.Error(), http.StatusInternalServerError)
			return
		}
		err = r.clusterRequest(http.MethodPost, "/addmismatches", map[string]string{headers.ContentType: `application/json`}, string(mismatches), http.StatusOK, true,
			func(clusterUrl, responseBody string) (bool, error) {
				r.logger.LogWhenVerbose(fmt.Sprintf("%d mismatches successfully transferred to: %s", r.mockRouter.MismatchesCount, clusterUrl))
				r.deleteMismatches()
				return true, nil
			})
		if err != nil {
			http.Error(writer, "Problem transferring mismatches: "+err.Error(), http.StatusInternalServerError)
		} else {
			writer.WriteHeader(http.StatusOK)
		}
	}
}

func (r *ConfigRouter) setKVStore(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	key := vars["key"]
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "Problem reading request body: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if r.clusterSetup == Disabled || request.Header.Get(NoAdvertiseHeader) == "true" {
		err = r.kvstore.PutAsJson(key, string(body))
		if err != nil {
			http.Error(writer, "Problem with kvstore value, ( is it valid JSON?): "+err.Error(), http.StatusBadRequest)
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	} else {
		err := r.setKVStoreToCluster(key, string(body))
		if err != nil {
			http.Error(writer, "Problem advertising kvstore value : "+err.Error(), http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	}
}

func (r *ConfigRouter) addKVStore(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	key := vars["key"]
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "Problem reading request body: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var addKvStoreRequest AddKVStoreRequest
	err = json.Unmarshal(body, &addKvStoreRequest)
	if err != nil {
		http.Error(writer, fmt.Sprintf("Can't parse request body '%s' : %v", body, err), http.StatusBadRequest)
		return
	}
	if r.clusterSetup == Disabled || request.Header.Get(NoAdvertiseHeader) == "true" {
		err = r.kvstore.PatchAdd(key, addKvStoreRequest.Path, addKvStoreRequest.Value)
		if err != nil {
			http.Error(writer, fmt.Sprintf("Problem adding kvstore path: '%s' value: '%s', : %v ", addKvStoreRequest.Path, addKvStoreRequest.Value, err), http.StatusBadRequest)
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	} else {
		err := r.addKVStoreToCluster(key, addKvStoreRequest.Path, addKvStoreRequest.Value)
		if err != nil {
			http.Error(writer, "Problem adding kvstore value : "+err.Error(), http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	}
}

func (r *ConfigRouter) removeKVStore(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	key := vars["key"]
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "Problem reading request body: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var removeKvStoreRequest RemoveKVStoreRequest
	err = json.Unmarshal(body, &removeKvStoreRequest)
	if err != nil {
		http.Error(writer, "Can't parse request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if r.clusterSetup == Disabled || request.Header.Get(NoAdvertiseHeader) == "true" {
		err = r.kvstore.PatchRemove(key, removeKvStoreRequest.Path)
		if err != nil {
			http.Error(writer, fmt.Sprintf("Problem removing kvstore '%s', path: '%s' : %v ", key, removeKvStoreRequest.Path, err), http.StatusBadRequest)
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	} else {
		err := r.removeKVStoreToCluster(key, removeKvStoreRequest.Path)
		if err != nil {
			http.Error(writer, fmt.Sprintf("Problem removing kvstore '%s', path: '%s' : %v ", key, removeKvStoreRequest.Path, err), http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	}
}
