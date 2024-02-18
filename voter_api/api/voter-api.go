package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ea673/voter-api/voter"
	"github.com/gofiber/fiber/v2"
)

type VoterApi struct {
	voterMap     voter.VoterMap
	apiBootTime  time.Time
	numApiCalls  int
	numApiErrors int
}

type healthCheckResponse struct {
	Status       string `json:"status"`
	Uptime       string `json:"uptime"`
	NumApiCalls  int    `json:"numApiCalls"`
	NumApiErrors int    `json:"numApiErrors"`
}

type addOrUpdateVoterJson struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type addVoterHistoryJson struct {
	VoteId uint `json:"voteId"`
}

type updateVoterHistoryJson struct {
	VoteId   *uint   `json:"voteId"`
	VoteDate *string `json:"voteDate"`
}

func NewVoterApi() *VoterApi {
	return &VoterApi{
		*voter.NewVoterMap(),
		time.Now(),
		0,
		0,
	}
}

func (voterApi *VoterApi) GetVotersHandler(c *fiber.Ctx) error {
	voterApi.numApiCalls++
	voters := voterApi.voterMap.GetVoters()
	return c.JSON(voters)
}

func (voterApi *VoterApi) DeleteVotersHandler(c *fiber.Ctx) error {
	voterApi.numApiCalls++
	err := voterApi.voterMap.DeleteVoters()
	if err != nil {
		voterApi.numApiErrors++
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}
	return c.SendString("All voters deleted")
}

func (voterApi *VoterApi) AddVoterHandler(c *fiber.Ctx) error {
	voterApi.numApiCalls++
	voterId, err := c.ParamsInt("id", -1)

	if err != nil || voterId < 0 {
		voterApi.numApiErrors++
		return fiber.NewError(http.StatusBadRequest, "Invalid voter id")
	}

	var voterJson addOrUpdateVoterJson
	if err := c.BodyParser(&voterJson); err != nil {
		voterApi.numApiErrors++
		log.Println("Error binding JSON: ", err)
		return fiber.NewError(http.StatusBadRequest, "Invalid request")
	}

	if voterJson.Name == "" || voterJson.Email == "" {
		return fiber.NewError(http.StatusBadRequest, "Name and email are required")
	}

	newVoter, err := voterApi.voterMap.AddVoter(*voter.NewVoter(uint(voterId), voterJson.Name, voterJson.Email))
	if err != nil {
		return fiber.NewError(http.StatusConflict, err.Error())
	}
	res := fmt.Sprintf("Voter added: %v", convertToJson(newVoter))
	return c.SendString(res)
}

func (voterApi *VoterApi) UpdateVoterHandler(c *fiber.Ctx) error {
	voterApi.numApiCalls++
	voterId, err := c.ParamsInt("id", -1)

	if err != nil || voterId < 0 {
		voterApi.numApiErrors++
		return fiber.NewError(http.StatusBadRequest, "Invalid voter id")
	}

	var voterJson addOrUpdateVoterJson
	if err := c.BodyParser(&voterJson); err != nil {
		voterApi.numApiErrors++
		log.Println("Error binding JSON: ", err)
		return fiber.NewError(http.StatusBadRequest, "Invalid request")
	}

	updatedVoter, err := voterApi.voterMap.UpdateVoter(uint(voterId), voterJson.Name, voterJson.Email)
	if err != nil {
		voterApi.numApiErrors++
		return fiber.NewError(http.StatusNotFound, err.Error())
	}

	res := fmt.Sprintf("Voter updated: %v", convertToJson(updatedVoter))
	return c.SendString(res)
}

func (voterApi *VoterApi) DeleteVoterHandler(c *fiber.Ctx) error {
	voterApi.numApiCalls++
	voterId, err := c.ParamsInt("id", -1)
	if err != nil || voterId < 0 {
		voterApi.numApiErrors++
		return fiber.NewError(http.StatusBadRequest, "Invalid voter id")
	}

	voter, err := voterApi.voterMap.DeleteVoter(uint(voterId))
	if err != nil {
		voterApi.numApiErrors++
		return fiber.NewError(http.StatusNotFound, err.Error())
	}

	res := fmt.Sprintf("Voter deleted: %v", voter)
	return c.SendString(res)
}

func (voterApi *VoterApi) GetVoterHandler(c *fiber.Ctx) error {
	voterApi.numApiCalls++
	voterId, err := c.ParamsInt("id", -1)
	if err != nil || voterId < 0 {
		voterApi.numApiErrors++
		return fiber.NewError(http.StatusBadRequest, "Invalid voter id")
	}

	voter, err := voterApi.voterMap.GetVoter(uint(voterId))
	if err != nil {
		voterApi.numApiErrors++
		return fiber.NewError(http.StatusNotFound, err.Error())
	}

	return c.JSON(voter)
}

func (voterApi *VoterApi) GetVoterHistoriesHandler(c *fiber.Ctx) error {
	voterApi.numApiCalls++
	voterId, err := c.ParamsInt("id", -1)
	if err != nil || voterId < 0 {
		voterApi.numApiErrors++
		return fiber.NewError(http.StatusBadRequest, "Invalid voter id")
	}

	voterHistory, err := voterApi.voterMap.GetVoterHistories(uint(voterId))
	if err != nil {
		voterApi.numApiErrors++
		return fiber.NewError(http.StatusNotFound, err.Error())
	}

	return c.JSON(voterHistory)
}

