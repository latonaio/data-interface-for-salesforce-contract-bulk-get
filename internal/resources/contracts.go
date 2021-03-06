package resources

import (
	"errors"
	"fmt"
)

// Contract struct
type Contract struct {
	method   string
	metadata map[string]interface{}
}

func (c *Contract) objectName() string {
	const obName = "ContractRelatedList"
	return obName
}

// newContract writes that new Customer instance
func NewContract(metadata map[string]interface{}) (*Contract, error) {
	rawMethod, ok := metadata["method"]
	if !ok {
		return nil, errors.New("missing required parameters: method")
	}
	method, ok := rawMethod.(string)
	if !ok {
		return nil, errors.New("failed to convert interface{} to string")
	}
	return &Contract{
		method:   method,
		metadata: metadata,
	}, nil
}

// getMetadata mold customer get metadata
func (c *Contract) getMetadata() (map[string]interface{}, error) {
	return buildMetadata(c.method, c.objectName(), "", nil, ""), nil
}

// BuildMetadata
func (c *Contract) BuildMetadata() (map[string]interface{}, error) {
	switch c.method {
	case "get":
		return c.getMetadata()
	}
	return nil, fmt.Errorf("invalid method: %s", c.method)
}

func buildMetadata(method, object, pathParam string, queryParams map[string]string, body string) map[string]interface{} {
	metadata := map[string]interface{}{"method": method, "object": object}
	if len(pathParam) > 0 {
		metadata["path_param"] = pathParam
	}
	if queryParams != nil {
		metadata["query_params"] = queryParams
	}
	if body != "" {
		metadata["body"] = body
	}
	metadata["connection_key"] = "contract_bulk_get"
	return metadata
}