package conch_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/require"
)

func TestInfoClient_Get(t *testing.T) {
	cli := &wrapClient{Client: retryablehttp.NewClient()}

	c := InfoClient{
		HTTPDoerWithContext: cli,
	}

	s := time.Now()
	info, err := c.Get(context.Background(), "54329")
	fmt.Println(time.Since(s))
	require.NoError(t, err)

	bb, err := json.Marshal(info)
	require.NoError(t, err)
	fmt.Println(string(bb))
}

type HTTPDoerWithContext interface {
	Do(*http.Request) (*http.Response, error)
}

type InfoClient struct {
	HTTPDoerWithContext
}

const URLModel = "https://geo.api.gouv.fr/communes/92035?fields=nom,code,codesPostaux,siren,centre,surface,contour,mairie,bbox,codeEpci,epci,codeDepartement,departement,codeRegion,region,population,zone&format=json&geometry=centre"

func (cli *InfoClient) Get(ctx context.Context, inseeCode Param) (
	*Resp,
	error,
) {
	u, err := url.Parse(URLModel)
	if err != nil {
		return nil, err
	}
	u.Path = fmt.Sprintf("communes/%s", string(inseeCode))

	finalURL := u.String()
	fmt.Println("URL:", finalURL)

	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		finalURL,
		nil,
	)
	if err != nil {
		return nil, err
	}

	do, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer do.Body.Close()

	var r Resp

	err = json.NewDecoder(do.Body).Decode(&r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

// https://geo.api.gouv.fr/communes/92035?fields=nom,code,codesPostaux,siren,centre,surface,contour,mairie,bbox,codeEpci,epci,codeDepartement,departement,codeRegion,region,population,zone&format=json&geometry=centre

type Resp struct {
	Nom          string   `json:"nom"`
	Code         string   `json:"code"`
	CodesPostaux []string `json:"codesPostaux"`
	Siren        string   `json:"siren"`
	Centre       struct {
		Type        string    `json:"type"`
		Coordinates []float64 `json:"coordinates"`
	} `json:"centre"`
	Surface float64 `json:"surface"`
	Contour struct {
		Type        string        `json:"type"`
		Coordinates [][][]float64 `json:"coordinates"`
	} `json:"contour"`
	Mairie struct {
		Type        string    `json:"type"`
		Coordinates []float64 `json:"coordinates"`
	} `json:"mairie"`
	Bbox struct {
		Type        string        `json:"type"`
		Coordinates [][][]float64 `json:"coordinates"`
	} `json:"bbox"`
	CodeEpci        string `json:"codeEpci"`
	CodeDepartement string `json:"codeDepartement"`
	CodeRegion      string `json:"codeRegion"`
	Population      int    `json:"population"`
	Zone            string `json:"zone"`
	Epci            struct {
		Code string `json:"code"`
		Nom  string `json:"nom"`
	} `json:"epci"`
	Departement struct {
		Code string `json:"code"`
		Nom  string `json:"nom"`
	} `json:"departement"`
	Region struct {
		Code string `json:"code"`
		Nom  string `json:"nom"`
	} `json:"region"`
}
