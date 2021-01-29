//@author: hdsfade
//@date: 2021-01-26-12:53
package chaincode

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

var cargoIndexName = "cargo"

type Cargo struct { //货物清单
	TrainNumber        string   `json:"trainNumber"`
	TotalTypeNum       int      `json:"totalTypeNum"`
	CargoType          []string `json:"cargoType"`
	GoodsNum           []int    `json:"goodsNum"`
	GoodsName          []string `json:"goodsName"`
	GoodsOrderId       []int    `json"goodsOrderId"`
	StationCheckResult []bool   `json:"stationCheckResult"`
	CheckDescription   []string `json:"checkDescription"`
}

//CargoQueryResult structure used for handing result of query
type CargoQueryResult struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data Cargo  `json:"data"`
}

//CargoExists judges a order if exists or not
func (s *SmartContract) CargoExists(ctx contractapi.TransactionContextInterface, trainNumber string) (bool, error) {
	cargoIndexKey, err := ctx.GetStub().CreateCompositeKey(cargoIndexName, []string{trainNumber})
	if err != nil {
		return false, fmt.Errorf("failed to read from world state %v", err)
	}

	cargoJSON, err := ctx.GetStub().GetState(cargoIndexKey)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state %v", err)
	}
	return cargoJSON != nil, nil
}

//CreateCargo issues a new cargo to the world state with orders.
func (s *SmartContract) CreateCargo(ctx contractapi.TransactionContextInterface, trainNumber string) Result {
	cargoIndexKey, err := ctx.GetStub().CreateCompositeKey(cargoIndexName, []string{trainNumber})
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}

	exists, err := s.CargoExists(ctx, trainNumber)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	if exists {
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("the cargo %s already exists", trainNumber),
		}
	}

	//if the train trainNumber does not exist, the order couldn't be created
	exists, err = s.TrainExists(ctx, trainNumber)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	if exists == false {
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("the train %s does not exist", trainNumber),
		}
	}

	cargo := Cargo{
		TrainNumber:        trainNumber,
		TotalTypeNum:       0,
		CargoType:          []string{},
		GoodsNum:           []int{},
		GoodsName:          []string{},
		GoodsOrderId:       []int{},
		StationCheckResult: []bool{},
		CheckDescription:   []string{},
	}

	//iterate all orders
	orderResultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(trainorderIndexName, []string{trainNumber})
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	var order Order
	for orderResultsIterator.HasNext() {
		orderQueryResponse, err := orderResultsIterator.Next()
		if err != nil {
			return Result{
				Code: 402,
				Msg:  err.Error(),
			}
		}
		orderId := orderQueryResponse.Value
		orderIndexKey, err := ctx.GetStub().CreateCompositeKey(orderIndexName, []string{string(orderId)})
		if err != nil {
			return Result{
				Code: 402,
				Msg:  err.Error(),
			}
		}
		orderJSON, err := ctx.GetStub().GetState(orderIndexKey)
		if err != nil {
			return Result{
				Code: 402,
				Msg:  err.Error(),
			}
		}
		if orderJSON == nil {
			return Result{
				Code: 402,
				Msg:  fmt.Sprintf("the order %s does not exist", orderId),
			}
		}
		err = json.Unmarshal(orderJSON, &order)
		if err != nil {
			return Result{
				Code: 402,
				Msg:  err.Error(),
			}
		}
		if order.CheckResult {
			cargo.TotalTypeNum += order.TotalTypeNum
			cargo.CargoType = append(cargo.CargoType, order.CargoType...)
			cargo.GoodsNum = append(cargo.GoodsNum, order.GoodsNum...)
			cargo.GoodsName = append(cargo.GoodsName, order.GoodsName...)
			cargo.GoodsOrderId = append(cargo.GoodsOrderId, order.OrderId)
		}
	}
	cargoJSON, err := json.Marshal(cargo)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}

	err = ctx.GetStub().PutState(cargoIndexKey, cargoJSON)
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

