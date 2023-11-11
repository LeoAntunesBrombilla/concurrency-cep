package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type ViaCep struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
}

type CdnCep struct {
	Cep        string `json:"code"`
	Estado     string `json:"state"`
	Cidade     string `json:"city"`
	Bairro     string `json:"district"`
	Logradouro string `json:"address"`
}

type ViaCepResponse struct {
	Data  *ViaCep
	Error error
}

type CdnCepResponse struct {
	Data  *CdnCep
	Error error
}

func main() {
	http.HandleFunc("/", searchCep)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}
}

func searchCep(writer http.ResponseWriter, request *http.Request) {
	c1 := make(chan ViaCepResponse)
	c2 := make(chan CdnCepResponse)
	timeout := time.NewTimer(1 * time.Second)

	if request.URL.Path != "/" {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	reqParam := request.URL.Query().Get("cep")
	if reqParam == "" {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	go func() {
		c1 <- BuscaViaCep(reqParam)
	}()

	go func() {
		c2 <- BuscaCdnCep(reqParam)
	}()

	select {
	case cep := <-c1:
		if cep.Data.Cep == "" {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		fmt.Printf("ViaCep")
		writer.Header().Set("Content-Type", "applications/json")
		writer.WriteHeader(http.StatusOK)
		json.NewEncoder(writer).Encode(cep)
		return
	case cep := <-c2:
		if cep.Data.Cep == "" {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		fmt.Printf("CdnCep")
		writer.Header().Set("Content-Type", "applications/json")
		writer.WriteHeader(http.StatusOK)
		json.NewEncoder(writer).Encode(cep)
		return
	case <-timeout.C:
		writer.WriteHeader(http.StatusRequestTimeout)
		return
	}

}

func BuscaViaCep(cep string) ViaCepResponse {
	resp, err := http.Get("https://viacep.com.br/ws/" + cep + "/json/")
	if err != nil {
		return ViaCepResponse{Data: nil, Error: err}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ViaCepResponse{Data: nil, Error: err}
	}
	var c ViaCep
	err = json.Unmarshal(body, &c)
	if err != nil {
		return ViaCepResponse{Data: nil, Error: err}
	}
	return ViaCepResponse{Data: &c, Error: err}
}

func BuscaCdnCep(cep string) CdnCepResponse {
	formattedCep := cep[:5] + "-" + cep[5:]
	resp, err := http.Get("https://cdn.apicep.com/file/apicep/" + formattedCep + ".json")
	if err != nil {
		return CdnCepResponse{Data: nil, Error: err}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return CdnCepResponse{Data: nil, Error: err}
	}
	var c CdnCep
	err = json.Unmarshal(body, &c)
	if err != nil {
		return CdnCepResponse{Data: nil, Error: err}
	}
	return CdnCepResponse{Data: &c, Error: err}
}
