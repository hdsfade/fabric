//@author: hdsfade
//@date: 2021-01-13-21:45
package chaincode

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"strconv"
)

//Line describes details of a line
type Line struct {
	LineNumber     int      `json:"lineNumber"`
	WayStation     []string `json:"wayStation"`
	WayStationType []string `json:"wayStationType"`
	Using          bool     `json:"using"`
}

type Lines struct {
	LinesData []Line `json:"lines"`
}

//LineQueryResult structure used for handing result of query
type LineQueryResult struct {
	Code    int      `json:"code"`
	Msg     string   `json:"msg"`
	Data    Line     `json:"data"`
	SubData Stations `json:"subData"`
}

//LineQueryResults structure used for handing result of queryAll
type LineQueryResults struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data Lines  `json:"data"`
}

//LineExists judges a line if exists or not.
func (s *SmartContract) LineExists(ctx contractapi.TransactionContextInterface, lineNumber int) (bool, error) {
	lineIndexKey, err := ctx.GetStub().CreateCompositeKey(lineIndexName, []string{strconv.Itoa(lineNumber)})
	if err != nil {
		return false, fmt.Errorf("failed to read from world state %v", err)
	}

	lineJSON, err := ctx.GetStub().GetState(lineIndexKey)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state %v", err)
	}
	return lineJSON != nil, nil
}

//CreateLine issues a new line to the world state with given details.
func (s *SmartContract) CreateLine(ctx contractapi.TransactionContextInterface, lineNumber int, wayStation, wayStationType []string) Result {
	lineIndexKey, err := ctx.GetStub().CreateCompositeKey(lineIndexName, []string{strconv.Itoa(lineNumber)})
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}

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

	for _, stationName := range wayStation {
		exists, err := s.StationExists(ctx, stationName)
		if err != nil {
			return Result{
				Code: 402,
				Msg:  err.Error(),
			}
		}
		if exists == false {
			return Result{
				Code: 402,
				Msg:  fmt.Sprintf("the station %s does not exist", stationName),
			}
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

	err = ctx.GetStub().PutState(lineIndexKey, lineJSON)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}

	//create compositekey station~line
	value := []byte{0x00}
	for _, stationName := range line.WayStation {
		stationlineIndexKey, err := ctx.GetStub().CreateCompositeKey(stationlineIndexName, []string{stationName, strconv.Itoa(line.LineNumber)})
		if err != nil {
			return Result{
				Code: 402,
				Msg:  err.Error(),
			}
		}
		err = ctx.GetStub().PutState(stationlineIndexKey, value)
		if err != nil {
			return Result{
				Code: 402,
				Msg:  err.Error(),
			}
		}
	}

	return Result{
		Code: 200,
		Msg:  "success",
	}
}

//DeleteLine deletes a line by lineNumber from the world state.
func (s *SmartContract) DeleteLine(ctx contractapi.TransactionContextInterface, lineNumber int) Result {
	lineIndexKey, err := ctx.GetStub().CreateCompositeKey(lineIndexName, []string{strconv.Itoa(lineNumber)})
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}

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

	//if the line is used by some schedules, the line couldn't be delete.
	lineResultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(linescheduleIndexName, []string{strconv.Itoa(lineNumber)})
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	if lineResultsIterator.HasNext() == true {
		var useLineSchedules string
		for lineResultsIterator.HasNext() {
			lineQueryResponse, err := lineResultsIterator.Next()
			if err != nil {
				return Result{
					Code: 402,
					Msg:  err.Error(),
				}
			}
			_, compositeKeyParts, err := ctx.GetStub().SplitCompositeKey(lineQueryResponse.Key)
			if err != nil {
				return Result{
					Code: 402,
					Msg:  err.Error(),
				}
			}
			useLineSchedules += compositeKeyParts[1] + " "
		}
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("the line %d is used by schedules %s", lineNumber, useLineSchedules),
		}
	}

	lineJSON, err := ctx.GetStub().GetState(lineIndexKey)
	var line Line
	err = json.Unmarshal(lineJSON, &line)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}

	for _, stationName := range line.WayStation {
		stationLineIndexKey, err := ctx.GetStub().CreateCompositeKey(stationlineIndexName, []string{stationName, strconv.Itoa(line.LineNumber)})
		if err != nil {
			return Result{
				Code: 402,
				Msg:  err.Error(),
			}
		}
		err = ctx.GetStub().DelState(stationLineIndexKey)
		if err != nil {
			return Result{
				Code: 402,
				Msg:  err.Error(),
			}
		}
	}

	err = ctx.GetStub().DelState(lineIndexKey)
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

