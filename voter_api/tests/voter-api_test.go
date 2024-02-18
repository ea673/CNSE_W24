package api

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	fake "github.com/brianvoe/gofakeit/v6" //aliasing package name
	"github.com/ea673/voter-api/voter"
	"github.com/go-resty/resty/v2"
	"github.com/google/go-cmp/cmp"
)

var (
	DEFAULT_HOST = "localhost"
	DEFAULT_PORT = "8080"
	base_url     string
	cli          = resty.New()
)

func TestMain(m *testing.M) {
	// Configure the base api url
	host := os.Getenv("HOST")
	if host == "" {
		host = DEFAULT_HOST
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = DEFAULT_PORT
	}

	base_url = fmt.Sprintf("http://%s:%s", host, port)

	// Ensure that the server is running
	_, err := cli.R().Get(base_url + "/voters/health")
	if err != nil {
		log.Fatalln("Server is not running on localhost:8080. Please start the server and try again.")
	}

	// Run the tests
	code := m.Run()

	// teardown
	os.Exit(code)
}

func setupSuite(t *testing.T) func(tb testing.T) {
	log.Println("setup suite")

	// Ensure we start with an empty voter list
	_, err := cli.R().Delete(base_url + "/voters")
	if err != nil {
		log.Fatalf("Error deleting voters: %v", err)
	}

	return nil
}

func TestVoterApi_AddVoterHandler(t *testing.T) {
	setupSuite(t)
	tests := []struct {
		name             string
		voterId          int
		validVoter       bool
		emptyVoter       bool
		expectedHttpCode int
	}{
		{"TestValidVoter", 1, true, false, 200},
		{"TestInvalidVoterId", -1, true, false, 400},
		{"TestExistingVoter", 1, true, false, 409},
		{"TestEmptyVoter", 2, false, true, 400},
		{"TestBadAddRequest", 2, false, false, 400},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var voterBody interface{}
			if tc.validVoter {
				voterBody = createRandomVoter(uint(tc.voterId))
			} else if tc.emptyVoter {
				voterBody = voter.Voter{}
			} else {
				voterBody = "Some string"
			}

			url := fmt.Sprintf("%s/voters/%d", base_url, tc.voterId)
			resp, err := cli.R().SetBody(voterBody).Post(url)
			if err != nil {
				t.Error("error adding voter: ", err)
			}

			if resp.StatusCode() != tc.expectedHttpCode {
				t.Errorf("expected %d status, got %d", tc.expectedHttpCode, resp.StatusCode())
			}
		})
	}
}

func TestVoterApi_GetVotersHandler(t *testing.T) {
	setupSuite(t)
	addNumberOfRandomVoters(10)

	resp, err := cli.R().SetResult(&[]voter.Voter{}).Get(base_url + "/voters")
	if err != nil {
		t.Error(err)
	}

	if resp.StatusCode() != 200 {
		t.Error("expected 200 status")
	}

	if (len(*resp.Result().(*[]voter.Voter))) != 10 {
		t.Error("expected 10 voters, found ", len(*resp.Result().(*[]voter.Voter)))
	}
}

func TestVoterApi_DeleteVotersHandler(t *testing.T) {
	setupSuite(t)
	addNumberOfRandomVoters(10)

	resp, err := cli.R().Delete(base_url + "/voters")
	if err != nil {
		t.Error(err)
	}

	if resp.StatusCode() != 200 {
		t.Error("expected 200 status")
	}

	resp, err = cli.R().SetResult(&[]voter.Voter{}).Get(base_url + "/voters")
	if err != nil {
		t.Error(err)
	}

	if (len(*resp.Result().(*[]voter.Voter))) != 0 {
		t.Error("expected 0 voters")
	}
}

