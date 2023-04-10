package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/alitari/mockgo-server/mockgo/logging"
	mockgo "github.com/alitari/mockgo-server/mockgo/mock"
	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func main() {
	logLevelInt := flag.Int("logLevel", 0, "logLevel (0=debug, 1=verbose, 2=error)")
	help := flag.Bool("help", false, "show help")
	impl := flag.Bool("impl", false, "try to implement the mock")
	outputFile := flag.String("o", "", "output file")
	flag.Parse()
	if *help {
		fmt.Println("mockgo-gen [options] [openapi-file]")
		flag.PrintDefaults()
		os.Exit(0)
	}
	logger := logging.NewLoggerUtil(logging.ParseLogLevel(*logLevelInt))
	logger.LogWhenDebug("starting mockgo-gen with debug logging")

	loader := openapi3.NewLoader()
	var err error
	var apiSpec *openapi3.T
	if openApiFileArg := flag.Arg(0); openApiFileArg != "" {
		// check if openFileArg is an url
		if strings.HasPrefix(openApiFileArg, "http") {
			logger.LogWhenDebug(fmt.Sprintf("reading json from openAPIFileURL: %s", openApiFileArg))
			openAPIFileURL, err := url.Parse(openApiFileArg)
			if err != nil {
				logger.LogError("Error parsing OpenAPI spec URL", err)
				os.Exit(1)
			}
			apiSpec, err = loader.LoadFromURI(openAPIFileURL)
			if err != nil {
				logger.LogError("Error loading OpenAPI spec", err)
				os.Exit(1)
			}
			logger.LogWhenDebug(fmt.Sprintf("OpenAPI spec version: %s, title: %s", apiSpec.OpenAPI, apiSpec.Info.Title))
		} else {
			apiSpec, err = loader.LoadFromFile(openApiFileArg)
			if err != nil {
				logger.LogError("Error loading OpenAPI spec", err)
				os.Exit(1)
			}
		}
	} else {
		logger.LogWhenDebug("no file given, reading json from stdin")
		data := make([]byte, 100000)
		os.Stdin.Read(data)
		apiSpec, err = loader.LoadFromData(data)
		if err != nil {
			logger.LogError("Error loading OpenAPI spec", err)
			os.Exit(1)
		}
	}
	logger.LogWhenVerbose(fmt.Sprintf("OpenAPI spec version: %s, title: %s", apiSpec.OpenAPI, apiSpec.Info.Title))
	mockName := apiSpec.Info.Title
	if mockName == "" {
		mockName = "mock OpenAPI"
	}
	mock := &mockgo.Mock{Name: mockName}
	counter := 0

	for path, pathItem := range apiSpec.Paths {
		logger.LogWhenDebug(fmt.Sprintf("path: %s", path))
		if len(pathItem.Operations()) == 0 {
			logger.LogWhenDebug("  no operations defined, skipping")
			continue
		}
		for op, opItem := range pathItem.Operations() {
			counter++
			if op == http.MethodGet || op == http.MethodDelete || op == http.MethodPost || op == http.MethodPut {
				logger.LogWhenDebug(fmt.Sprintf("  operation: %s", op))
				id := opItem.OperationID
				if id == "" {
					id = fmt.Sprintf("%d-%s-%s", counter, strings.ToLower(op), path[1:])
				}
				var ep *mockgo.Endpoint
				if len(opItem.Responses) == 0 {
					logger.LogWhenDebug("    no responses defined, using default 200")
					ep = createEndpoint(id, op, path, "200")
				} else {
					status, successResponse := findSuccessResponse(opItem.Responses)
					logger.LogWhenDebug(fmt.Sprintf("    response: %s:", status))
					ep = createEndpoint(id, op, path, status)
					if successResponse.Value != nil {
						if len(successResponse.Value.Headers) > 0 {
							for header, _ := range successResponse.Value.Headers {
								ep.Response.Headers += fmt.Sprintf("%s: !!Value %s !!\n", header, header)
							}
						}
						addBody(ep, *impl, opItem, successResponse, logger)
					}
				}
				mock.Endpoints = append(mock.Endpoints, ep)
			} else {
				logger.LogError(fmt.Sprintf("Unsupported method: %s", op), fmt.Errorf("skipping operation"))
				continue
			}
		}
	}

	// marshal mock to yaml
	data, err := yaml.Marshal(mock)
	if err != nil {
		logger.LogError("Error marshaling mock", err)
		os.Exit(1)
	}
	if *outputFile != "" {
		err = os.WriteFile(*outputFile, data, 0644)
		if err != nil {
			logger.LogError("Error writing to file", err)
			os.Exit(1)
		}
		logger.LogWhenVerbose(fmt.Sprintf("Wrote mock to file: %s", *outputFile))
	} else {
		fmt.Println(string(data))
	}
}