//DeleteCargo deletes a cargo by trainNumber from the world state.
func (s *SmartContract) DeleteCargo(ctx contractapi.TransactionContextInterface, trainNumber string) Result {
	trainIndexKey, err := ctx.GetStub().CreateCompositeKey(trainIndexName, []string{trainNumber})
	exists, err := s.CargoExists(ctx, trainNumber)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	if !exists {
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("the cargo %s does not exist", trainNumber),
		}
	}

	err = ctx.GetStub().DelState(trainIndexKey)
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

//UpdateCargo updates an existing cargo in the world state with provided parameters
func (s *SmartContract) UpdateCargo(ctx contractapi.TransactionContextInterface, trainNumber string, stationCheckResult bool, checkDescription string) Result {
	cargoIndexKey, err := ctx.GetStub().CreateCompositeKey(cargoIndexName, []string{trainNumber})
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	cargoJSON, err := ctx.GetStub().GetState(cargoIndexKey)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	if cargoJSON == nil {
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("the cargo %s does not exist", trainNumber),
		}
	}

	var cargo Cargo
	err = json.Unmarshal(cargoJSON, &cargo)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	//overwriting original details
	cargo.StationCheckResult = append(cargo.StationCheckResult, stationCheckResult)
	cargo.CheckDescription = append(cargo.CheckDescription, checkDescription)
	cargoJSON, err = json.Marshal(cargo)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	err = ctx.GetStub().PutState(cargoIndexKey, cargoJSON)
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

// CheckCargo updates cargo's details
func (s *SmartContract) CheckCargo(ctx contractapi.TransactionContextInterface, trainNumber string, stationCheckResult bool, checkDescription string) Result {
	return s.UpdateCargo(ctx, trainNumber, stationCheckResult, checkDescription)
}

//QueryCargoBytrainnumber returns the cargo in the world state with given trainnumber
func (s *SmartContract) QueryCargoBytrainnumber(ctx contractapi.TransactionContextInterface, trainNumber string) CargoQueryResult {
	cargoIndexKey, err := ctx.GetStub().CreateCompositeKey(cargoIndexName, []string{trainNumber})
	if err != nil {
		return CargoQueryResult{
			Code: 402,
			Msg:  err.Error(),
			Data: Cargo{
				TrainNumber:        " ",
				TotalTypeNum:       0,
				CargoType:          []string{},
				GoodsNum:           []int{},
				GoodsName:          []string{},
				GoodsOrderId:       []int{},
				StationCheckResult: []bool{},
				CheckDescription:   []string{},
			},
		}
	}

	cargoJSON, err := ctx.GetStub().GetState(cargoIndexKey)
	if err != nil {
		return CargoQueryResult{
			Code: 402,
			Msg:  fmt.Sprintf("failed to read from world state: %v", err),
			Data: Cargo{
				TrainNumber:        " ",
				TotalTypeNum:       0,
				CargoType:          []string{},
				GoodsNum:           []int{},
				GoodsName:          []string{},
				GoodsOrderId:       []int{},
				StationCheckResult: []bool{},
				CheckDescription:   []string{},
			},
		}
	}
	if cargoJSON == nil {
		return CargoQueryResult{
			Code: 402,
			Msg:  fmt.Sprintf("the cargo %s does not exist", trainNumber),
			Data: Cargo{
				TrainNumber:        " ",
				TotalTypeNum:       0,
				CargoType:          []string{},
				GoodsNum:           []int{},
				GoodsName:          []string{},
				GoodsOrderId:       []int{},
				StationCheckResult: []bool{},
				CheckDescription:   []string{},
			},
		}
	}

	var cargo Cargo
	err = json.Unmarshal(cargoJSON, &cargo)
	if err != nil {
		return CargoQueryResult{
			Code: 402,
			Msg:  err.Error(),
			Data: Cargo{
				TrainNumber:        " ",
				TotalTypeNum:       0,
				CargoType:          []string{},
				GoodsNum:           []int{},
				GoodsName:          []string{},
				GoodsOrderId:       []int{},
				StationCheckResult: []bool{},
				CheckDescription:   []string{},
			},
		}
	}

	return CargoQueryResult{
		Code: 200,
		Msg:  "success",
		Data: cargo,
	}
}