func TestVoterApi_GetVoterHandler(t *testing.T) {
	setupSuite(t)
	tests := []struct {
		name             string
		input            int
		addVoter         bool
		expectedHttpCode int
	}{
		{"TestValidVoter", 1, true, 200},
		{"TestInvalidVoterId", -1, false, 400},
		{"TestNotFoundVoter", 2, false, 404},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var v voter.Voter
			if tc.addVoter {
				v = createRandomVoter(uint(tc.input))
				addVoter(v)
			}

			url := fmt.Sprintf("%s/voters/%d", base_url, tc.input)
			resp, err := cli.R().SetResult(&voter.Voter{}).Get(url)
			if err != nil {
				t.Error(err)
			}

			if resp.StatusCode() != tc.expectedHttpCode {
				t.Errorf("expected %d status, got %d", tc.expectedHttpCode, resp.StatusCode())
			}

			if tc.addVoter && int(resp.Result().(*voter.Voter).VoterId) != tc.input {
				t.Error("expected voter id ", tc.input)
			}

			if tc.addVoter && !cmp.Equal(*resp.Result().(*voter.Voter), v) {
				t.Error("voter does not match the added voter")
			}
		})
	}
}

func TestVoterApi_UpdateVoterHandler(t *testing.T) {
	setupSuite(t)
	tests := []struct {
		name             string
		voterId          int
		voterName        string
		voterEmail       string
		addRandomVoter   bool
		hasNameChanged   bool
		hasEmailChanged  bool
		isInvalidRequest bool
		expectedHttpCode int
	}{
		{"TestValidUpdate", 1, "John Doe", "john.doe@gmail.com", true, true, true, false, 200},
		{"TestInvalidVoterId", -1, "", "", false, false, false, false, 400},
		{"TestNonExistingVoter", 2, "John Doe", "john.doe@gmail.com", false, false, false, false, 404},
		{"TestEmptyVoter", 3, "", "", true, false, false, false, 200},
		{"TestBadUpdateRequest", 4, "", "", true, false, false, true, 400},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var addedVoter voter.Voter
			if tc.addRandomVoter {
				addedVoter = createRandomVoter(uint(tc.voterId))
				addVoter(addedVoter)

			}

			var voterUpdateBody interface{}
			if tc.isInvalidRequest {
				voterUpdateBody = "Some string"
			} else {
				voterUpdateBody = voter.Voter{
					Name:  tc.voterName,
					Email: tc.voterEmail,
				}
			}

			url := fmt.Sprintf("%s/voters/%d", base_url, tc.voterId)
			resp, err := cli.R().SetBody(voterUpdateBody).Put(url)
			if err != nil {
				t.Error("error updating voter: ", err)
			}

			if resp.StatusCode() != tc.expectedHttpCode {
				t.Errorf("expected %d status, got %d", tc.expectedHttpCode, resp.StatusCode())
			}

			resp, err = cli.R().SetResult(&voter.Voter{}).Get(url)
			if err != nil {
				t.Error("error getting updated voter: ", err)
			}
			updatedVoter := resp.Result().(*voter.Voter)

			if tc.hasNameChanged && updatedVoter.Name != tc.voterName {
				t.Errorf("expected name %s, got %s", tc.voterName, updatedVoter.Name)
			} else if !tc.hasNameChanged && updatedVoter.Name != addedVoter.Name {
				t.Errorf("expected name %s, got %s", addedVoter.Name, updatedVoter.Name)
			}

			if tc.hasEmailChanged && updatedVoter.Email != tc.voterEmail {
				t.Errorf("expected email %s, got %s", tc.voterEmail, updatedVoter.Email)
			} else if !tc.hasEmailChanged && updatedVoter.Email != addedVoter.Email {
				t.Errorf("expected email %s, got %s", addedVoter.Email, updatedVoter.Email)
			}
		})
	}
}

