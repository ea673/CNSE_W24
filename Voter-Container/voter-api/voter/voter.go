package voter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/nitishm/go-rejson/v4"
	"github.com/redis/go-redis/v9"
)

type CustomTime struct {
	time.Time
}

type VoterHistory struct {
	PollId   uint       `json:"pollId"`
	VoteId   uint       `json:"voteId"`
	VoteDate CustomTime `json:"voteDate"`
}

type Voter struct {
	VoterId     uint           `json:"voterId"`
	Name        string         `json:"name"`
	Email       string         `json:"email"`
	VoteHistory []VoterHistory `json:"voteHistory"`
}

type VoterMap struct {
	cache cache
	// voters map[uint]*Voter
}

const (
	RedisNilError        = "redis: nil"
	RedisDefaultLocation = "0.0.0.0:6379"
	RedisKeyPrefix       = "voter:"
)

type cache struct {
	cacheClient *redis.Client
	jsonHelper  *rejson.Handler
	context     context.Context
}

func NewVoterHistory(pollId uint, voteId uint, voteDate time.Time) *VoterHistory {
	return &VoterHistory{
		pollId,
		voteId,
		CustomTime{voteDate},
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

func NewVoterMap() (*VoterMap, error) {
	// Get the REDIS_URL environment variable
	redisUrl := os.Getenv("REDIS_URL")

	// If the REDIS_URL environment variable is not set, use the default location 0.0.0.0:6379
	if redisUrl == "" {
		redisUrl = RedisDefaultLocation
	}

	client := redis.NewClient(&redis.Options{
		Addr: redisUrl,
	})

	ctx := context.Background()

	err := client.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}

	jsonHelper := rejson.NewReJSONHandler()
	jsonHelper.SetGoRedisClientWithContext(ctx, client)

	return &VoterMap{
		cache{
			client,
			jsonHelper,
			ctx,
		},
	}, nil
}

func (voterMap *VoterMap) GetVoters() *[]Voter {
	voters := make([]Voter, 0)

	redisKey := RedisKeyPrefix + "*"
	keys, _ := voterMap.cache.cacheClient.Keys(voterMap.cache.context, redisKey).Result()
	for _, key := range keys {
		voter, err := voterMap.getItemFromRedis(key)
		if err != nil {
			return nil
		}
		voters = append(voters, *voter)
	}

	return &voters
}
func (voterMap *VoterMap) DeleteVoters() error {
	reisKey := RedisKeyPrefix + "*"
	keys, _ := voterMap.cache.cacheClient.Keys(voterMap.cache.context, reisKey).Result()
	numDeleted, err := voterMap.cache.cacheClient.Del(voterMap.cache.context, keys...).Result()
	if err != nil {
		return err
	}

	if numDeleted != int64(len(keys)) {
		return errors.New("failed to delete all voters")
	}
	return nil
}

func (voterMap *VoterMap) AddVoter(voter Voter) (*Voter, error) {
	// Check if the voter already exists
	redisKey := getRedisKeyFromId(voter.VoterId)
	numVotersFound, err := voterMap.cache.cacheClient.Exists(voterMap.cache.context, redisKey).Result()
	if err != nil {
		return nil, err
	}

	// If the voter already exists, return an error
	if numVotersFound > 0 {
		return nil, errors.New("voter already exists")
	}

	// Add the voter to Redis
	_, err = voterMap.cache.jsonHelper.JSONSet(redisKey, ".", voter)
	if err != nil {
		return nil, err
	}

	return &voter, nil
}

func (voterMap *VoterMap) GetVoter(voterId uint) (*Voter, error) {
	redisKey := getRedisKeyFromId(voterId)
	log.Println("redisKey: ", redisKey)
	voter, err := voterMap.getItemFromRedis(redisKey)
	if err != nil {
		return nil, err
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

	_, err = voterMap.cache.jsonHelper.JSONSet(getRedisKeyFromId(voterId), ".", voter)
	if err != nil {
		return nil, err
	}

	return voter, nil
}

func (voterMap *VoterMap) DeleteVoter(voterId uint) (*Voter, error) {
	voter, err := voterMap.GetVoter(voterId)
	if err != nil {
		return nil, err
	}

	_, err = voterMap.cache.cacheClient.Del(voterMap.cache.context, getRedisKeyFromId(voterId)).Result()
	if err != nil {
		return nil, err
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

	voterHistory, _ := getVoterHistoryForPollId(voter, pollId)
	if voterHistory != nil {
		return nil, errors.New("Voter has already voted for this poll")
	}

	voterHistory = NewVoterHistory(pollId, voteId, time.Now())
	voter.VoteHistory = append(voter.VoteHistory, *voterHistory)

	// Save the updated voter with the new voter history to Redis
	_, err = voterMap.cache.jsonHelper.JSONSet(getRedisKeyFromId(voterId), ".", voter)
	if err != nil {
		return nil, err
	}

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
		voterHistory.VoteDate = CustomTime{*time}
	}

	// Save the updated voter with the updated voter history to Redis
	_, err = voterMap.cache.jsonHelper.JSONSet(getRedisKeyFromId(voterId), ".", voter)
	if err != nil {
		return nil, err
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

	// Save the updated voter with the deleted voter history to Redis
	_, err = voterMap.cache.jsonHelper.JSONSet(getRedisKeyFromId(voterId), ".", voter)
	if err != nil {
		return nil, err
	}

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

// Override the default JSON unmarshalling to parse the date as RFC822Z
func (customTime *CustomTime) UnmarshalJSON(data []byte) error {
	timeString := strings.Trim(string(data), "\"")
	if timeString == "null" {
		customTime.Time = time.Time{}
		return nil
	}
	var err error
	customTime.Time, err = time.Parse(time.RFC822Z, timeString)
	return err
}

func getVoterHistoryForPollId(voter *Voter, pollId uint) (*VoterHistory, error) {
	for index := range voter.VoteHistory { // range produces a copy of the value at a specific index, so we should use the index, not the value, so we can return the correct pointer.
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

// Redis Helper Methods

func getRedisKeyFromId(id uint) string {
	return fmt.Sprintf("%s%d", RedisKeyPrefix, id)
}

func isRedisNilError(err error) bool {
	return err != nil && err.Error() == RedisNilError
}

func (voterMap *VoterMap) getItemFromRedis(key string) (*Voter, error) {
	//Get the item from Redis
	itemObject, err := voterMap.cache.jsonHelper.JSONGet(key, ".")
	if err != nil {
		if isRedisNilError(err) {
			return nil, errors.New("voter not found")
		} else {
			return nil, err
		}
	}

	// Unmarshal the item into a Voter
	var voter Voter
	err = json.Unmarshal(itemObject.([]byte), &voter)
	if err != nil {
		return nil, err
	}

	return &voter, nil
}
