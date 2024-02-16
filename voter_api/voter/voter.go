package voter

import (
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

	voterHistory, _ := getVoteHistoryForPollId(voter, pollId)
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

	return getVoteHistoryForPollId(voter, pollId)
}

func getVoteHistoryForPollId(voter *Voter, pollId uint) (*VoterHistory, error) {
	for _, vh := range voter.VoteHistory {
		if vh.PollId == pollId {
			return &vh, nil
		}
	}

	return nil, errors.New("vote history not found for specified poll")
}