func TestVoterApi_DeleteVoterHandler(t *testing.T) {
	setupSuite(t)
	tests := []struct {
		name             string
		voterId          int
		addVoter         bool
		expectedHttpCode int
	}{
		{"TestValidDelete", 1, true, 200},
		{"TestInvalidVoterId", -1, false, 400},
		{"TestDeleteNonExistentVoter", 2, false, 404},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var voterBody interface{}
			if tc.addVoter {
				voterBody = createRandomVoter(uint(tc.voterId))
				addVoter(voterBody.(voter.Voter))
			} else {
				voterBody = "Some string"
			}

			url := fmt.Sprintf("%s/voters/%d", base_url, tc.voterId)
			resp, err := cli.R().SetBody(voterBody).Delete(url)
			if err != nil {
				t.Error("error deleting voter: ", err)
			}

			if resp.StatusCode() != tc.expectedHttpCode {
				t.Errorf("expected %d status, got %d", tc.expectedHttpCode, resp.StatusCode())
			}
		})
	}
}

func TestVoterApi_AddVoterHistoryHandler(t *testing.T) {
	setupSuite(t)
	tests := []struct {
		name             string
		voterId          int
		pollId           int
		voteId           int
		addVoter         bool
		isInvalidRequest bool
		expectedHttpCode int
	}{
		{"TestValidAdd", 1, 1, 1, true, false, 200},
		{"TestVoteAgain", 1, 1, 1, false, false, 404},
		{"TestInvalidVoterId", -1, 1, 1, false, false, 400},
		{"TestInvalidPollId", 1, -1, 1, false, false, 400},
		{"TestInvalidRequest", 2, 2, -2, false, true, 400},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.addVoter {
				addVoter(createRandomVoter(uint(tc.voterId)))
			}

			var voterHistoryBody interface{}
			if tc.isInvalidRequest {
				voterHistoryBody = "Some string"
			} else {
				voterHistoryBody = createVoterHistory(uint(tc.voteId))
			}

			url := fmt.Sprintf("%s/voters/%d/polls/%d", base_url, tc.voterId, tc.pollId)
			resp, err := cli.R().SetBody(voterHistoryBody).Post(url)
			if err != nil {
				t.Error("error adding voterHistory: ", err)
			}

			if resp.StatusCode() != tc.expectedHttpCode {
				t.Errorf("expected %d status, got %d", tc.expectedHttpCode, resp.StatusCode())
			}
		})
	}
}

func TestVoterApi_GetVoterHistoryHandler(t *testing.T) {
	setupSuite(t)
	tests := []struct {
		name             string
		voterId          int
		pollId           int
		voteId           int
		addVoter         bool
		addVoterHistory  bool
		expectedHttpCode int
	}{
		{"TestValidAdd", 1, 1, 1, true, true, 200},
		{"TestNoVoter", 2, 2, 2, false, false, 404},
		{"TestNoVoterHistory", 2, 1, 1, true, false, 404},
		{"TestNoPoll", 1, 2, 1, false, false, 404},
		{"TestInvalidVoterId", -1, 1, 1, false, false, 400},
		{"TestInvalidPollId", 1, -1, 1, false, false, 400},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.addVoter {
				addVoter(createRandomVoter(uint(tc.voterId)))
			}

			if tc.addVoterHistory {
				addVoterHistory(uint(tc.voterId), uint(tc.pollId), uint(tc.voteId))
			}

			url := fmt.Sprintf("%s/voters/%d/polls/%d", base_url, tc.voterId, tc.pollId)
			resp, err := cli.R().SetResult(&voter.VoterHistory{}).Get(url)
			if err != nil {
				t.Error("error getting voterHistory: ", err)
			}

			if resp.StatusCode() != tc.expectedHttpCode {
				t.Errorf("expected %d status, got %d", tc.expectedHttpCode, resp.StatusCode())
			}

			if resp.StatusCode() == 200 {
				voterHistoryRes := resp.Result().(*voter.VoterHistory)
				if voterHistoryRes == nil {
					t.Error("expected voter history but got nil")
				}
				if int((*voterHistoryRes).VoteId) != tc.voteId {
					t.Errorf("expected vote id %d, got %d", tc.voteId, (*voterHistoryRes).VoteId)
				}
			}
		})
	}
}

