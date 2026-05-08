package jsonrpc

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	stdjson "github.com/goccy/go-json"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/require"
)

// needed to retrieve requests that arrived at httpServer for further investigation
var requestChan = make(chan *RequestData, 1)

// the request datastructure that can be retrieved for test assertions
type RequestData struct {
	request *http.Request
	body    string
}

// set the response body the httpServer should return for the next request
var responseBody = ""

var httpServer *httptest.Server

// start the testhttp server and stop it when tests are finished
func TestMain(m *testing.M) {
	httpServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		defer r.Body.Close()
		// put request and body to channel for the client to investigate them
		requestChan <- &RequestData{r, string(data)}

		fmt.Fprint(w, responseBody)
	}))
	defer httpServer.Close()

	os.Exit(m.Run())
}

func TestSimpleRpcCallHeaderCorrect(t *testing.T) {
	RegisterTestingT(t)

	rpcClient := NewClient(httpServer.URL)
	rpcClient.Call(context.Background(), "add", 1, 2)

	req := (<-requestChan).request

	Expect(req.Method).To(Equal("POST"))
	Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))
	Expect(req.Header.Get("Accept")).To(Equal("application/json"))
}

func TestCustomHeader_PreservesMultiValue(t *testing.T) {
	RegisterTestingT(t)

	hdr := http.Header{}
	hdr.Add("X-Forwarded-For", "1.1.1.1")
	hdr.Add("X-Forwarded-For", "2.2.2.2")
	hdr.Add("Cookie", "a=1")
	hdr.Add("Cookie", "b=2")
	hdr.Set("Authorization", "Bearer t")

	rpcClient := NewClientWithOpts(httpServer.URL, &RPCClientOpts{
		CustomHeader: hdr,
	})
	rpcClient.Call(context.Background(), "noop")

	req := (<-requestChan).request

	Expect(req.Header.Values("X-Forwarded-For")).To(Equal([]string{"1.1.1.1", "2.2.2.2"}))
	Expect(req.Header.Values("Cookie")).To(Equal([]string{"a=1", "b=2"}))
	Expect(req.Header.Get("Authorization")).To(Equal("Bearer t"))
}

func TestCustomHeader_TakesPrecedenceOverCustomHeaders(t *testing.T) {
	RegisterTestingT(t)

	rpcClient := NewClientWithOpts(httpServer.URL, &RPCClientOpts{
		CustomHeaders: map[string]string{
			"X-Test": "from-map",
		},
		CustomHeader: http.Header{
			"X-Test": []string{"from-header-a", "from-header-b"},
		},
	})
	rpcClient.Call(context.Background(), "noop")

	req := (<-requestChan).request
	Expect(req.Header.Values("X-Test")).To(Equal([]string{"from-header-a", "from-header-b"}))
}

func TestCustomHeaders_StillWorks(t *testing.T) {
	RegisterTestingT(t)

	rpcClient := NewClientWithOpts(httpServer.URL, &RPCClientOpts{
		CustomHeaders: map[string]string{
			"X-Backcompat": "yes",
		},
	})
	rpcClient.Call(context.Background(), "noop")

	req := (<-requestChan).request
	Expect(req.Header.Get("X-Backcompat")).To(Equal("yes"))
}

