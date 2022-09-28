package util

import (
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/go-resty/resty/v2"
)

var etherscanAPIKey = os.Getenv("EHTERSCAN_API_KEY")

// taken from https://gist.github.com/crazygit/9279a3b26461d7cb03e807a6362ec855
type RawABIResponse struct {
	Status  *string `json:"status"`
	Message *string `json:"message"`
	Result  *string `json:"result"`
}

func GetContractRawABI(address string, apiKey string) (*RawABIResponse, error) {
	client := resty.New()
	rawABIResponse := &RawABIResponse{}
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

func GetContractABI(contractAddress string) (*abi.ABI, error) {
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