func findSuccessResponse(responses openapi3.Responses) (string, *openapi3.ResponseRef) {
	for status, resp := range responses {
		if strings.HasPrefix(status, "2") {
			return status, resp
		}
	}
	for status, resp := range responses {
		if strings.HasPrefix(status, "3") {
			return status, resp
		}
	}
	for status, resp := range responses {
		if strings.HasPrefix(status, "4") {
			return status, resp
		}
	}
	for status, resp := range responses {
		if strings.HasPrefix(status, "5") {
			return status, resp
		}
	}
	return "200", &openapi3.ResponseRef{Value: &openapi3.Response{}}
}

func createEndpoint(id, method, path, responseStatus string) *mockgo.Endpoint {
	responseStatus = strings.TrimSpace(responseStatus)
	if _, err := strconv.Atoi(responseStatus); err != nil {
		responseStatus = "200"
	}
	return &mockgo.Endpoint{ID: id,
		Request:  &mockgo.MatchRequest{Method: method, Path: path},
		Response: &mockgo.Response{StatusCode: responseStatus}}
}

func addBody(ep *mockgo.Endpoint, impl bool, opItem *openapi3.Operation, resp *openapi3.ResponseRef, logger *logging.LoggerUtil) {
	if !impl {
		if resp.Value != nil {
			if resp.Value.Content != nil {
				for contentType, content := range resp.Value.Content {
					logger.LogWhenDebug(fmt.Sprintf("      content-type: %s", contentType))
					ep.Response.Headers += fmt.Sprintf("content-type: %s\n", contentType)
					if content != nil {
						if content.Example != nil {
							addBodyExample(ep, content.Example, logger)
						}
						if content.Examples != nil {
							for _, example := range content.Examples {
								addBodyExample(ep, example.Value.Value, logger)
								break // only use first example body
							}
						}
					}
				}
			}
		}
	} else {
		addBodyTemplate(ep, opItem, logger)
	}
}

func addBodyTemplate(ep *mockgo.Endpoint, item *openapi3.Operation, logger *logging.LoggerUtil) {
	template := ""
	var keyParam *openapi3.Parameter
	hasRequestBody := false
	if item.Parameters != nil {
		for _, param := range item.Parameters {
			switch param.Value.In {
			case openapi3.ParameterInPath:
				if keyParam == nil {
					keyParam = param.Value
					logger.LogWhenDebug(fmt.Sprintf("      identified key param: '%s'", keyParam.Name))
				}
				logger.LogWhenDebug(fmt.Sprintf("      found path param: '%s'", param.Value.Name))
				template += fmt.Sprintf("{{ $%s := .RequestPathParams.%s }}\n", param.Value.Name, param.Value.Name)
			case openapi3.ParameterInQuery:
				if keyParam == nil {
					keyParam = param.Value
					logger.LogWhenDebug(fmt.Sprintf("      identified key param: '%s'", keyParam.Name))
				}
				logger.LogWhenDebug(fmt.Sprintf("      found query param: '%s'", param.Value.Name))
				template += fmt.Sprintf("{{ $%s := .RequestQueryParams.%s }}\n", param.Value.Name, param.Value.Name)
			case openapi3.ParameterInHeader:
				logger.LogWhenDebug(fmt.Sprintf("      found header param: '%s'", param.Value.Name))
				template += fmt.Sprintf("{{ $%s := .RequestHeader.%s }}\n", param.Value.Name, param.Value.Name)
			case "body":
				logger.LogWhenDebug(fmt.Sprintf("      found body param"))
				hasRequestBody = true
				template += "{{ $reqBody := .RequestBody }}\n"
			}

		}
	}
	if keyParam != nil {
		keyTemplate := fmt.Sprintf("( printf \"%s-%%s\" $%s )", keyParam.Name, keyParam.Name)
		switch ep.Request.Method {
		case http.MethodGet:
			template += fmt.Sprintf("{{ kvStoreGet %s -}}\n", keyTemplate)
		case http.MethodPost, http.MethodPut:
			if hasRequestBody {
				template += fmt.Sprintf("{{ kvStorePut %s $reqBody -}}\n", keyTemplate)
			} else {
				template += fmt.Sprintf("{{ kvStorePut %s $%s -}}\n", keyTemplate, keyParam.Name)
			}
		case http.MethodDelete:
			template += fmt.Sprintf("{{ kvStorePut %s \"\" -}}\n", keyTemplate)
		}
	}
	ep.Response.Body = template
}

func hasJsonRequestBody(item *openapi3.Operation) bool {
	if item.RequestBody != nil {
		if item.RequestBody.Value != nil {
			if item.RequestBody.Value.Content != nil {
				for contentType := range item.RequestBody.Value.Content {
					if strings.Contains(contentType, "json") {
						return true
					}
				}
			}
		}
	}
	return false
}

func addBodyExample(ep *mockgo.Endpoint, example interface{}, logger *logging.LoggerUtil) {
	data, err := json.MarshalIndent(example, "", "  ")
	if err != nil {
		logger.LogError("Error marshaling example body", err)
		return
	}
	ep.Response.Body = string(data)
	logger.LogWhenDebug(fmt.Sprintf("       example body: %s", ep.Response.Body))
}
