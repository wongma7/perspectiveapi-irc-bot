package perspective

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/mitchellh/mapstructure"
	"github.com/wongma7/perspectiveapi-irc-bot/pkg/toxic"
)

const (
	modelToxicity = "TOXICITY"
	baseURL       = "https://commentanalyzer.googleapis.com/v1alpha1/comments:analyze"
)

type Request struct {
	Comment             Comment          `json:"comment"`
	RequestedAttributes map[string]Score `json:"requestedAttributes"`
	Languages           []string         `json:"languages",omitempty"`
	DoNotStore          bool             `json:"doNotStore,omitempty"`
}

type Comment struct {
	Text  string `json:"text"`
	Ctype string `json:"type,omitempty"`
}

type Score struct {
	ScoreType      string  `json:"scoreType,omitempty"`
	ScoreThreshold float64 `json:"scoreThreshold,omitempty"`
}

type Response struct {
	AttributeScores map[string]Scores
}

type Scores struct {
	SummaryScore SummaryScore `json:"summaryScore"`
}

type SummaryScore struct {
	Value float64 `json:"value"`
	Stype string  `json:"type"`
}

type Perspective struct {
}

var _ toxic.Analyzer = &Perspective{}

func (p *Perspective) ScoreComment(comment string) (float64, error) {
	request := &Request{
		Comment: Comment{
			Text: comment,
		},
		RequestedAttributes: map[string]Score{
			modelToxicity: {},
		},
		Languages:  []string{"en"},
		DoNotStore: true,
	}

	requ, err := json.Marshal(request)
	if err != nil {
		return 0, err
	}

	apiKey := os.Getenv("PERSPECTIVE_API_KEY")
	url := baseURL + "?key=" + apiKey
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requ))
	if err != nil {
		return 0, err
	}

	var f interface{}
	json.NewDecoder(resp.Body).Decode(&f)
	m := f.(map[string]interface{})
	for k, v := range m {
		switch k {
		case "error":
			return 0, fmt.Errorf("%v", v)
		case "attributeScores":
			var response Response
			err = mapstructure.Decode(m, &response)
			if err != nil {
				return 0, err
			}

			return response.AttributeScores[modelToxicity].SummaryScore.Value, nil
		}
	}

	return 0, nil
}