// test if the structure of an rpc request is built correctly by validating the data that arrived on the test server
func TestRpcClient_Call(t *testing.T) {
	RegisterTestingT(t)
	rpcClient := NewClient(httpServer.URL)

	person := Person{
		Name:    "Alex",
		Age:     35,
		Country: "Germany",
	}

	drink := Drink{
		Name:        "Cuba Libre",
		Ingredients: []string{"rum", "cola"},
	}

	useFixedID = true

	rpcClient.Call(context.Background(), "missingParam")
	Expect((<-requestChan).body).To(Equal(`{"method":"missingParam","id":1,"jsonrpc":"2.0"}`))

	rpcClient.Call(context.Background(), "nullParam", nil)
	Expect((<-requestChan).body).To(Equal(`{"method":"nullParam","params":[null],"id":1,"jsonrpc":"2.0"}`))

	rpcClient.Call(context.Background(), "nullParams", nil, nil)
	Expect((<-requestChan).body).To(Equal(`{"method":"nullParams","params":[null,null],"id":1,"jsonrpc":"2.0"}`))

	rpcClient.Call(context.Background(), "emptyParams", []any{})
	Expect((<-requestChan).body).To(Equal(`{"method":"emptyParams","params":[],"id":1,"jsonrpc":"2.0"}`))

	rpcClient.Call(context.Background(), "emptyAnyParams", []string{})
	Expect((<-requestChan).body).To(Equal(`{"method":"emptyAnyParams","params":[],"id":1,"jsonrpc":"2.0"}`))

	rpcClient.Call(context.Background(), "emptyObject", struct{}{})
	Expect((<-requestChan).body).To(Equal(`{"method":"emptyObject","params":{},"id":1,"jsonrpc":"2.0"}`))

	rpcClient.Call(context.Background(), "emptyObjectList", []struct{}{{}, {}})
	Expect((<-requestChan).body).To(Equal(`{"method":"emptyObjectList","params":[{},{}],"id":1,"jsonrpc":"2.0"}`))

	rpcClient.Call(context.Background(), "boolParam", true)
	Expect((<-requestChan).body).To(Equal(`{"method":"boolParam","params":[true],"id":1,"jsonrpc":"2.0"}`))

	rpcClient.Call(context.Background(), "boolParams", true, false, true)
	Expect((<-requestChan).body).To(Equal(`{"method":"boolParams","params":[true,false,true],"id":1,"jsonrpc":"2.0"}`))

	rpcClient.Call(context.Background(), "stringParam", "Alex")
	Expect((<-requestChan).body).To(Equal(`{"method":"stringParam","params":["Alex"],"id":1,"jsonrpc":"2.0"}`))

	rpcClient.Call(context.Background(), "stringParams", "JSON", "RPC")
	Expect((<-requestChan).body).To(Equal(`{"method":"stringParams","params":["JSON","RPC"],"id":1,"jsonrpc":"2.0"}`))

	rpcClient.Call(context.Background(), "numberParam", 123)
	Expect((<-requestChan).body).To(Equal(`{"method":"numberParam","params":[123],"id":1,"jsonrpc":"2.0"}`))

	rpcClient.Call(context.Background(), "numberParams", 123, 321)
	Expect((<-requestChan).body).To(Equal(`{"method":"numberParams","params":[123,321],"id":1,"jsonrpc":"2.0"}`))

	rpcClient.Call(context.Background(), "floatParam", 1.23)
	Expect((<-requestChan).body).To(Equal(`{"method":"floatParam","params":[1.23],"id":1,"jsonrpc":"2.0"}`))

	rpcClient.Call(context.Background(), "floatParams", 1.23, 3.21)
	Expect((<-requestChan).body).To(Equal(`{"method":"floatParams","params":[1.23,3.21],"id":1,"jsonrpc":"2.0"}`))

	rpcClient.Call(context.Background(), "manyParams", "Alex", 35, true, nil, 2.34)
	Expect((<-requestChan).body).To(Equal(`{"method":"manyParams","params":["Alex",35,true,null,2.34],"id":1,"jsonrpc":"2.0"}`))

	rpcClient.Call(context.Background(), "emptyMissingPublicFieldObject", struct{ name string }{name: "Alex"})
	Expect((<-requestChan).body).To(Equal(`{"method":"emptyMissingPublicFieldObject","params":{},"id":1,"jsonrpc":"2.0"}`))

	rpcClient.Call(context.Background(), "singleStruct", person)
	Expect((<-requestChan).body).To(Equal(`{"method":"singleStruct","params":{"name":"Alex","age":35,"country":"Germany"},"id":1,"jsonrpc":"2.0"}`))

	rpcClient.Call(context.Background(), "singlePointerToStruct", &person)
	Expect((<-requestChan).body).To(Equal(`{"method":"singlePointerToStruct","params":{"name":"Alex","age":35,"country":"Germany"},"id":1,"jsonrpc":"2.0"}`))

	pp := &person
	rpcClient.Call(context.Background(), "doublePointerStruct", &pp)
	Expect((<-requestChan).body).To(Equal(`{"method":"doublePointerStruct","params":{"name":"Alex","age":35,"country":"Germany"},"id":1,"jsonrpc":"2.0"}`))

	rpcClient.Call(context.Background(), "multipleStructs", person, &drink)
	Expect((<-requestChan).body).To(Equal(`{"method":"multipleStructs","params":[{"name":"Alex","age":35,"country":"Germany"},{"name":"Cuba Libre","ingredients":["rum","cola"]}],"id":1,"jsonrpc":"2.0"}`))

	rpcClient.Call(context.Background(), "singleStructInArray", []any{person})
	Expect((<-requestChan).body).To(Equal(`{"method":"singleStructInArray","params":[{"name":"Alex","age":35,"country":"Germany"}],"id":1,"jsonrpc":"2.0"}`))

	rpcClient.Call(context.Background(), "namedParameters", map[string]any{
		"name": "Alex",
		"age":  35,
	})
	Expect((<-requestChan).body).To(Equal(`{"method":"namedParameters","params":{"age":35,"name":"Alex"},"id":1,"jsonrpc":"2.0"}`))

	rpcClient.Call(context.Background(), "anonymousStructNoTags", struct {
		Name string
		Age  int
	}{"Alex", 33})
	Expect((<-requestChan).body).To(Equal(`{"method":"anonymousStructNoTags","params":{"Name":"Alex","Age":33},"id":1,"jsonrpc":"2.0"}`))

	rpcClient.Call(context.Background(), "anonymousStructWithTags", struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{"Alex", 33})
	Expect((<-requestChan).body).To(Equal(`{"method":"anonymousStructWithTags","params":{"name":"Alex","age":33},"id":1,"jsonrpc":"2.0"}`))

	rpcClient.Call(context.Background(), "structWithNullField", struct {
		Name    string  `json:"name"`
		Address *string `json:"address"`
	}{"Alex", nil})
	Expect((<-requestChan).body).To(Equal(`{"method":"structWithNullField","params":{"name":"Alex","address":null},"id":1,"jsonrpc":"2.0"}`))
}

