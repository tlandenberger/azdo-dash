package data

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type PullRequestData struct {
	ID                 int
	Title              string
	Status             string
	MergeStatus        string
	SourceBranch       string
	CreatedBy          string
	IsDraft            bool
	RepositoryName     string
	RepositoryID       string
	IsRequiredReviewer bool
	Vote               int
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
	Title         string             `json:"title"`
	Status        string             `json:"status"`
	IsDraft       bool               `json:"isDraft"`
	CreatedBy     UserResponse       `json:"createdBy"`
	Reviewers     []ReviewerResponse `json:"reviewers"`
	SourceRefName string             `json:"sourceRefName"`
	TargetRefName string             `json:"targetRefName"`
	MergeStatus   string             `json:"mergeStatus"`
}

type RepositoryResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UserResponse struct {
	DisplayName string `json:"displayName"`
	UniqueName  string `json:"uniqueName"`
}

type ReviewerResponse struct {
	DisplayName string `json:"displayName"`
	UniqueName  string `json:"uniqueName"`
	Vote        int    `json:"vote"`
	HasDeclined bool   `json:"hasDeclined"`
	IsRequired  bool   `json:"isRequired"`
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

		var userReviewer *ReviewerResponse = nil
		for _, reviewer := range prResponse.Reviewers {
			if reviewer.DisplayName == "Tobias Landenberger" {
				userReviewer = &reviewer
				break
			}
		}

		vote := 0
		isRequiredReviewer := false
		if userReviewer != nil {
			vote = userReviewer.Vote
			isRequiredReviewer = userReviewer.IsRequired
		}

		result = append(result, PullRequestData{
			ID:                 prResponse.PullRequestID,
			Title:              prResponse.Title,
			CreatedBy:          prResponse.CreatedBy.DisplayName,
			Status:             prResponse.Status,
			MergeStatus:        prResponse.MergeStatus,
			IsDraft:            prResponse.IsDraft,
			RepositoryID:       prResponse.Repository.ID,
			RepositoryName:     prResponse.Repository.Name,
			IsRequiredReviewer: isRequiredReviewer,
			SourceBranch:       prResponse.SourceRefName,
			Vote:               vote,
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
