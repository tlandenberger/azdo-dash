package data

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type PullRequestData struct {
	ID             int
	Status         string
	RepositoryName string
	RepositoryID   string
}

type FetchPRRequest struct {
	OrgName             string
	ProjectID           string
	RepoID              string
	PersonalAccessToken string
}

type PullRequests struct {
	Prs        []PullRequestData
	TotalCount int
}

type FetchPRResponse struct {
	Value []PullRequestResponse `json:"value"`
	Count int                   `json:"count"`
}

type PullRequestResponse struct {
	Repository    RepositoryResponse `json:"repository"`
	PullRequestID int                `json:"pullRequestId"`
	Status        string             `json:"status"`
}

type RepositoryResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func FetchPullRequests(configs []FetchPRRequest) (PullRequests, error) {
	prs := make([]PullRequestData, 0)

	for _, config := range configs {
		response, err := FetchPullRequestsByProject(config)
		if err != nil {
			return PullRequests{}, fmt.Errorf("fetching pull request: %w", err)
		}

		if response != nil {
			projectPrs := getPullRequestData(response)
			prs = append(prs, projectPrs...)
		}
	}

	return PullRequests{prs, len(prs)}, nil
}

func getPullRequestData(response *FetchPRResponse) []PullRequestData {
	result := make([]PullRequestData, 0)

	for _, prResponse := range response.Value {
		result = append(result, PullRequestData{
			ID:             prResponse.PullRequestID,
			Status:         prResponse.Status,
			RepositoryID:   prResponse.Repository.ID,
			RepositoryName: prResponse.Repository.Name,
		})
	}

	return result
}

func FetchPullRequestsByProject(config FetchPRRequest) (*FetchPRResponse, error) {
	url := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/git/repositories/%s/pullrequests?api-version=6.0", config.OrgName, config.ProjectID, config.RepoID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Add the personal access token for authentication
	auth := "any:" + config.PersonalAccessToken
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Set("Authorization", "Basic "+encodedAuth)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		return nil, fmt.Errorf("error response: %s", bodyString)
	}

	var response FetchPRResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &response, nil
}
