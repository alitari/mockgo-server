package mock

import (
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/alitari/mockgo-server/mockgo/logging"
	"github.com/alitari/mockgo-server/mockgo/matches"
	"github.com/alitari/mockgo-server/mockgo/testutil"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

type mockTestCase struct {
	name                   string
	method                 string
	path                   string
	header                 testutil.Header
	body                   string
	expectedStatusCode     int
	expectedResponseBody   string
	expectedResponseHeader map[string]string
}

var router = mux.NewRouter()

func TestMain(m *testing.M) {
	mockRequestHandler := NewMockRequestHandler("../../test/mocks", "*-mock.yaml", matches.NewInMemoryMatchstore(uint16(100)), logging.NewLoggerUtil(logging.Debug))
	if err := mockRequestHandler.LoadFiles(nil); err != nil {
		log.Fatal(err)
	}
	mockRequestHandler.AddRoutes(router)

	testutil.StartServing(router)
	code := testutil.RunAndCheckCoverage("requestHandlerTest", m, 0.49)
	testutil.StopServing()
	os.Exit(code)
}

func TestMockRequestHandler_serving_matches(t *testing.T) {
	testCases := []*mockTestCase{
		{name: "match first", method: http.MethodGet, path: "/first",
			expectedStatusCode:     http.StatusNoContent,
			expectedResponseHeader: map[string]string{"Endpoint-Id": "1"},
		},
		{name: "no match wrong path", method: http.MethodGet, path: "/minimalwrong",
			expectedStatusCode:     http.StatusNotFound,
			expectedResponseHeader: map[string]string{"Content-Type": "text/plain; charset=utf-8"},
			expectedResponseBody:   "404 page not found\n"},
		{name: "no match path length too long", method: http.MethodGet, path: "/minimal/foo",
			expectedStatusCode:     http.StatusNotFound,
			expectedResponseHeader: map[string]string{"Content-Type": "text/plain; charset=utf-8"},
			expectedResponseBody:   "404 page not found\n"},
		{name: "no match path length too short", method: http.MethodGet, path: "/",
			expectedStatusCode:     http.StatusNotFound,
			expectedResponseHeader: map[string]string{"Content-Type": "text/plain; charset=utf-8"},
			expectedResponseBody:   "404 page not found\n"},
		{name: "no match wrong method", method: http.MethodPost, path: "/minimal",
			expectedStatusCode:     http.StatusNotFound,
			expectedResponseHeader: map[string]string{"Content-Type": "text/plain; charset=utf-8"},
			expectedResponseBody:   "404 page not found\n"},
		{name: "minimal match ", method: http.MethodGet, path: "/minimal",
			expectedStatusCode:     http.StatusNoContent,
			expectedResponseHeader: map[string]string{"Endpoint-Id": "minimal"},
			expectedResponseBody:   ""},
		{name: "minimal match with header", method: http.MethodGet, path: "/minimal", header: testutil.CreateHeader().WithJsonAccept(),
			expectedStatusCode:     http.StatusNoContent,
			expectedResponseHeader: map[string]string{"Endpoint-Id": "minimal"},
			expectedResponseBody:   ""},
		{name: "minimal match with query", method: http.MethodGet, path: "/minimal?firstQueryParam=value1&secondQueryParam=value2",
			expectedStatusCode:     http.StatusNoContent,
			expectedResponseHeader: map[string]string{"Endpoint-Id": "minimal"},
			expectedResponseBody:   ""},
		{name: "minimal match with body", method: http.MethodGet, path: "/minimal", body: "my body",
			expectedStatusCode:     http.StatusNoContent,
			expectedResponseHeader: map[string]string{"Endpoint-Id": "minimal"},
			expectedResponseBody:   ""},
		{name: "maximal no match wrong query params", method: http.MethodPost,
			path:                   "/maximal?firstQueryParam=value1&wrongQueryParam=value",
			header:                 testutil.CreateHeader().WithJsonContentType().WithKeyValue("Myheader", "myheaderValue"),
			body:                   "{\n  \"mybody\": \"is max\"\n}\n",
			expectedStatusCode:     http.StatusNotFound,
			expectedResponseHeader: map[string]string{"Content-Type": "text/plain; charset=utf-8"},
			expectedResponseBody:   "404 page not found\n"},
		{name: "maximal no match wrong header", method: http.MethodPost,
			path:                   "/maximal?firstQueryParam=value1&secondQueryParam=value2",
			header:                 testutil.CreateHeader().WithJsonContentType().WithKeyValue("Myheader", "wrong"),
			body:                   "{\n  \"mybody\": \"is max\"\n}\n",
			expectedStatusCode:     http.StatusNotFound,
			expectedResponseHeader: map[string]string{"Content-Type": "text/plain; charset=utf-8"},
			expectedResponseBody:   "404 page not found\n"},
		{name: "maximal no match wrong body", method: http.MethodPost,
			path:                   "/maximal?firstQueryParam=value1&secondQueryParam=value2",
			header:                 testutil.CreateHeader().WithJsonContentType().WithKeyValue("Myheader", "myheaderValue"),
			body:                   "{\n  \"mybody\": \"is wrong\"\n}\n",
			expectedStatusCode:     http.StatusNotFound,
			expectedResponseHeader: map[string]string{"Content-Type": "text/plain; charset=utf-8"},
			expectedResponseBody:   "404 page not found\n"},
		{name: "maximal match", method: http.MethodPost,
			path:                   "/maximal?firstQueryParam=value1&secondQueryParam=value2",
			header:                 testutil.CreateHeader().WithJsonContentType().WithKeyValue("Myheader", "myheaderValue"),
			body:                   "{\n  \"mybody\": \"is max\"\n}\n",
			expectedStatusCode:     http.StatusNoContent,
			expectedResponseHeader: map[string]string{"Endpoint-Id": "maximal"},
			expectedResponseBody:   ""},
		{name: "maximal match header and query superset", method: http.MethodPost,
			path:                   "/maximal?firstQueryParam=value1&secondQueryParam=value2&thirdQueryParam=value3",
			header:                 testutil.CreateHeader().WithJsonContentType().WithKeyValue("Myheader", "myheaderValue").WithKeyValue("AnotherHeader", "anotherheaderValue"),
			body:                   "{\n  \"mybody\": \"is max\"\n}\n",
			expectedStatusCode:     http.StatusNoContent,
			expectedResponseHeader: map[string]string{"Endpoint-Id": "maximal"},
			expectedResponseBody:   ""},
		{name: "no match single wildcard first path segment", method: http.MethodGet, path: "/wildcardwrong/bar/foo",
			expectedStatusCode:     http.StatusNotFound,
			expectedResponseHeader: map[string]string{"Content-Type": "text/plain; charset=utf-8"},
			expectedResponseBody:   "404 page not found\n"},
		{name: "no match single wildcard path too long", method: http.MethodGet, path: "/wildcard/bar/foo/toolong",
			expectedStatusCode:     http.StatusNotFound,
			expectedResponseHeader: map[string]string{"Content-Type": "text/plain; charset=utf-8"},
			expectedResponseBody:   "404 page not found\n"},
		{name: "no match single wildcard path too short", method: http.MethodGet, path: "/wildcard/bar",
			expectedStatusCode:     http.StatusNotFound,
			expectedResponseHeader: map[string]string{"Content-Type": "text/plain; charset=utf-8"},
			expectedResponseBody:   "404 page not found\n"},
		{name: "match single wildcard 1", method: http.MethodGet, path: "/wildcard/bar/foo",
			expectedStatusCode:     http.StatusNoContent,
			expectedResponseHeader: map[string]string{"Endpoint-Id": "singlewildcard"},
			expectedResponseBody:   ""},
		{name: "match single wildcard 2", method: http.MethodGet, path: "/wildcard/foo/foo",
			expectedStatusCode:     http.StatusNoContent,
			expectedResponseHeader: map[string]string{"Endpoint-Id": "singlewildcard"},
			expectedResponseBody:   ""},
		{name: "match multi wildcard", method: http.MethodGet, path: "/multiwildcard/bar/foo/bar",
			expectedStatusCode:     http.StatusNoContent,
			expectedResponseHeader: map[string]string{"Endpoint-Id": "multiwildcard"},
			expectedResponseBody:   ""},
		{name: "no match allmatchwildcardAtTheEnds first path segment", method: http.MethodGet, path: "/allmatchwildcardAtTheEnds/foo",
			expectedStatusCode:     http.StatusNotFound,
			expectedResponseHeader: map[string]string{"Content-Type": "text/plain; charset=utf-8"},
			expectedResponseBody:   "404 page not found\n"},

		{name: "match allmatchwildcardAtTheEnd 1", method: http.MethodGet, path: "/allmatchwildcardAtTheEnd/bar",
			expectedStatusCode:     http.StatusNoContent,
			expectedResponseHeader: map[string]string{"Endpoint-Id": "allmatchwildcardAtTheEnd"},
			expectedResponseBody:   ""},
		{name: "match allmatchwildcardAtTheEnd 2", method: http.MethodGet, path: "/allmatchwildcardAtTheEnd/foo/bar",
			expectedStatusCode:     http.StatusNoContent,
			expectedResponseHeader: map[string]string{"Endpoint-Id": "allmatchwildcardAtTheEnd"},
			expectedResponseBody:   ""},
		{name: "no match allmatchwildcardInTheMiddle endsegments wrong", method: http.MethodGet, path: "/allmatchwildcardInTheMiddle/bar/foo/foo",
			expectedStatusCode:     http.StatusNotFound,
			expectedResponseHeader: map[string]string{"Content-Type": "text/plain; charset=utf-8"},
			expectedResponseBody:   "404 page not found\n"},
		{name: "match allmatchwildcardInTheMiddle 1", method: http.MethodGet, path: "/allmatchwildcardInTheMiddle/bar/foo",
			expectedStatusCode:     http.StatusNoContent,
			expectedResponseHeader: map[string]string{"Endpoint-Id": "allmatchwildcardInTheMiddle"},
			expectedResponseBody:   ""},
		{name: "match allmatchwildcardInTheMiddle 2", method: http.MethodGet, path: "/allmatchwildcardInTheMiddle/bar/ext/foo",
			expectedStatusCode:     http.StatusNoContent,
			expectedResponseHeader: map[string]string{"Endpoint-Id": "allmatchwildcardInTheMiddle"},
			expectedResponseBody:   ""},
		{name: "match allmatchwildcardInTheMiddle 3", method: http.MethodGet, path: "/allmatchwildcardInTheMiddle/bar/ext/rem/foo",
			expectedStatusCode:     http.StatusNoContent,
			expectedResponseHeader: map[string]string{"Endpoint-Id": "allmatchwildcardInTheMiddle"},
			expectedResponseBody:   ""},
		{name: "no match combined wildcards last segment missing", method: http.MethodGet, path: "/combinedwildcards1/bar/a/b/c/foo",
			expectedStatusCode:     http.StatusNotFound,
			expectedResponseHeader: map[string]string{"Content-Type": "text/plain; charset=utf-8"},
			expectedResponseBody:   "404 page not found\n"},
		{name: "match combined wildcards single segment", method: http.MethodGet, path: "/combinedwildcards1/bar/foo/ext",
			expectedStatusCode:     http.StatusNoContent,
			expectedResponseHeader: map[string]string{"Endpoint-Id": "combinedwildcards1"},
			expectedResponseBody:   ""},
		{name: "match combined wildcards multiple segment", method: http.MethodGet, path: "/combinedwildcards1/bar/a/b/c/foo/d",
			expectedStatusCode:     http.StatusNoContent,
			expectedResponseHeader: map[string]string{"Endpoint-Id": "combinedwildcards1"},
			expectedResponseBody:   ""},
		{name: "no match single path param", method: http.MethodGet, path: "/pathParams/bar/foos",
			expectedStatusCode:     http.StatusNotFound,
			expectedResponseHeader: map[string]string{"Content-Type": "text/plain; charset=utf-8"},
			expectedResponseBody:   "404 page not found\n"},
		{name: "match single path param", method: http.MethodGet, path: "/pathParams/bar/foo",
			expectedStatusCode:     http.StatusOK,
			expectedResponseHeader: map[string]string{"Endpoint-Id": "singlepathparam"},
			expectedResponseBody:   "pathParam=bar"},
		{name: "match multi path param", method: http.MethodGet, path: "/multipathParams/val1/foo/val2",
			expectedStatusCode:     http.StatusOK,
			expectedResponseHeader: map[string]string{"Endpoint-Id": "multipathparam"},
			expectedResponseBody:   "pathParam1=val1\npathParam2=val2"},
		{name: "match simple prio", method: http.MethodGet, path: "/prio",
			expectedStatusCode:     http.StatusNoContent,
			expectedResponseHeader: map[string]string{"Endpoint-Id": "mustwin"},
			expectedResponseBody:   ""},
		{name: "no match regexp 1 wrong body", method: http.MethodPost, path: "/regexp1", body: "{ alex ",
			expectedStatusCode:     http.StatusNotFound,
			expectedResponseHeader: map[string]string{"Content-Type": "text/plain; charset=utf-8"},
			expectedResponseBody:   "404 page not found\n"},
		{name: "match regexp 1", method: http.MethodPost, path: "/regexp1", body: "{ alex }",
			expectedStatusCode:     http.StatusNoContent,
			expectedResponseHeader: map[string]string{"Endpoint-Id": "regexpmock1"},
			expectedResponseBody:   ""},
		{name: "match regexp 2", method: http.MethodPost, path: "/regexp2", body: "{\n alex \n}",
			expectedStatusCode:     http.StatusNoContent,
			expectedResponseHeader: map[string]string{"Endpoint-Id": "regexpmock2"},
			expectedResponseBody:   ""},
		{name: "match regexp 3", method: http.MethodPost, path: "/regexp3", body: `{ "email": "foo@bar.com" }`,
			expectedStatusCode:     http.StatusNoContent,
			expectedResponseHeader: map[string]string{"Endpoint-Id": "regexpmock3"},
			expectedResponseBody:   ""},
		{name: "match responsetemplates", method: http.MethodGet, path: "/responsetemplates/foo?query1=queryvalue1&query2=queryvalue2",
			body:                   `{ "mybody": "is cool!" }`,
			header:                 testutil.CreateHeader().WithKeyValue("headerKey", "headerValue"),
			expectedStatusCode:     http.StatusOK,
			expectedResponseHeader: map[string]string{"Endpoint-Id": "response-templates"},
			expectedResponseBody: `RequestPathParams=map[pathparam1:foo]
RequestQueryParams=map[query1:queryvalue1 query2:queryvalue2]
RequestUrl=/responsetemplates/foo?query1=queryvalue1&query2=queryvalue2
RequestUser=
RequestPath=/responsetemplates/foo
RequestHost=
RequestBody={ "mybody": "is cool!" }
RequestBodyJsonData=map[mybody:is cool!]`},
	}
	assertTestcases(t, testCases)
}

func assertTestcases(t *testing.T, mockTestCases []*mockTestCase) {
	for _, testCase := range mockTestCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.NoError(t, testutil.AssertResponseOfRequestCall(t,
				testutil.CreateOutgoingRequest(t, testCase.method, testCase.path, testCase.header, testCase.body),
				func(response *http.Response, responseBody string) {
					assert.Equal(t, testCase.expectedStatusCode, response.StatusCode, "unexpected response status code")
					for k, v := range testCase.expectedResponseHeader {
						hk := response.Header[k]
						if len(hk) == 0 {
							assert.Failf(t, "header key not found", "expected header key '%s' does not exist, actual header=%v", k, response.Header)
						} else {
							assert.Equalf(t, v, response.Header[k][0], "unexpected response header value of key '%s': expected '%s' , but is '%s'", k, v, response.Header[k][0])
						}
					}
					assert.Equal(t, testCase.expectedResponseBody, responseBody, "unexpected response body")
				}))
		})
	}
}