func TestRpcClient_CallBatch(t *testing.T) {
	RegisterTestingT(t)
	rpcClient := NewClient(httpServer.URL)

	person := Person{
		Name:    "Alex",
		Age:     35,
		Country: "Germany",
	}

	drink := Drink{
		Name:        "Cuba Libre",
		Ingredients: []string{"rum", "cola"},
	}

	// invalid parameters are possible by manually defining *RPCRequest
	rpcClient.CallBatch(context.Background(), RPCRequests{
		{
			Method: "singleRequest",
			Params: 3, // invalid, should be []int{3}
		},
	})
	Expect((<-requestChan).body).To(Equal(`[{"method":"singleRequest","params":3,"id":0,"jsonrpc":"2.0"}]`))

	// better use Params() unless you know what you are doing
	rpcClient.CallBatch(context.Background(), RPCRequests{
		{
			Method: "singleRequest",
			Params: Params(3), // always valid json rpc
		},
	})
	Expect((<-requestChan).body).To(Equal(`[{"method":"singleRequest","params":[3],"id":0,"jsonrpc":"2.0"}]`))

	// even better, use NewRequest()
	rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("multipleRequests1", 1),
		NewRequest("multipleRequests2", 2),
		NewRequest("multipleRequests3", 3),
	})
	Expect((<-requestChan).body).To(Equal(`[{"method":"multipleRequests1","params":[1],"id":0,"jsonrpc":"2.0"},{"method":"multipleRequests2","params":[2],"id":1,"jsonrpc":"2.0"},{"method":"multipleRequests3","params":[3],"id":2,"jsonrpc":"2.0"}]`))

	// test a huge batch request
	requests := RPCRequests{
		NewRequest("nullParam", nil),
		NewRequest("nullParams", nil, nil),
		NewRequest("emptyParams", []any{}),
		NewRequest("emptyAnyParams", []string{}),
		NewRequest("emptyObject", struct{}{}),
		NewRequest("emptyObjectList", []struct{}{{}, {}}),
		NewRequest("boolParam", true),
		NewRequest("boolParams", true, false, true),
		NewRequest("stringParam", "Alex"),
		NewRequest("stringParams", "JSON", "RPC"),
		NewRequest("numberParam", 123),
		NewRequest("numberParams", 123, 321),
		NewRequest("floatParam", 1.23),
		NewRequest("floatParams", 1.23, 3.21),
		NewRequest("manyParams", "Alex", 35, true, nil, 2.34),
		NewRequest("emptyMissingPublicFieldObject", struct{ name string }{name: "Alex"}),
		NewRequest("singleStruct", person),
		NewRequest("singlePointerToStruct", &person),
		NewRequest("multipleStructs", person, &drink),
		NewRequest("singleStructInArray", []any{person}),
		NewRequest("namedParameters", map[string]any{
			"name": "Alex",
			"age":  35,
		}),
		NewRequest("anonymousStructNoTags", struct {
			Name string
			Age  int
		}{"Alex", 33}),
		NewRequest("anonymousStructWithTags", struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}{"Alex", 33}),
		NewRequest("structWithNullField", struct {
			Name    string  `json:"name"`
			Address *string `json:"address"`
		}{"Alex", nil}),
	}
	rpcClient.CallBatch(context.Background(), requests)

	Expect((<-requestChan).body).To(Equal(`[{"method":"nullParam","params":[null],"id":0,"jsonrpc":"2.0"},` +
		`{"method":"nullParams","params":[null,null],"id":1,"jsonrpc":"2.0"},` +
		`{"method":"emptyParams","params":[],"id":2,"jsonrpc":"2.0"},` +
		`{"method":"emptyAnyParams","params":[],"id":3,"jsonrpc":"2.0"},` +
		`{"method":"emptyObject","params":{},"id":4,"jsonrpc":"2.0"},` +
		`{"method":"emptyObjectList","params":[{},{}],"id":5,"jsonrpc":"2.0"},` +
		`{"method":"boolParam","params":[true],"id":6,"jsonrpc":"2.0"},` +
		`{"method":"boolParams","params":[true,false,true],"id":7,"jsonrpc":"2.0"},` +
		`{"method":"stringParam","params":["Alex"],"id":8,"jsonrpc":"2.0"},` +
		`{"method":"stringParams","params":["JSON","RPC"],"id":9,"jsonrpc":"2.0"},` +
		`{"method":"numberParam","params":[123],"id":10,"jsonrpc":"2.0"},` +
		`{"method":"numberParams","params":[123,321],"id":11,"jsonrpc":"2.0"},` +
		`{"method":"floatParam","params":[1.23],"id":12,"jsonrpc":"2.0"},` +
		`{"method":"floatParams","params":[1.23,3.21],"id":13,"jsonrpc":"2.0"},` +
		`{"method":"manyParams","params":["Alex",35,true,null,2.34],"id":14,"jsonrpc":"2.0"},` +
		`{"method":"emptyMissingPublicFieldObject","params":{},"id":15,"jsonrpc":"2.0"},` +
		`{"method":"singleStruct","params":{"name":"Alex","age":35,"country":"Germany"},"id":16,"jsonrpc":"2.0"},` +
		`{"method":"singlePointerToStruct","params":{"name":"Alex","age":35,"country":"Germany"},"id":17,"jsonrpc":"2.0"},` +
		`{"method":"multipleStructs","params":[{"name":"Alex","age":35,"country":"Germany"},{"name":"Cuba Libre","ingredients":["rum","cola"]}],"id":18,"jsonrpc":"2.0"},` +
		`{"method":"singleStructInArray","params":[{"name":"Alex","age":35,"country":"Germany"}],"id":19,"jsonrpc":"2.0"},` +
		`{"method":"namedParameters","params":{"age":35,"name":"Alex"},"id":20,"jsonrpc":"2.0"},` +
		`{"method":"anonymousStructNoTags","params":{"Name":"Alex","Age":33},"id":21,"jsonrpc":"2.0"},` +
		`{"method":"anonymousStructWithTags","params":{"name":"Alex","age":33},"id":22,"jsonrpc":"2.0"},` +
		`{"method":"structWithNullField","params":{"name":"Alex","address":null},"id":23,"jsonrpc":"2.0"}]`))

	// create batch manually
	requests = []*RPCRequest{
		{
			Method:  "myMethod1",
			Params:  []int{1},
			ID:      123,   // will be forced to requests[i].ID == i unless you use CallBatchRaw
			JSONRPC: "7.0", // will be forced to "2.0"  unless you use CallBatchRaw
		},
		{
			Method:  "myMethod2",
			Params:  &person,
			ID:      321,     // will be forced to requests[i].ID == i unless you use CallBatchRaw
			JSONRPC: "wrong", // will be forced to "2.0" unless you use CallBatchRaw
		},
	}
	rpcClient.CallBatch(context.Background(), requests)

	Expect((<-requestChan).body).To(Equal(`[{"method":"myMethod1","params":[1],"id":0,"jsonrpc":"2.0"},` +
		`{"method":"myMethod2","params":{"name":"Alex","age":35,"country":"Germany"},"id":1,"jsonrpc":"2.0"}]`))

	// use raw batch
	requests = []*RPCRequest{
		{
			Method:  "myMethod1",
			Params:  []int{1},
			ID:      123,
			JSONRPC: "7.0",
		},
		{
			Method:  "myMethod2",
			Params:  &person,
			ID:      321,
			JSONRPC: "wrong",
		},
	}
	rpcClient.CallBatchRaw(context.Background(), requests)

	Expect((<-requestChan).body).To(Equal(`[{"method":"myMethod1","params":[1],"id":123,"jsonrpc":"7.0"},` +
		`{"method":"myMethod2","params":{"name":"Alex","age":35,"country":"Germany"},"id":321,"jsonrpc":"wrong"}]`))
}