// QueryLineBylinenumber returns the line stored in the world state with given lineNumver
func (s *SmartContract) QueryLineBylinenumber(ctx contractapi.TransactionContextInterface, lineNumber int) LineQueryResult {
	lineIndexKey, err := ctx.GetStub().CreateCompositeKey(lineIndexName, []string{strconv.Itoa(lineNumber)})
	if err != nil {
		return LineQueryResult{
			Code: 402,
			Msg:  fmt.Sprintf("failed to read from world state: %v", err),
			Data: Line{
				LineNumber:     0,
				WayStation:     []string{},
				WayStationType: []string{},
				Using:          false,
			},
			SubData: Stations{StationsData: []Station{
				{
					StationName: " ",
					Country:     " ",
					Using:       false,
					Describtion: " ",
				},
			}},
		}
	}

	lineJSON, err := ctx.GetStub().GetState(lineIndexKey)
	if err != nil {
		return LineQueryResult{
			Code: 402,
			Msg:  fmt.Sprintf("failed to read from world state: %v", err),
			Data: Line{
				LineNumber:     0,
				WayStation:     []string{},
				WayStationType: []string{},
				Using:          false,
			},
			SubData: Stations{StationsData: []Station{
				{
					StationName: " ",
					Country:     " ",
					Using:       false,
					Describtion: " ",
				},
			}},
		}
	}
	if lineJSON == nil {
		return LineQueryResult{
			Code: 402,
			Msg:  fmt.Sprintf("the line %d doesnto exist", lineNumber),
			Data: Line{
				LineNumber:     0,
				WayStation:     []string{},
				WayStationType: []string{},
				Using:          false,
			},
			SubData: Stations{StationsData: []Station{
				{
					StationName: " ",
					Country:     " ",
					Using:       false,
					Describtion: " ",
				},
			}},
		}
	}

	var line Line
	err = json.Unmarshal(lineJSON, &line)
	if err != nil {
		return LineQueryResult{
			Code: 402,
			Msg:  fmt.Sprintf("failed to read from world state: %v", err),
			Data: Line{
				LineNumber:     0,
				WayStation:     []string{},
				WayStationType: []string{},
				Using:          false,
			},
			SubData: Stations{StationsData: []Station{
				{
					StationName: " ",
					Country:     " ",
					Using:       false,
					Describtion: " ",
				},
			}},
		}
	}

	var stations []Station
	for _, stationName := range line.WayStation {
		stationIndexKey, err := ctx.GetStub().CreateCompositeKey(stationIndexName, []string{stationName})
		if err != nil {
			return LineQueryResult{
				Code: 402,
				Msg:  fmt.Sprintf("failed to read from world state: %v", err),
				Data: Line{
					LineNumber:     0,
					WayStation:     []string{},
					WayStationType: []string{},
					Using:          false,
				},
				SubData: Stations{StationsData: []Station{
					{
						StationName: " ",
						Country:     " ",
						Using:       false,
						Describtion: " ",
					},
				}},
			}
		}

		stationJSON, err := ctx.GetStub().GetState(stationIndexKey)
		if err != nil {
			return LineQueryResult{
				Code: 402,
				Msg:  fmt.Sprintf("failed to read from world state: %v", err),
				Data: Line{
					LineNumber:     0,
					WayStation:     []string{},
					WayStationType: []string{},
					Using:          false,
				},
				SubData: Stations{StationsData: []Station{
					{
						StationName: " ",
						Country:     " ",
						Using:       false,
						Describtion: " ",
					},
				}},
			}
		}
		var station Station
		err = json.Unmarshal(stationJSON, &station)
		if err != nil {
			return LineQueryResult{
				Code: 402,
				Msg:  fmt.Sprintf("failed to read from world state: %v", err),
				Data: Line{
					LineNumber:     0,
					WayStation:     []string{},
					WayStationType: []string{},
					Using:          false,
				},
				SubData: Stations{StationsData: []Station{
					{
						StationName: " ",
						Country:     " ",
						Using:       false,
						Describtion: " ",
					},
				}},
			}
		}
		stations = append(stations, station)
	}

	if stations == nil
		stations = append(stations, Station{
			StationName: " ",
			Country:     " ",
			Using:       false,
			Describtion: " ",
		})
	}
	return LineQueryResult{
		Code:    200,
		Msg:     "success",
		Data:    line,
		SubData: Stations{StationsData: stations},
	}
}

// QueryAllLines returns all lines found in world state
func (s *SmartContract) QueryAllLines(ctx contractapi.TransactionContextInterface) LineQueryResults {
	var emptylines []Line
	emptylines = append(emptylines, Line{
		LineNumber:     0,
		WayStation:     []string{},
		WayStationType: []string{},
		Using:          false,
	})

	lineResultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(lineIndexName, []string{})
	if err != nil {
		return LineQueryResults{
			Code: 402,
			Msg:  err.Error(),
			Data: Lines{LinesData: emptylines},
		}
	}
	defer lineResultsIterator.Close()

	var lines []Line
	for lineResultsIterator.HasNext() {
		lineQueryResponse, err := lineResultsIterator.Next()
		if err != nil {
			return LineQueryResults{
				Code: 402,
				Msg:  err.Error(),
				Data: Lines{LinesData: emptylines},
			}
		}

		var line Line
		err = json.Unmarshal(lineQueryResponse.Value, &line)
		if err != nil {
			return LineQueryResults{
				Code: 402,
				Msg:  err.Error(),
				Data: Lines{LinesData: emptylines},
			}
		}
		lines = append(lines, line)
	}
	if lines == nil {
		return LineQueryResults{
			Code: 402,
			Msg:  "No line",
			Data: Lines{LinesData: emptylines},
		}
	}

	return LineQueryResults{
		Code: 200,
		Msg:  "success",
		Data: Lines{LinesData: lines},
	}
}