func (voterApi *VoterApi) AddVoterHistoryHandler(c *fiber.Ctx) error {
	voterApi.numApiCalls++
	voterId, err := c.ParamsInt("id", -1)
	if err != nil || voterId < 0 {
		voterApi.numApiErrors++
		return fiber.NewError(http.StatusBadRequest, "Invalid voter id")
	}

	pollId, err := c.ParamsInt("pollid", -1)
	if err != nil || pollId < 0 {
		voterApi.numApiErrors++
		return fiber.NewError(http.StatusBadRequest, "Invalid poll id")
	}

	var voterHistoryJson addVoterHistoryJson
	if err := c.BodyParser(&voterHistoryJson); err != nil {
		voterApi.numApiErrors++
		log.Println("Error binding JSON: ", err)
		return fiber.NewError(http.StatusBadRequest, "Invalid request")
	}

	newVoterHistory, err := voterApi.voterMap.AddVoterHistory(uint(voterId), uint(pollId), voterHistoryJson.VoteId)
	if err != nil {
		voterApi.numApiErrors++
		return fiber.NewError(http.StatusNotFound, err.Error())
	}

	res := fmt.Sprintf("Voter history added: %v", convertToJson(newVoterHistory))
	return c.SendString(res)
}

func (voterApi *VoterApi) UpdateVoterHistoryHandler(c *fiber.Ctx) error {
	voterApi.numApiCalls++
	voterId, err := c.ParamsInt("id", -1)
	if err != nil || voterId < 0 {
		voterApi.numApiErrors++
		return fiber.NewError(http.StatusBadRequest, "Invalid voter id")
	}

	pollId, err := c.ParamsInt("pollid", -1)
	if err != nil || pollId < 0 {
		voterApi.numApiErrors++
		return fiber.NewError(http.StatusBadRequest, "Invalid poll id")
	}

	var voterHistoryJson updateVoterHistoryJson
	if err := c.BodyParser(&voterHistoryJson); err != nil {
		voterApi.numApiErrors++
		log.Println("Error binding JSON: ", err)
		return fiber.NewError(http.StatusBadRequest, "Invalid request")
	}

	var voteDatePointer *time.Time
	if voterHistoryJson.VoteDate != nil {
		voteDate, err := time.Parse(time.RFC822Z, *voterHistoryJson.VoteDate)
		if err != nil {
			voterApi.numApiErrors++
			return fiber.NewError(http.StatusBadRequest, "Invalid date. Expected format: RFC822Z")
		}
		voteDatePointer = &voteDate
	}

	updatedVoterHistory, err := voterApi.voterMap.UpdateVoterHistory(uint(voterId), uint(pollId), voterHistoryJson.VoteId, voteDatePointer)
	if err != nil {
		voterApi.numApiErrors++
		return fiber.NewError(http.StatusNotFound, err.Error())
	}

	res := fmt.Sprintf("Voter history updated: %v", convertToJson(updatedVoterHistory))
	return c.SendString(res)
}

func (voterApi *VoterApi) DeleteVoterHistoryHandler(c *fiber.Ctx) error {
	voterApi.numApiCalls++
	voterId, err := c.ParamsInt("id", -1)
	if err != nil || voterId < 0 {
		voterApi.numApiErrors++
		return fiber.NewError(http.StatusBadRequest, "Invalid voter id")
	}

	pollId, err := c.ParamsInt("pollid", -1)
	if err != nil || pollId < 0 {
		voterApi.numApiErrors++
		return fiber.NewError(http.StatusBadRequest, "Invalid poll id")
	}

	voterHistory, err := voterApi.voterMap.DeleteVoterHistory(uint(voterId), uint(pollId))
	if err != nil {
		voterApi.numApiErrors++
		return fiber.NewError(http.StatusNotFound, err.Error())
	}

	res := fmt.Sprintf("Voter history deleted: %v", convertToJson(voterHistory))
	return c.SendString(res)
}

func (voterApi *VoterApi) GetVoterHistoryHandler(c *fiber.Ctx) error {
	voterApi.numApiCalls++
	voterId, err := c.ParamsInt("id", -1)
	if err != nil || voterId < 0 {
		voterApi.numApiErrors++
		return fiber.NewError(http.StatusBadRequest, "Invalid voter id")
	}

	pollId, err := c.ParamsInt("pollid", -1)
	if err != nil || pollId < 0 {
		voterApi.numApiErrors++
		return fiber.NewError(http.StatusBadRequest, "Invalid poll id")
	}

	voterHistory, err := voterApi.voterMap.GetVoterHistory(uint(voterId), uint(pollId))
	if err != nil {
		voterApi.numApiErrors++
		return fiber.NewError(http.StatusNotFound, err.Error())
	}

	return c.JSON(voterHistory)
}

func (voteApi *VoterApi) GetHealthHandler(c *fiber.Ctx) error {
	healthStatus := healthCheckResponse{
		Status:       "OK",
		Uptime:       time.Since(voteApi.apiBootTime).Round(time.Second).String(),
		NumApiCalls:  voteApi.numApiCalls,
		NumApiErrors: voteApi.numApiErrors,
	}
	return c.JSON(healthStatus)
}

func convertToJson(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