// test if the result of an an rpc request is parsed correctly and if errors are thrown correctly
func TestRpcJsonResponseStruct(t *testing.T) {
	RegisterTestingT(t)
	rpcClient := NewClient(httpServer.URL)

	// empty return body is an error
	responseBody = ``
	res, err := rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	Expect(err).NotTo(BeNil())
	Expect(res).To(BeNil())

	// not a json body is an error
	responseBody = `{ "not": "a", "json": "object"`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	Expect(err).NotTo(BeNil())
	Expect(res).To(BeNil())

	// field "anotherField" not allowed in rpc response is an error
	responseBody = `{ "anotherField": "norpc"}`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	Expect(err).NotTo(BeNil())
	Expect(res).To(BeNil())

	// TODO: result must contain one of "result", "error"
	// TODO: is there an efficient way to do this?
	/*responseBody = `{}`
	res, err = rpcClient.Call("something", 1, 2, 3)
	<-requestChan
	Expect(err).NotTo(BeNil())
	Expect(res).To(BeNil())*/

	// result null is ok
	responseBody = `{"result": null}`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res.Result).To(BeNil())
	Expect(res.Error).To(BeNil())

	// error null is ok
	responseBody = `{"error": null}`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res.Result).To(BeNil())
	Expect(res.Error).To(BeNil())

	// result and error null is ok
	responseBody = `{"result": null, "error": null}`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res.Result).To(BeNil())
	Expect(res.Error).To(BeNil())

	// TODO: result must not contain both of "result", "error" != null
	// TODO: is there an efficient way to do this?
	/*responseBody = `{ "result": 123, "error": {"code": 123, "message": "something wrong"}}`
	res, err = rpcClient.Call("something", 1, 2, 3)
	<-requestChan
	Expect(err).NotTo(BeNil())
	Expect(res).To(BeNil())*/

	// result string is ok
	responseBody = `{"result": "ok"}`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res.Result).To(Equal(stdjson.RawMessage([]byte(strconv.Quote("ok")))))

	// result with error null is ok
	responseBody = `{"result": "ok", "error": null}`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res.Result).To(Equal(stdjson.RawMessage([]byte(strconv.Quote("ok")))))

	// error with result null is ok
	responseBody = `{"error": {"code": 123, "message": "something wrong"}, "result": null}`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res.Result).To(BeNil())
	Expect(res.Error.Code).To(Equal(123))
	Expect(res.Error.Message).To(Equal("something wrong"))

	// TODO: empty error is not ok, must at least contain code and message
	/*responseBody = `{ "error": {}}`
	res, err = rpcClient.Call("something", 1, 2, 3)
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res.Result).To(BeNil())
	Expect(res.Error).NotTo(BeNil())*/

	// TODO: only code in error is not ok, must at least contain code and message
	/*responseBody = `{ "error": {"code": 123}}`
	res, err = rpcClient.Call("something", 1, 2, 3)
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res.Result).To(BeNil())
	Expect(res.Error).NotTo(BeNil())*/

	// TODO: only message in error is not ok, must at least contain code and message
	/*responseBody = `{ "error": {"message": "something wrong"}}`
	res, err = rpcClient.Call("something", 1, 2, 3)
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res.Result).To(BeNil())
	Expect(res.Error).NotTo(BeNil())*/

	// error with code and message is ok
	responseBody = `{ "error": {"code": 123, "message": "something wrong"}}`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res.Result).To(BeNil())
	Expect(res.Error.Code).To(Equal(123))
	Expect(res.Error.Message).To(Equal("something wrong"))

	// check results

	var p *Person
	responseBody = `{ "result": {"name": "Alex", "age": 35, "anotherField": "something"} }`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res.Error).To(BeNil())
	err = res.GetObject(&p)
	Expect(err).To(BeNil())
	Expect(p.Name).To(Equal("Alex"))
	Expect(p.Age).To(Equal(35))
	Expect(p.Country).To(Equal(""))

	// TODO: How to check if result could be parsed or if it is default?
	p = nil
	responseBody = `{ "result": {"anotherField": "something"} }`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res.Error).To(BeNil())
	err = res.GetObject(&p)
	Expect(err).To(BeNil())
	Expect(p).NotTo(BeNil())

	// TODO: HERE######
	var pp *PointerFieldPerson
	responseBody = `{ "result": {"anotherField": "something", "country": "Germany"} }`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res.Error).To(BeNil())
	err = res.GetObject(&pp)
	Expect(err).To(BeNil())
	Expect(pp.Name).To(BeNil())
	Expect(pp.Age).To(BeNil())
	Expect(*pp.Country).To(Equal("Germany"))

	p = nil
	responseBody = `{ "result": null }`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res.Error).To(BeNil())
	err = res.GetObject(&p)
	Expect(err).To(BeNil())
	Expect(p).To(BeNil())

	// passing nil is an error
	// TODO
	// p = nil
	// responseBody = `{ "result": null }`
	// res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	// <-requestChan
	// Expect(err).To(BeNil())
	// Expect(res.Error).To(BeNil())
	// err = res.GetObject(p)
	// Expect(err).NotTo(BeNil())
	// Expect(p).To(BeNil())

	p2 := &Person{
		Name: "Alex",
	}
	responseBody = `{ "result": null }`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res.Error).To(BeNil())
	err = res.GetObject(&p2)
	Expect(err).To(BeNil())
	Expect(p2).To(BeNil())

	p2 = &Person{
		Name: "Alex",
	}
	responseBody = `{ "result": {"age": 35} }`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res.Error).To(BeNil())
	err = res.GetObject(p2)
	Expect(err).To(BeNil())
	Expect(p2.Name).To(Equal("Alex"))
	Expect(p2.Age).To(Equal(35))

	// prefilled struct is kept on no result
	p3 := Person{
		Name: "Alex",
	}
	responseBody = `{ "result": null }`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res.Error).To(BeNil())
	err = res.GetObject(&p3)
	Expect(err).To(BeNil())
	Expect(p3.Name).To(Equal("Alex"))

	// prefilled struct is extended / overwritten
	p3 = Person{
		Name: "Alex",
		Age:  123,
	}
	responseBody = `{ "result": {"age": 35, "country": "Germany"} }`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res.Error).To(BeNil())
	err = res.GetObject(&p3)
	Expect(err).To(BeNil())
	Expect(p3.Name).To(Equal("Alex"))
	Expect(p3.Age).To(Equal(35))
	Expect(p3.Country).To(Equal("Germany"))

	// nil is an error
	responseBody = `{ "result": {"age": 35} }`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res.Error).To(BeNil())
	err = res.GetObject(nil)
	Expect(err).NotTo(BeNil())
}

