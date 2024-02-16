package voter

import (
	"encoding/json"
	"errors"
	"time"
)

type VoterHistory struct {
	PollId   uint      `json:"pollId"`
	VoteId   uint      `json:"voteId"`
	VoteDate time.Time `json:"voteDate"`
}

type Voter struct {
	VoterId     uint           `json:"voterId"`
	Name        string         `json:"name"`
	Email       string         `json:"email"`
	VoteHistory []VoterHistory `json:"voteHistory"`
}

type VoterMap struct {
	voters map[uint]*Voter
}

func NewVoterHistory(pollId uint, voteId uint, voteDate time.Time) *VoterHistory {
	return &VoterHistory{
		pollId,
		voteId,
		voteDate,
	}
}

func NewVoter(voterId uint, name string, email string) *Voter {
	return &Voter{
		voterId,
		name,
		email,
		make([]VoterHistory, 0),
	}
}

func NewVoterMap() *VoterMap {
	return &VoterMap{
		make(map[uint]*Voter),
	}
}

func (voterMap *VoterMap) GetVoters() *[]Voter {
	voters := make([]Voter, 0, len(voterMap.voters))
	for _, voter := range voterMap.voters {
		voters = append(voters, *voter)
	}

	return &voters
}

func (voterMap *VoterMap) AddVoter(voter Voter) (*Voter, error) {
	_, ok := voterMap.voters[voter.VoterId]
	if ok {
		return nil, errors.New("voter already exists")
	}

	voterMap.voters[voter.VoterId] = &voter
	return &voter, nil
}

func (voterMap *VoterMap) GetVoter(voterId uint) (*Voter, error) {
	voter, ok := voterMap.voters[voterId]
	if !ok {
		return nil, errors.New("voter not found")
	}

	return voter, nil
}

func (voterMap *VoterMap) UpdateVoter(voterId uint, name string, email string) (*Voter, error) {
	voter, err := voterMap.GetVoter(voterId)
	if err != nil {
		return nil, err
	}

	if len(name) > 0 {
		voter.Name = name
	}

	if len(email) > 0 {
		voter.Email = email
	}

	return voter, nil
}

func (voterMap *VoterMap) DeleteVoter(voterId uint) (*Voter, error) {
	voter, err := voterMap.GetVoter(voterId)
	if err != nil {
		return nil, err
	}

	delete(voterMap.voters, voterId)
	return voter, nil
}

func (voterMap *VoterMap) GetVoterHistories(voterId uint) (*[]VoterHistory, error) {
	voter, err := voterMap.GetVoter(voterId)
	if err != nil {
		return nil, err
	}

	if len(voter.VoteHistory) == 0 {
		return nil, errors.New("no vote history")
	}

	return &voter.VoteHistory, nil
}

func (voterMap *VoterMap) AddVoterHistory(voterId uint, pollId uint, voteId uint) (*VoterHistory, error) {
	voter, err := voterMap.GetVoter(voterId)
	if err != nil {
		return nil, err
	}

	voterHistory, _ := getVoterHistoryForPollId(voter, pollId)
	if voterHistory != nil {
		return nil, errors.New("Voter has already voted for this poll")
	}

	voterHistory = NewVoterHistory(pollId, voteId, time.Now())
	voter.VoteHistory = append(voter.VoteHistory, *voterHistory)

	return voterHistory, nil
}

func (voterMap *VoterMap) GetVoterHistory(voterId uint, pollId uint) (*VoterHistory, error) {
	voter, err := voterMap.GetVoter(voterId)
	if err != nil {
		return nil, err
	}

	return getVoterHistoryForPollId(voter, pollId)
}

func (voterMap *VoterMap) UpdateVoterHistory(voterId uint, pollId uint, voteIdPointer *uint, time *time.Time) (*VoterHistory, error) {
	voter, err := voterMap.GetVoter(voterId)
	if err != nil {
		return nil, err
	}

	voterHistory, err := getVoterHistoryForPollId(voter, pollId)
	if err != nil {
		return nil, err
	}

	if voteIdPointer != nil {
		voterHistory.VoteId = *voteIdPointer
	}

	if time != nil {
		voterHistory.VoteDate = *time
	}

	return voterHistory, nil
}

func (voterMap *VoterMap) DeleteVoterHistory(voterId uint, pollId uint) (*VoterHistory, error) {
	voter, err := voterMap.GetVoter(voterId)
	if err != nil {
		return nil, err
	}

	voterHistory, err := getVoterHistoryForPollId(voter, pollId)
	if err != nil {
		return nil, err
	}

	voter.VoteHistory = getVoterHistorySliceWithoutPollId(voter, pollId)

	return voterHistory, nil
}

// Override the default JSON marshalling to format the date as RFC822Z
// Modified from: https://stackoverflow.com/a/35744769
func (voterHistory *VoterHistory) MarshalJSON() ([]byte, error) {
	type Alias VoterHistory
	return json.Marshal(&struct {
		*Alias
		VoteDate string `json:"voteDate"`
	}{
		Alias:    (*Alias)(voterHistory),
		VoteDate: voterHistory.VoteDate.Format(time.RFC822Z),
	})
}

func getVoterHistoryForPollId(voter *Voter, pollId uint) (*VoterHistory, error) {
	for index, _ := range voter.VoteHistory { // range produces a copy of the value at a specific index, so we should use the index, not the value, so we can return the correct pointer.
		if voter.VoteHistory[index].PollId == pollId {
			return &voter.VoteHistory[index], nil
		}
	}

	return nil, errors.New("vote history not found for specified poll")
}

func getVoterHistorySliceWithoutPollId(voter *Voter, pollId uint) []VoterHistory {
	voteHistory := make([]VoterHistory, 0, len(voter.VoteHistory))
	for _, vh := range voter.VoteHistory { // unlike getVoterHistoryForPollId, we can use the value here, since we're not working with pointers
		if vh.PollId != pollId {
			voteHistory = append(voteHistory, vh)
		}
	}

	return voteHistory
}
