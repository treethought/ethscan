package util

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/go-resty/resty/v2"
)

// taken from https://gist.github.com/crazygit/9279a3b26461d7cb03e807a6362ec855
type rawABIResponse struct {
	Status  *string `json:"status"`
	Message *string `json:"message"`
	Result  *string `json:"result"`
}

type rawContractData struct {
	SourceCode           string
	ABI                  string
	ContractName         string
	CompilerVersion      string
	OptimizationUsed     string
	Runs                 string
	ConstructorArguments string
	EVMVersion           string
	Library              string
	LicenseType          string
	Proxy                string
	Implementation       string
	SwarmSource          string
}

type ContractData struct {
	SourceCode           string
	ABI                  *abi.ABI
	ContractName         string
	CompilerVersion      string
	OptimizationUsed     bool
	Runs                 int
	ConstructorArguments string
	EVMVersion           string
	Library              string
	LicenseType          string
	Proxy                bool
	Implementation       string
	SwarmSource          string
}

type contractDataResponse struct {
	Status  string
	Message string
	Result  []rawContractData
	// Result *
}

func GetContractRawABI(address string, apiKey string) (*rawABIResponse, error) {
	client := resty.New()
	rawABIResponse := &rawABIResponse{}
	resp, err := client.R().
		SetQueryParams(map[string]string{
			"module":  "contract",
			"action":  "getabi",
			"address": address,
			"apikey":  apiKey,
		}).
		SetResult(rawABIResponse).
		Get("https://api.etherscan.io/api")

	if err != nil {
		return nil, err
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf(fmt.Sprintf("Get contract raw abi failed: %s\n", resp))
	}
	if *rawABIResponse.Status != "1" {
		return nil, fmt.Errorf(fmt.Sprintf("Get contract raw abi failed: %s\n", *rawABIResponse.Result))
	}

	return rawABIResponse, nil
}

func GetContractABI(contractAddress string, etherscanAPIKey string) (*abi.ABI, error) {
	rawABIResponse, err := GetContractRawABI(contractAddress, etherscanAPIKey)
	if err != nil {
		return nil, err
	}

	contractABI, err := abi.JSON(strings.NewReader(*rawABIResponse.Result))
	if err != nil {
		return nil, err
	}
	return &contractABI, nil
}

// refer
// https://github.com/ethereum/web3.py/blob/master/web3/contract.py#L435
func DecodeTransactionInputData(contractABI *abi.ABI, data []byte) (mehtod string, inputs map[string]interface{}, err error) {
	methodSigData := data[:4]
	inputsSigData := data[4:]
	method, err := contractABI.MethodById(methodSigData)
	if err != nil {
		return "", nil, err
	}
	inputsMap := make(map[string]interface{})
	if err := method.Inputs.UnpackIntoMap(inputsMap, inputsSigData); err != nil {
		return "", nil, err
	}
	return method.Name, inputsMap, nil
}

func getRawContractData(address string, apiKey string) (*contractDataResponse, error) {
	client := resty.New()
	rawResponse := &contractDataResponse{}
	resp, err := client.R().
		SetQueryParams(map[string]string{
			"module":  "contract",
			"action":  "getsourcecode",
			"address": address,
			"apikey":  apiKey,
		}).
		SetResult(rawResponse).
		Get("https://api.etherscan.io/api")

	if err != nil {
		return nil, err
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf(fmt.Sprintf("Get contract source code failed: %s\n", resp))
	}
	if rawResponse.Status != "1" {
		return nil, fmt.Errorf(fmt.Sprintf("Get contract source code failed: %s\n", *&rawResponse.Result))
	}

	return rawResponse, nil
}

func GetContractData(address, apiKey string) (*ContractData, error) {
	contractResp, err := getRawContractData(address, apiKey)
	if err != nil {
		return nil, err
	}

	if len(contractResp.Result) == 0 {
		return nil, fmt.Errorf("empty result from etherscan")
	}

	result := contractResp.Result[0]

	contractABI, err := abi.JSON(strings.NewReader(result.ABI))
	if err != nil {
		return nil, err
	}

	optBool, _ := strconv.ParseBool(result.OptimizationUsed)
	runsInt, _ := strconv.Atoi(result.Runs)
	proxyBool, _ := strconv.ParseBool(result.Proxy)

	c := &ContractData{
		SourceCode:           result.SourceCode,
		ABI:                  &contractABI,
		ContractName:         result.ContractName,
		CompilerVersion:      result.CompilerVersion,
		OptimizationUsed:     optBool,
		Runs:                 runsInt,
		ConstructorArguments: result.ConstructorArguments,
		EVMVersion:           result.EVMVersion,
		Library:              result.Library,
		LicenseType:          result.LicenseType,
		Proxy:                proxyBool,
		Implementation:       result.Implementation,
		SwarmSource:          result.SwarmSource,
	}
	return c, nil
}