func TestRpcBatchJsonResponseStruct(t *testing.T) {
	RegisterTestingT(t)
	rpcClient := NewClient(httpServer.URL)

	// empty return body is an error
	responseBody = ``
	res, err := rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	Expect(err).NotTo(BeNil())
	Expect(res).To(BeNil())

	// not a json body is an error
	responseBody = `{ "not": "a", "json": "object"`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	Expect(err).NotTo(BeNil())
	Expect(res).To(BeNil())

	// field "anotherField" not allowed in rpc response is an error
	responseBody = `{ "anotherField": "norpc"}`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	Expect(err).NotTo(BeNil())
	Expect(res).To(BeNil())

	// TODO: result must contain one of "result", "error"
	// TODO: is there an efficient way to do this?
	/*responseBody = `[{}]`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	Expect(err).NotTo(BeNil())
	Expect(res).To(BeNil())*/

	// result must be wrapped in array on batch request
	responseBody = `{"result": null}`
	_, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	Expect(err.Error()).NotTo(BeNil())

	// result ok since in array
	responseBody = `[{"result": null}]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	Expect(err).To(BeNil())
	Expect(len(res)).To(Equal(1))
	Expect(res[0].Result).To(BeNil())

	// error null is ok
	responseBody = `[{"error": null}]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res[0].Result).To(BeNil())
	Expect(res[0].Error).To(BeNil())

	// result and error null is ok
	responseBody = `[{"result": null, "error": null}]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res[0].Result).To(BeNil())
	Expect(res[0].Error).To(BeNil())

	// TODO: result must not contain both of "result", "error" != null
	// TODO: is there an efficient way to do this?
	/*responseBody = `[{ "result": 123, "error": {"code": 123, "message": "something wrong"}}]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
	NewRequest("something",1, 2, 3),
	})
	<-requestChan
	Expect(err).NotTo(BeNil())
	Expect(res).To(BeNil())*/

	// result string is ok
	responseBody = `[{"result": "ok","id":1}]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res[0].Result).To(Equal(stdjson.RawMessage([]byte(strconv.Quote("ok")))))
	Expect(res[0].ID).To(Equal(stdjson.Number("1")))

	// result with error null is ok
	responseBody = `[{"result": "ok", "error": null}]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res[0].Result).To(Equal(stdjson.RawMessage([]byte(strconv.Quote("ok")))))

	// error with result null is ok
	responseBody = `[{"error": {"code": 123, "message": "something wrong"}, "result": null}]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res[0].Result).To(BeNil())
	Expect(res[0].Error.Code).To(Equal(123))
	Expect(res[0].Error.Message).To(Equal("something wrong"))

	// TODO: empty error is not ok, must at least contain code and message
	/*responseBody = `[{ "error": {}}]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
	NewRequest("something",1, 2, 3),
	})
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res[0].Result).To(BeNil())
	Expect(res[0].Error).NotTo(BeNil())*/ /*

		// TODO: only code in error is not ok, must at least contain code and message
	*/ /*responseBody = `[{ "error": {"code": 123}}]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
	NewRequest("something",1, 2, 3),
	})
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res[0].Result).To(BeNil())
	Expect(res[0].Error).NotTo(BeNil())*/ /*

		// TODO: only message in error is not ok, must at least contain code and message
	*/ /*responseBody = `[{ "error": {"message": "something wrong"}}]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
	NewRequest("something",1, 2, 3),
	})
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res[0].Result).To(BeNil())
	Expect(res[0].Error).NotTo(BeNil())*/

	// error with code and message is ok
	responseBody = `[{ "error": {"code": 123, "message": "something wrong"}}]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res[0].Result).To(BeNil())
	Expect(res[0].Error.Code).To(Equal(123))
	Expect(res[0].Error.Message).To(Equal("something wrong"))

	// check results

	var p *Person
	responseBody = `[{"id":1, "result": {"name": "Alex", "age": 35}}, {"id":2, "result": {"name": "Lena", "age": 2}}]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})

	<-requestChan
	Expect(err).To(BeNil())

	Expect(res[0].Error).To(BeNil())
	Expect(res[0].ID).To(Equal(stdjson.Number("1")))

	Expect(res[1].Error).To(BeNil())
	Expect(res[1].ID).To(Equal(stdjson.Number("2")))

	err = res[0].GetObject(&p)
	require.NoError(t, err)
	Expect(p.Name).To(Equal("Alex"))
	Expect(p.Age).To(Equal(35))

	err = res[1].GetObject(&p)
	require.NoError(t, err)
	Expect(p.Name).To(Equal("Lena"))
	Expect(p.Age).To(Equal(2))

	// check if error occurred
	responseBody = `[{ "result": "someresult", "error": null}, { "result": null, "error": {"code": 123, "message": "something wrong"}}]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res.HasError()).To(BeTrue())

	// check if error occurred
	responseBody = `[{ "result": null, "error": {"code": 123, "message": "something wrong"}}]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res.HasError()).To(BeTrue())
	// check if error occurred
	responseBody = `[{ "result": null, "error": {"code": 123, "message": "something wrong"}}]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res.HasError()).To(BeTrue())

	// check if response mapping works
	responseBody = `[{ "id":123,"result": 123},{ "id":1,"result": 1}]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res.HasError()).To(BeFalse())
	resMap := res.AsMap()

	var int1 int64
	require.NotNil(t, resMap[1])
	resMap[1].GetObject(&int1)
	var int123 int64
	require.NotNil(t, resMap[123])
	resMap[123].GetObject(&int123)
	Expect(int1).To(Equal(int64(1)))
	Expect(int123).To(Equal(int64(123)))

	// check if getByID works
	res.GetByID(123).GetObject(&int123)
	Expect(int123).To(Equal(int64(123)))

	// check if error occurred
	responseBody = `[{ "result": null, "error": {"code": 123, "message": "something wrong"}}]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	Expect(err).To(BeNil())
	Expect(res.HasError()).To(BeTrue())

	/*
		// TODO: How to check if result could be parsed or if it is default?
		p = nil
		responseBody = `{ "result": {"anotherField": "something"} }`
		res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
			{"something", Params(1, 2, 3)},
		})
		<-requestChan
		Expect(err).To(BeNil())
		Expect(res.Error).To(BeNil())
		err = res.GetObject(&p)
		Expect(err).To(BeNil())
		Expect(p).NotTo(BeNil())

		// TODO: HERE######
		var pp *PointerFieldPerson
		responseBody = `{ "result": {"anotherField": "something", "country": "Germany"} }`
		res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
			{"something", Params(1, 2, 3)},
		})
		<-requestChan
		Expect(err).To(BeNil())
		Expect(res.Error).To(BeNil())
		err = res.GetObject(&pp)
		Expect(err).To(BeNil())
		Expect(pp.Name).To(BeNil())
		Expect(pp.Age).To(BeNil())
		Expect(*pp.Country).To(Equal("Germany"))

		p = nil
		responseBody = `{ "result": null }`
		res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
			{"something", Params(1, 2, 3)},
		})
		<-requestChan
		Expect(err).To(BeNil())
		Expect(res.Error).To(BeNil())
		err = res.GetObject(&p)
		Expect(err).To(BeNil())
		Expect(p).To(BeNil())

		// passing nil is an error
		p = nil
		responseBody = `{ "result": null }`
		res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
			{"something", Params(1, 2, 3)},
		})
		<-requestChan
		Expect(err).To(BeNil())
		Expect(res.Error).To(BeNil())
		err = res.GetObject(p)
		Expect(err).NotTo(BeNil())
		Expect(p).To(BeNil())

		p2 := &Person{
			Name: "Alex",
		}
		responseBody = `{ "result": null }`
		res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
			{"something", Params(1, 2, 3)},
		})
		<-requestChan
		Expect(err).To(BeNil())
		Expect(res.Error).To(BeNil())
		err = res.GetObject(&p2)
		Expect(err).To(BeNil())
		Expect(p2).To(BeNil())

		p2 = &Person{
			Name: "Alex",
		}
		responseBody = `{ "result": {"age": 35} }`
		res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
			{"something", Params(1, 2, 3)},
		})
		<-requestChan
		Expect(err).To(BeNil())
		Expect(res.Error).To(BeNil())
		err = res.GetObject(p2)
		Expect(err).To(BeNil())
		Expect(p2.Name).To(Equal("Alex"))
		Expect(p2.Age).To(Equal(35))

		// prefilled struct is kept on no result
		p3 := Person{
			Name: "Alex",
		}
		responseBody = `{ "result": null }`
		res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
			{"something", Params(1, 2, 3)},
		})
		<-requestChan
		Expect(err).To(BeNil())
		Expect(res.Error).To(BeNil())
		err = res.GetObject(&p3)
		Expect(err).To(BeNil())
		Expect(p3.Name).To(Equal("Alex"))

		// prefilled struct is extended / overwritten
		p3 = Person{
			Name: "Alex",
			Age:  123,
		}
		responseBody = `{ "result": {"age": 35, "country": "Germany"} }`
		res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
			{"something", Params(1, 2, 3)},
		})
		<-requestChan
		Expect(err).To(BeNil())
		Expect(res.Error).To(BeNil())
		err = res.GetObject(&p3)
		Expect(err).To(BeNil())
		Expect(p3.Name).To(Equal("Alex"))
		Expect(p3.Age).To(Equal(35))
		Expect(p3.Country).To(Equal("Germany"))

		// nil is an error
		responseBody = `{ "result": {"age": 35} }`
		res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
			{"something", Params(1, 2, 3)},
		})
		<-requestChan
		Expect(err).To(BeNil())
		Expect(res.Error).To(BeNil())
		err = res.GetObject(nil)
		Expect(err).NotTo(BeNil())
	*/
}

func TestRpcClient_CallFor(t *testing.T) {
	RegisterTestingT(t)
	rpcClient := NewClient(httpServer.URL)

	i := 0
	responseBody = `{"result":3,"id":1,"jsonrpc":"2.0"}`
	err := rpcClient.CallFor(context.Background(), &i, "something", 1, 2, 3)
	<-requestChan
	Expect(err).To(BeNil())
	Expect(i).To(Equal(3))

	/*
		i = 3
		responseBody = `{"result":null,"id":1,"jsonrpc":"2.0"}`
		err = rpcClient.CallFor(&i, "something", 1, 2, 3)
		<-requestChan
		Expect(err).To(BeNil())
		// i is not modified when result is empty since null (nil) value cannot be stored in int
		Expect(i).To(Equal(3))

		var pi *int
		responseBody = `{"result":4,"id":1,"jsonrpc":"2.0"}`
		err = rpcClient.CallFor(pi, "something", 1, 2, 3)
		<-requestChan
		Expect(err).NotTo(BeNil())
		Expect(pi).To(BeNil())

		responseBody = `{"result":4,"id":1,"jsonrpc":"2.0"}`
		err = rpcClient.CallFor(&pi, "something", 1, 2, 3)
		<-requestChan
		Expect(err).To(BeNil())
		Expect(*pi).To(Equal(4))

		*pi = 3
		responseBody = `{"result":null,"id":1,"jsonrpc":"2.0"}`
		err = rpcClient.CallFor(&pi, "something", 1, 2, 3)
		<-requestChan
		Expect(err).To(BeNil())
		// since pi has a value it is not overwritten by null result
		Expect(pi).To(BeNil())

		p := &Person{}
		responseBody = `{"result":null,"id":1,"jsonrpc":"2.0"}`
		err = rpcClient.CallFor(p, "something", 1, 2, 3)
		<-requestChan
		Expect(err).To(BeNil())
		// p is not changed since it has a value and result is null
		Expect(p).NotTo(BeNil())

		var p2 *Person
		responseBody = `{"result":null,"id":1,"jsonrpc":"2.0"}`
		err = rpcClient.CallFor(p2, "something", 1, 2, 3)
		<-requestChan
		Expect(err).NotTo(BeNil())
		// p is not changed since it has a value and result is null
		Expect(p2).To(BeNil())

		p3 := Person{}
		responseBody = `{"result":null,"id":1,"jsonrpc":"2.0"}`
		err = rpcClient.CallFor(&p3, "something", 1, 2, 3)
		<-requestChan
		Expect(err).To(BeNil())
		// p is not changed since it has a value and result is null
		Expect(p).NotTo(BeNil())

		p = &Person{Age: 35}
		responseBody = `{"result":{"name":"Alex"},"id":1,"jsonrpc":"2.0"}`
		err = rpcClient.CallFor(p, "something", 1, 2, 3)
		<-requestChan
		Expect(err).To(BeNil())
		// p is not changed since it has a value and result is null
		Expect(p.Name).To(Equal("Alex"))
		Expect(p.Age).To(Equal(35))

		p2 = nil
		responseBody = `{"result":{"name":"Alex"},"id":1,"jsonrpc":"2.0"}`
		err = rpcClient.CallFor(p2, "something", 1, 2, 3)
		<-requestChan
		Expect(err).NotTo(BeNil())
		// p is not changed since it has a value and result is null
		Expect(p2).To(BeNil())

		p2 = nil
		responseBody = `{"result":{"name":"Alex"},"id":1,"jsonrpc":"2.0"}`
		err = rpcClient.CallFor(&p2, "something", 1, 2, 3)
		<-requestChan
		Expect(err).To(BeNil())
		// p is not changed since it has a value and result is null
		Expect(p2).NotTo(BeNil())
		Expect(p2.Name).To(Equal("Alex"))

		p3 = Person{Age: 35}
		responseBody = `{"result":{"name":"Alex"},"id":1,"jsonrpc":"2.0"}`
		err = rpcClient.CallFor(&p3, "something", 1, 2, 3)
		<-requestChan
		Expect(err).To(BeNil())
		// p is not changed since it has a value and result is null
		Expect(p.Name).To(Equal("Alex"))
		Expect(p.Age).To(Equal(35))

		p3 = Person{Age: 35}
		responseBody = `{"result":{"name":"Alex"},"id":1,"jsonrpc":"2.0"}`
		err = rpcClient.CallFor(&p3, "something", 1, 2, 3)
		<-requestChan
		Expect(err).To(BeNil())
		// p is not changed since it has a value and result is null
		Expect(p.Name).To(Equal("Alex"))
		Expect(p.Age).To(Equal(35))

		var intArray []int
		responseBody = `{"result":[1, 2, 3],"id":1,"jsonrpc":"2.0"}`
		err = rpcClient.CallFor(&intArray, "something", 1, 2, 3)
		<-requestChan
		Expect(err).To(BeNil())
		// p is not changed since it has a value and result is null
		Expect(intArray).To(ContainElement(1))
		Expect(intArray).To(ContainElement(2))
		Expect(intArray).To(ContainElement(3))*/
}

type Person struct {
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Country string `json:"country"`
}

type PointerFieldPerson struct {
	Name    *string `json:"name"`
	Age     *int    `json:"age"`
	Country *string `json:"country"`
}

type Drink struct {
	Name        string   `json:"name"`
	Ingredients []string `json:"ingredients"`
}