func TestVoterApi_UpdateVoterHistoryHandler(t *testing.T) {
	setupSuite(t)
	tests := []struct {
		name                  string
		voterId               int
		pollId                int
		voteId                int
		timeString            string
		addRandomVoter        bool
		addRandomVoterHistory bool
		hasVoteIdChanged      bool
		hasTimeChanged        bool
		isInvalidRequest      bool
		expectedHttpCode      int
	}{
		{"TestValidUpdate", 1, 1, 1, "25 Dec 07 18:37 -0600", true, true, true, true, false, 200},
		{"TestInvalidVoterId", -1, 2, 2, "", false, false, false, false, false, 400},
		{"TestInvalidPollId", 2, -1, 2, "", false, false, false, false, false, 400},
		{"TestNonExistentVoter", 2, 2, 2, "", false, false, false, false, false, 404},
		{"TestNonExistentPoll", 1, 2, 2, "", false, false, false, false, false, 404},
		{"TestNonExistentHistory", 2, 2, 2, "", true, false, false, false, false, 404},
		{"TestBadUpdateRequest", 3, 3, 3, "", false, false, false, false, true, 400},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var addedVoter voter.Voter
			if tc.addRandomVoter {
				addedVoter = createRandomVoter(uint(tc.voterId))
				addVoter(addedVoter)
			}

			addedVoteId := fake.Uint8()
			if tc.addRandomVoterHistory {
				addVoterHistory(uint(tc.voterId), uint(tc.pollId), uint(addedVoteId))
			}

			var voterHistoryUpdateBody interface{}
			if tc.isInvalidRequest {
				voterHistoryUpdateBody = "Some string"
			} else {
				voterHistoryUpdateBody = struct {
					VoteId   uint   `json:"voteId,omitempty"`
					VoteDate string `json:"voteDate,omitempty"`
				}{
					VoteId:   uint(tc.voteId),
					VoteDate: tc.timeString,
				}
			}

			url := fmt.Sprintf("%s/voters/%d/polls/%d", base_url, tc.voterId, tc.pollId)
			resp, err := cli.R().SetBody(voterHistoryUpdateBody).Put(url)
			if err != nil {
				t.Error("error updating voter history: ", err)
			}

			if resp.StatusCode() != tc.expectedHttpCode {
				t.Errorf("expected %d status, got %d", tc.expectedHttpCode, resp.StatusCode())
			}

			if tc.addRandomVoterHistory {
				resp, err = cli.R().SetResult(&voter.VoterHistory{}).Get(url)
				if err != nil {
					t.Error("error getting updated voter history: ", err)
				}
				updatedVoterHistory := resp.Result().(*voter.VoterHistory)

				if tc.hasVoteIdChanged && updatedVoterHistory.VoteId != uint(tc.voteId) {
					t.Errorf("expected voteId %d, got %d", tc.voteId, updatedVoterHistory.VoteId)
				} else if !tc.hasVoteIdChanged && updatedVoterHistory.VoteId != uint(addedVoteId) {
					t.Errorf("expected voteId %d, got %d", addedVoteId, updatedVoterHistory.VoteId)
				}

				if tc.hasTimeChanged && updatedVoterHistory.VoteDate.Format(time.RFC822Z) != tc.timeString {
					t.Errorf("expected time %s, got %s", tc.timeString, updatedVoterHistory.VoteDate.Format(time.RFC822Z))
				}
			}
		})
	}
}

