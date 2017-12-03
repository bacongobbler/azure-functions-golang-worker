package executor

import (
	"fmt"
	"plugin"

	log "github.com/Sirupsen/logrus"
	"github.com/radu-matei/azure-functions-golang-worker/rpc"
)

var (
	functionMap = make(map[string]interface{})
)

// LoadMethod takes a .so object from the function's bin directory and loads it
func LoadMethod(request *rpc.FunctionLoadRequest) error {

	path := fmt.Sprintf("%s/bin/%s.so", request.Metadata.Directory, request.Metadata.Name)
	plugin, err := plugin.Open(path)
	if err != nil {
		log.Debugf("cannot get .so object from path %s: %v", path, err)
		return err
	}

	symbol, err := plugin.Lookup(request.Metadata.EntryPoint)
	if err != nil {
		log.Debugf("cannot look up symbol for entrypoint function %s: %v", request.Metadata.EntryPoint, err)
	}
	f, ok := symbol.(func(*rpc.RpcHttp) []byte)
	if !ok {
		log.Debug("incorrect function signature")
		return err
	}

	functionMap[request.FunctionId] = f

	return nil
}

// ExecuteMethod takes an InvocationRequest and executes the function with corresponding function ID
func ExecuteMethod(request *rpc.InvocationRequest) (response *rpc.InvocationResponse) {
	var output []byte

	switch r := request.TriggerMetadata["req"].Data.(type) {
	case *rpc.TypedData_Http:
		output = functionMap[request.FunctionId].(func(*rpc.RpcHttp) []byte)(r.Http)
	}

	log.Debugf("received output: %v", output)

	invocationResponse := &rpc.InvocationResponse{
		InvocationId: request.InvocationId,
		Result: &rpc.StatusResult{
			Status: rpc.StatusResult_Success,
		},
		ReturnValue: &rpc.TypedData{
			Data: &rpc.TypedData_Json{
				Json: string(output),
			},
		},
	}

	return invocationResponse
}
