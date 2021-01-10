//@author: hdsfade
//@date: 2021-01-06-14:39
package station

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

//Station describes details of an station
type Station struct {
	StationName string `json:"stationName"`
	Country     string `json:"country"`
	Using       bool   `json:"using"`
	Describtion string `json:"describtion"`
}

type Stations []Station

//Result structure used for handing result of create or delete
type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

//QueryResult structure used for handing result of query
type QueryResult struct {
	Code int     `json:"code"`
	Msg  string  `json:"msg"`
	Data Station `json:"data"`
}

//QueryResult structure used for handing result of queryAll
type QueryResults struct {
	Code int      `json:"code"`
	Msg  string   `json:"msg"`
	Data Stations `json:"data"`
}

// Init stations' ledger(can add a default set of stations to the ledger)
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	stations := []Station{
		{StationName: "上海", Country: "中国", Using: true, Describtion: ""},
		{StationName: "北京", Country: "中国", Using: true, Describtion: ""},
		{StationName: "广州", Country: "中国", Using: true, Describtion: ""},
		{StationName: "杭州", Country: "中国", Using: true, Describtion: ""},
		{StationName: "宁波", Country: "中国", Using: true, Describtion: ""},
	}
	for _, station := range stations {
		stationJSON, err := json.Marshal(station)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(station.StationName, stationJSON)
		if err != nil {
			return nil
		}
	}
	return nil
}

//StationExists judges a station if exists or not.
func (s *SmartContract) StationExists(ctx contractapi.TransactionContextInterface, stationName string) (bool, error) {
	stationJSON, err := ctx.GetStub().GetState(stationName)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state %v", err)
	}
	return stationJSON != nil, nil
}

//CreateStation issues a new station to the world state with given details.
func (s *SmartContract) CreateStation(ctx contractapi.TransactionContextInterface, stationName, country string, description string) Result {
	exists, err := s.StationExists(ctx, stationName)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	if exists {
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("the station %s already exists", stationName),
		}
	}

	station := Station{
		StationName: stationName,
		Country:     country,
		Using:       true,
		Describtion: description,
	}
	stationJSON, err := json.Marshal(station)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}

	err = ctx.GetStub().PutState(stationName, stationJSON)
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

//DeleteStation deletes an station by stationName from the world state.
func (s *SmartContract) DeleteStation(ctx contractapi.TransactionContextInterface, stationName string) Result {
	exists, err := s.StationExists(ctx, stationName)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	if !exists {
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("the station %s does not exist", stationName),
		}
	}
	err = ctx.GetStub().DelState(stationName)
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

// QueryStationBystationname returns the station stored in the world state with given stationName
func (s *SmartContract) QueryStationBystationname(ctx contractapi.TransactionContextInterface, stationName string) QueryResult {
	stationJSON, err := ctx.GetStub().GetState(stationName)
	if err != nil {
		return QueryResult{
			Code: 402,
			Msg:  fmt.Sprintf("failed to read from world state: %v", err),
			Data: Station{},
		}
	}
	if stationJSON == nil {
		return QueryResult{
			Code: 402,
			Msg:  fmt.Sprintf("the station %s does not exist", stationName),
			Data: Station{},
		}
	}

	var station Station
	err = json.Unmarshal(stationJSON, &station)
	if err != nil {
		return QueryResult{
			Code: 402,
			Msg:  err.Error(),
			Data: Station{},
		}
	}
	return QueryResult{
		Code: 200,
		Msg:  "",
		Data: station,
	}
}

// QueryAllStations returns all stations found in world state
func (s *SmartContract) QueryAllStations(ctx contractapi.TransactionContextInterface) QueryResults {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return QueryResults{
			Code: 402,
			Msg:  err.Error(),
			Data: Stations{},
		}
	}
	defer resultsIterator.Close()

	var stations Stations
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return QueryResults{
				Code: 402,
				Msg:  err.Error(),
				Data: Stations{},
			}
		}

		var station Station
		err = json.Unmarshal(queryResponse.Value, &station)
		if err != nil {
			return QueryResults{
				Code: 402,
				Msg:  err.Error(),
				Data: Stations{},
			}
		}
		stations = append(stations, station)
	}

	return QueryResults{
		Code: 200,
		Msg:  "",
		Data: stations,
	}
}
