package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
)

const baseURL = "https://www.4byte.directory"

type signatureResponse struct {
	Count    int         `json:"count,omitempty"`
	Next     string      `json:"next,omitempty"`
	Previous string      `json:"previous,omitempty"`
	Results  []Signature `json:"results,omitempty"`
}

type Signature struct {
	ID             int    `json:"id,omitempty"`
	TextSignature  string `json:"text_signature,omitempty"`
	BytesSignature string `json:"bytes_signature,omitempty"`
	HexSignature   string `json:"hex_signature,omitempty"`
}

type SignatureDB struct {
	client  *http.Client
	baseURL string
}

func NewSignatureDB() *SignatureDB {
	return &SignatureDB{
		client:  http.DefaultClient,
		baseURL: baseURL,
	}

}

func (db *SignatureDB) GetSignature(hex string) (*Signature, error) {
	endpoint := fmt.Sprintf("%s/api/v1/signatures/", db.baseURL)

	url, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	params := url.Query()
	params.Add("hex_signature", hex)
	url.RawQuery = params.Encode()

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("ContentType", "application/json")
	resp, err := db.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("request failed")
	}

	sigResp := &signatureResponse{}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	ioutil.WriteFile("sigresp.json", data, 0755)
	err = json.Unmarshal(data, sigResp)
	if err != nil {
		return nil, err
	}

	if sigResp.Count < 1 {
		return nil, errors.New("not found")
	}

	var result *Signature
	// TODO: do this better
	// we take lowest ID to try to get the most likely signature
	// since there may be duplicates

	// {
	//   "id": 844488,
	//   "created_at": "2022-09-02T10:24:06.656097Z",
	//   "text_signature": "join_tg_invmru_haha_b76d2fb(address,address,uint256)",
	//   "hex_signature": "0xd78ad95f",
	//   "bytes_signature": "×Ù_"
	// },
	// {
	//   "id": 645993,
	//   "created_at": "2022-04-22T15:29:38.762772Z",
	//   "text_signature": "Swap(address,uint256,uint256,uint256,uint256,address)",
	//   "hex_signature": "0xd78ad95f",
	//   "bytes_signature": "×Ù_"
	// }

	lowestId := math.MaxInt
	for _, s := range sigResp.Results {
		if s.ID < lowestId {
			lowestId = s.ID
			result = &s
		}
	}

	return result, nil

}
