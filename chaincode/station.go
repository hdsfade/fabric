//@author: hdsfade
//@date: 2021-01-13-21:44
package chaincode

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

//Station describes details of an station
type Station struct {
	StationName string `json:"stationName"`
	Country     string `json:"country"`
	Using       bool   `json:"using"`
	Describtion string `json:"describtion"`
}

type Stations struct {
	StationsData []Station `json:"stations"`
}

//QueryResult structure used for handing result of query
type StationQueryResult struct {
	Code int     `json:"code"`
	Msg  string  `json:"msg"`
	Data Station `json:"data"`
}

//QueryResult structure used for handing result of queryAll
type StationQueryResults struct {
	Code int      `json:"code"`
	Msg  string   `json:"msg"`
	Data Stations `json:"data"`
}

//StationExists judges a station if exists or not.
func (s *SmartContract) StationExists(ctx contractapi.TransactionContextInterface, stationName string) (bool, error) {
	stationIndexKey, err := ctx.GetStub().CreateCompositeKey(stationIndexName, []string{stationName})
	if err != nil {
		return false, fmt.Errorf("failed to read from world state %v", err)
	}
	stationJSON, err := ctx.GetStub().GetState(stationIndexKey)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state %v", err)
	}
	return stationJSON != nil, nil
}

//CreateStation issues a new station to the world state with given details.
func (s *SmartContract) CreateStation(ctx contractapi.TransactionContextInterface, stationName, country string, description string) Result {
	stationIndexKey, err := ctx.GetStub().CreateCompositeKey(stationIndexName, []string{stationName})

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

	err = ctx.GetStub().PutState(stationIndexKey, stationJSON)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	return Result{
		Code: 200,
		Msg:  "success",
	}
}

//DeleteStation deletes an station by stationName from the world state.
func (s *SmartContract) DeleteStation(ctx contractapi.TransactionContextInterface, stationName string) Result {
	stationIndexKey, err := ctx.GetStub().CreateCompositeKey(stationIndexName, []string{stationName})
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

	stationResultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("station~line", []string{stationName})
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	if stationResultsIterator.HasNext() == true {
		var useStationLines string
		for stationResultsIterator.HasNext() {
			stationQueryResponse, err := stationResultsIterator.Next()
			if err != nil {
				return Result{
					Code: 402,
					Msg:  err.Error(),
				}
			}
			_, compositeKeyParts, err := ctx.GetStub().SplitCompositeKey(stationQueryResponse.Key)
			if err != nil {
				return Result{
					Code: 402,
					Msg:  err.Error(),
				}
			}
			useStationLines += compositeKeyParts[1] + " "
		}
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("the station %s is used by lines %v", stationName, useStationLines),
		}
	}

	err = ctx.GetStub().DelState(stationIndexKey)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	return Result{
		Code: 200,
		Msg:  "success",
	}
}

// QueryStationBystationname returns the station stored in the world state with given stationName
func (s *SmartContract) QueryStationBystationname(ctx contractapi.TransactionContextInterface, stationName string) StationQueryResult {
	stationIndexKey, err := ctx.GetStub().CreateCompositeKey(stationIndexName, []string{stationName})
	if err != nil {
		return StationQueryResult{
			Code: 402,
			Msg:  fmt.Sprintf("failed to read from world state: %v", err),
			Data: Station{},
		}
	}

	stationJSON, err := ctx.GetStub().GetState(stationIndexKey)
	if err != nil {
		return StationQueryResult{
			Code: 402,
			Msg:  fmt.Sprintf("failed to read from world state: %v", err),
			Data: Station{},
		}
	}
	if stationJSON == nil {
		return StationQueryResult{
			Code: 402,
			Msg:  fmt.Sprintf("the station %s does not exist", stationName),
			Data: Station{},
		}
	}

	var station Station
	err = json.Unmarshal(stationJSON, &station)
	if err != nil {
		return StationQueryResult{
			Code: 402,
			Msg:  err.Error(),
			Data: Station{},
		}
	}
	return StationQueryResult{
		Code: 200,
		Msg:  "success",
		Data: station,
	}
}

// QueryAllStations returns all stations found in world state
func (s *SmartContract) QueryAllStations(ctx contractapi.TransactionContextInterface) StationQueryResults {
	stationResultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(stationIndexName, []string{})
	if err != nil {
		return StationQueryResults{
			Code: 402,
			Msg:  err.Error(),
			Data: Stations{StationsData: []Station{}},
		}
	}
	defer stationResultsIterator.Close()

	var stations []Station
	for stationResultsIterator.HasNext() {
		stationQueryResponse, err := stationResultsIterator.Next()
		if err != nil {
			return StationQueryResults{
				Code: 402,
				Msg:  err.Error(),
				Data: Stations{StationsData: []Station{}},
			}
		}

		var station Station
		err = json.Unmarshal(stationQueryResponse.Value, &station)
		if err != nil {
			return StationQueryResults{
				Code: 402,
				Msg:  err.Error(),
				Data: Stations{StationsData: []Station{}},
			}
		}
		stations = append(stations, station)
	}

	if stations == nil {
		return StationQueryResults{
			Code: 402,
			Msg:  "No station",
			Data: Stations{StationsData: []Station{}},
		}
	}

	return StationQueryResults{
		Code: 200,
		Msg:  "success",
		Data: Stations{StationsData: stations},
	}
}
