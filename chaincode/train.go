//@author: hdsfade
//@date: 2021-01-20-09:29
package chaincode

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

var trainIndexName = "train"
var trainOrderIndexName = "train~order"

//Train describe details of a train
type Train struct {
	TrainNumber  string `json:"trainNumber"`
	CarriageLeft int    `json:"carriageLeft"`
}

type Trains struct {
	TrainsDate []Train `json:"trains"`
}

//TrainQueryResult structure used for handing result of query
type TrainQueryResult struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data Train  `json:"data"`
}

//TrainQueryResults structure used for handing result of queryAll
type TrainQueryResults struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data Trains `json:"data"`
}

//TrainExists judges a schedule if exists or not
func (s *SmartContract) TrainExists(ctx contractapi.TransactionContextInterface, trainNumber string) (bool, error) {
	trainIndexKey, err := ctx.GetStub().CreateCompositeKey(trainIndexName, []string{trainNumber})
	if err != nil {
		return false, fmt.Errorf("failed to read from world state %v", err)
	}

	trainJSON, err := ctx.GetStub().GetState(trainIndexKey)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state %v", err)
	}
	return trainJSON != nil, nil
}

//CreateTrain issues a new schedule to the world state with given details
func (s *SmartContract) CreateTrain(ctx contractapi.TransactionContextInterface, trainNumber string, carriageLeft int) Result {
	trainIndexKey, err := ctx.GetStub().CreateCompositeKey(trainIndexName, []string{trainNumber})
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	exists, err := s.TrainExists(ctx, trainNumber)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	if exists {
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("the train %s already exists", trainNumber),
		}
	}

	train := Train{
		TrainNumber:  trainNumber,
		CarriageLeft: carriageLeft,
	}
	trainJSON, err := json.Marshal(train)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}

	err = ctx.GetStub().PutState(trainIndexKey, trainJSON)
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

//UpdateTrain updates an existing train in the world state with provided parameters
func (s *SmartContract) UpdateTrain(ctx contractapi.TransactionContextInterface, trainNumber string, carriageNumber int) Result {
	trainIndexKey, err := ctx.GetStub().CreateCompositeKey(trainIndexName, []string{trainNumber})
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	trainJSON, err := ctx.GetStub().GetState(trainIndexKey)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	if trainJSON == nil {
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("the train %s does not exist", trainNumber),
		}
	}

	var train Train
	err = json.Unmarshal(trainJSON, &train)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	if train.CarriageLeft < carriageNumber {
		return Result{
			Code: 402,
			Msg: fmt.Sprintf("the train %s's carriageLeft is not enough: carriageLeft %d, carraigeNumber %d",
				trainNumber, train.CarriageLeft, carriageNumber),
		}
	}
	//overwriting original carriageLeft
	train.CarriageLeft -= carriageNumber
	trainJSON, err = json.Marshal(train)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	err = ctx.GetStub().PutState(trainIndexKey, trainJSON)
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

//QueryTrainBytrainnumber returns the train in the world state with given trainNumber
func (s *SmartContract) QueryTrainBytrainnumber(ctx contractapi.TransactionContextInterface, trainNumber string) TrainQueryResult {
	trainIndexKey, err := ctx.GetStub().CreateCompositeKey(trainIndexName, []string{trainNumber})
	if err != nil {
		return TrainQueryResult{
			Code: 402,
			Msg:  fmt.Sprintf("failed to read from world state: %v", err),
			Data: Train{
				TrainNumber:  "0",
				CarriageLeft: 0,
			},
		}
	}
	trainJSON, err := ctx.GetStub().GetState(trainIndexKey)

	if err != nil {
		return TrainQueryResult{
			Code: 402,
			Msg:  fmt.Sprintf("failed to read from world state: %v", err),
			Data: Train{
				TrainNumber:  "0",
				CarriageLeft: 0,
			},
		}
	}
	if trainJSON == nil {
		return TrainQueryResult{
			Code: 402,
			Msg:  fmt.Sprintf("the train %s does not exist", trainNumber),
			Data: Train{
				TrainNumber:  "0",
				CarriageLeft: 0,
			},
		}
	}

	var train Train
	err = json.Unmarshal(trainJSON, &train)
	if err != nil {
		return TrainQueryResult{
			Code: 402,
			Msg:  fmt.Sprintf("failed to read from world state: %v", err),
			Data: Train{
				TrainNumber:  "0",
				CarriageLeft: 0,
			},
		}
	}
	return TrainQueryResult{
		Code: 200,
		Msg:  "success",
		Data: train,
	}
}

//QueryAllTrains returns all trains found in world state
func (s *SmartContract) QueryAllTrains(ctx contractapi.TransactionContextInterface) TrainQueryResults {
	var emptytrains []Train
	emptytrains = append(emptytrains, Train{
		TrainNumber:  "0",
		CarriageLeft: 0,
	})

	trainResultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(trainIndexName, []string{})
	if err != nil {
		return TrainQueryResults{
			Code: 402,
			Msg:  err.Error(),
			Data: Trains{TrainsDate: emptytrains},
		}
	}
	defer trainResultsIterator.Close()

	var trains []Train
	for trainResultsIterator.HasNext() {
		trainQueryResponse, err := trainResultsIterator.Next()
		if err != nil {
			return TrainQueryResults{
				Code: 402,
				Msg:  err.Error(),
				Data: Trains{TrainsDate: emptytrains},
			}
		}

		var train Train
		err = json.Unmarshal(trainQueryResponse.Value, &train)
		if err != nil {
			return TrainQueryResults{
				Code: 402,
				Msg:  err.Error(),
				Data: Trains{TrainsDate: emptytrains},
			}
		}
		trains = append(trains, train)
	}

	if trains == nil {
		return TrainQueryResults{
			Code: 402,
			Msg:  "No train",
			Data: Trains{TrainsDate: emptytrains},
		}
	}

	return TrainQueryResults{
		Code: 200,
		Msg:  "success",
		Data: Trains{TrainsDate: trains},
	}
}
