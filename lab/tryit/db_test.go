package conch_test

import (
	"embed"
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestGetRandomINSEECodes(t *testing.T) {
	for _, i := range GetRandomINSEECodes(100) {
		fmt.Println(i)
	}
}

type Commune struct {
	Nom        string `json:"nom_de_la_commune"`
	CodePostal string `json:"code_postal"`
	CodeINSEE  string `json:"code_commune_insee"`
}

type Communes []*Commune

//go:embed communes.json
var fs embed.FS

var communes Communes

func init() {
	f, err := fs.Open("communes.json")
	if err != nil {
		panic(err)
	}

	if err := json.NewDecoder(f).Decode(&communes); err != nil {
		panic(err)
	}
}

func GetRandomINSEECodes(count int) []string {
	var result []string

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	rnd.Shuffle(
		len(communes), func(i, j int) {
			communes[i], communes[j] = communes[j], communes[i]
		},
	)

	for i := 0; i < count; i++ {
		result = append(result, communes[i].CodeINSEE)
	}

	return result
}