func TestVoterApi_DeleteVoterHistoryHandler(t *testing.T) {
	setupSuite(t)
	tests := []struct {
		name                  string
		voterId               int
		pollId                int
		addRandomVoter        bool
		addRandomVoterHistory bool
		expectedHttpCode      int
	}{
		{"TestValidDelete", 1, 1, true, true, 200},
		{"TestInvalidVoterId", -1, 2, false, false, 400},
		{"TestInvalidPollId", 2, -1, false, false, 400},
		{"TestNonExistentVoter", 2, 2, false, false, 404},
		{"TestNonExistentPoll", 3, 3, true, false, 404},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.addRandomVoter {
				addVoter(createRandomVoter(uint(tc.voterId)))
			}

			if tc.addRandomVoterHistory {
				addVoterHistory(uint(tc.voterId), uint(tc.pollId), uint(fake.Uint8()))
			}

			url := fmt.Sprintf("%s/voters/%d/polls/%d", base_url, tc.voterId, tc.pollId)
			resp, err := cli.R().Delete(url)
			if err != nil {
				t.Error("error deleting voterHistory: ", err)
			}

			if resp.StatusCode() != tc.expectedHttpCode {
				t.Errorf("expected %d status, got %d", tc.expectedHttpCode, resp.StatusCode())
			}
		})
	}
}

func TestVoterApi_GetVoterHistoriesHandler(t *testing.T) {
	setupSuite(t)
	tests := []struct {
		name              string
		voterId           int
		addRandomVoter    bool
		numHistoriesToAdd int
		expectedHttpCode  int
	}{
		{"TestValidGet", 1, true, 5, 200},
		{"TestInvalidVoterId", -1, false, 0, 400},
		{"TestNonExistentVoter", 2, false, 0, 404},
		{"TestNoHistories", 3, true, 0, 404},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.addRandomVoter {
				addVoter(createRandomVoter(uint(tc.voterId)))
			}

			if tc.numHistoriesToAdd > 0 {
				for i := 0; i < tc.numHistoriesToAdd; i++ {
					addVoterHistory(uint(tc.voterId), uint(fake.Uint8()), uint(fake.Uint8()))
				}
			}

			url := fmt.Sprintf("%s/voters/%d/polls", base_url, tc.voterId)
			resp, err := cli.R().SetResult(&[]voter.VoterHistory{}).Get(url)
			if err != nil {
				t.Error("error getting voterHistory: ", err)
			}

			if resp.StatusCode() != tc.expectedHttpCode {
				t.Errorf("expected %d status, got %d", tc.expectedHttpCode, resp.StatusCode())
			}

			if resp.StatusCode() == 200 {
				voterHistoriesRes := resp.Result().(*[]voter.VoterHistory)
				if voterHistoriesRes == nil {
					t.Error("expected voter histories but got nil")
				} else {
					if len(*voterHistoriesRes) != tc.numHistoriesToAdd {
						t.Errorf("expected %d voter histories, got %d", tc.numHistoriesToAdd, len(*voterHistoriesRes))
					}
				}
			}
		})
	}
}

func createRandomVoter(voterId uint) voter.Voter {
	return *voter.NewVoter(voterId, fake.Name(), fake.Email())
}

func addVoter(voter voter.Voter) {
	url := fmt.Sprintf("%s/voters/%d", base_url, voter.VoterId)
	_, err := cli.R().SetBody(voter).Post(url)
	if err != nil {
		log.Fatalf("Error adding voter: %v", err)
	}
}

func addNumberOfRandomVoters(numberOfVoters int) []voter.Voter {
	voters := make([]voter.Voter, 0, numberOfVoters)
	for i := 0; i < numberOfVoters; i++ {
		voter := createRandomVoter(uint(i))
		addVoter(voter)
		voters = append(voters, voter)
	}
	return voters
}

func createVoterHistory(voteId uint) voter.VoterHistory {
	return voter.VoterHistory{
		VoteId: voteId,
	}
}

func addVoterHistory(voterId uint, pollId uint, voteId uint) {
	url := fmt.Sprintf("%s/voters/%d/polls/%d", base_url, voterId, pollId)
	_, err := cli.R().SetBody(createVoterHistory(voteId)).Post(url)
	if err != nil {
		log.Fatalf("Error adding voter history: %v", err)
	}
}
