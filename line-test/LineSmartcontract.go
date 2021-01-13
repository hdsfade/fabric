//@author: hdsfade
//@date: 2021-01-06-14:54
package line

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

//Line describes details of a line
type Line struct {
	LineNumber     int      `json:"lineNumber"`
	WayStation     []string `json:"wayStation"`
	WayStationType []string `json:"wayStationType"`
	Using          bool     `json:"using"`
}

type Lines []Line

//Result structure used for handing result of create or delete
type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

//QueryResult structure used for handing result of query
type QueryResult struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data Line   `json:"data"`
}

//QueryResult structure used for handing result of queryAll
type QueryResults struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data Lines  `json:"data"`
}

// Init lines' ledger(can add a default set of lines to the ledger)
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	lines := []Line{
		{LineNumber: 1, WayStation: []string{"宁波", "杭州", "南京"}, WayStationType: []string{"始发站", "途径站", "终点站"}, Using: true},
		{LineNumber: 2, WayStation: []string{"宁波", "杭州", "上海"}, WayStationType: []string{"始发站", "途径站", "终点站"}, Using: true},
		{LineNumber: 3, WayStation: []string{"宁波", "嘉兴", "上海"}, WayStationType: []string{"始发站", "途径站", "终点站"}, Using: true},
	}
	for _, line := range lines {
		lineJSON, err := json.Marshal(line)
		if err != nil {
			return nil
		}

		err = ctx.GetStub().PutState(string(line.LineNumber), lineJSON)
		if err != nil {
			return nil
		}
	}
	return nil
}

//LineExists judges a  if exists or not.
func (s *SmartContract) LineExists(ctx contractapi.TransactionContextInterface, lineNumber int) (bool, error) {
	lineJSON, err := ctx.GetStub().GetState(string(lineNumber))
	if err != nil {
		return false, fmt.Errorf("failed to read from world state %v", err)
	}
	return lineJSON != nil, nil
}

//CreateLine issues a new line to the world state with given details.
func (s *SmartContract) CreateLine(ctx contractapi.TransactionContextInterface, lineNumber int, wayStation, wayStationType []string) Result {
	exists, err := s.LineExists(ctx, lineNumber)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	if exists {
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("the line %d already exists", lineNumber),
		}
	}

	line := Line{
		LineNumber:     lineNumber,
		WayStation:     wayStation,
		WayStationType: wayStationType,
		Using:          true,
	}
	lineJSON, err := json.Marshal(line)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}

	err = ctx.GetStub().PutState(string(lineNumber), lineJSON)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	return Result{
		Code: 200,
		Msg:  "",
	}
}

//DeleteLine deletes a line by lineNumber from the world state.
func (s *SmartContract) DeleteLine(ctx contractapi.TransactionContextInterface, lineNumber int) Result {
	exists, err := s.LineExists(ctx, lineNumber)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	if !exists {
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("the line %d does not exist", lineNumber),
		}
	}
	err = ctx.GetStub().DelState(string(lineNumber))
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	return Result{
		Code: 200,
		Msg:  "",
	}
}

// QueryLineBylinenumber returns the line stored in the world state with given lineNumver
func (s *SmartContract) QueryLineBylinenumber(ctx contractapi.TransactionContextInterface, lineNumber int) QueryResult {
	lineJSON, err := ctx.GetStub().GetState(string(lineNumber))
	if err != nil {
		return QueryResult{
			Code: 402,
			Msg:  fmt.Sprintf("failed to read from world state: %v", err),
			Data: Line{
				LineNumber:     0,
				WayStation:     []string{},
				WayStationType: []string{},
				Using:          false,
			},
		}
	}
	if lineJSON == nil {
		return QueryResult{
			Code: 402,
			Msg:  fmt.Sprintf("the lien %d does not exist", lineNumber),
			Data: Line{
				LineNumber:     0,
				WayStation:     []string{},
				WayStationType: []string{},
				Using:          false,
			},
		}
	}

	var line Line
	err = json.Unmarshal(lineJSON, &line)
	if err != nil {
		return QueryResult{
			Code: 402,
			Msg:  err.Error(),
			Data: Line{
				LineNumber:     0,
				WayStation:     []string{},
				WayStationType: []string{},
				Using:          false,
			},
		}
	}
	return QueryResult{
		Code: 200,
		Msg:  "",
		Data: line,
	}
}

// QueryAllLines returns all liness found in world state
func (s *SmartContract) QueryAllLines(ctx contractapi.TransactionContextInterface) QueryResults {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	var emptylines Lines
	if err != nil {
		return QueryResults{
			Code: 402,
			Msg:  err.Error(),
			Data: nil,
		}
	}
	defer resultsIterator.Close()

	var lines Lines
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return QueryResults{
				Code: 402,
				Msg:  err.Error(),
				Data: nil,
			}
		}

		var line Line
		err = json.Unmarshal(queryResponse.Value, &line)
		if err != nil {
			return QueryResults{
				Code: 402,
				Msg:  err.Error(),
				Data: nil,
			}
		}
		lines = append(lines, line)
	}

	if lines == nil {
		return QueryResults{
			Code: 402,
			Msg:  err.Error(),
			Data: nil,
		}
	}

	return QueryResults{
		Code: 200,
		Msg:  "",
		Data: lines,
	}
}
