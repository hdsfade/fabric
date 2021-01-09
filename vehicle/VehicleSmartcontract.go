//@author: hdsfade
//@date: 2021-01-06-15:03
package vehicle

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

//Vehicle describes details of a vehicle
type Vehicle struct { //车辆
	VehicleNumber int  `json:"vehicleNumber"`
	CarriageNum   int  `json:"carriageNum"`
	Using         bool `json:"using"`
}

type Vehicles []Vehicle

//Result structure used for handing result of create or delete
type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

//QueryResult structure used for handing result of query
type QueryResult struct {
	Code int     `json:"code"`
	Msg  string  `json:"msg"`
	Data Vehicle `json:"data"`
}

//QueryResult structure used for handing result of queryAll
type QueryResults struct {
	Code int      `json:"code"`
	Msg  string   `json:"msg"`
	Data Vehicles `json:"data"`
}

// Init vehicles' ledger(can add a default set of vehicles to the ledger)
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	vehicles := []Vehicle{
		{VehicleNumber: 0001, CarriageNum: 10, Using: true},
		{VehicleNumber: 0002, CarriageNum: 15, Using: true},
		{VehicleNumber: 0003, CarriageNum: 18, Using: true},
		{VehicleNumber: 0004, CarriageNum: 8, Using: true},
		{VehicleNumber: 0005, CarriageNum: 12, Using: true},
	}
	for _, vehicle := range vehicles {
		vehicleJSON, err := json.Marshal(vehicle)
		if err != nil {
			return nil
		}

		err = ctx.GetStub().PutState(string(vehicle.VehicleNumber), vehicleJSON)
		if err != nil {
			return nil
		}
	}
	return nil
}

//Vehicle Exists judges a vehicle if exists or not.
func (s *SmartContract) VehicleExists(ctx contractapi.TransactionContextInterface, vehicleNumber int) (bool, error) {
	vehicleJSON, err := ctx.GetStub().GetState(string(vehicleNumber))
	if err != nil {
		return false, fmt.Errorf("failed to read from world state %v", err)
	}
	return vehicleJSON != nil, nil
}

//CreateVehicle issues a new vehicle to the world state with given details.
func (s *SmartContract) CreateVehicle(ctx contractapi.TransactionContextInterface, vehicleNumber, carriageNum int) Result {
	exists, err := s.VehicleExists(ctx, vehicleNumber)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	if exists {
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("the vehicle %d already exists", vehicleNumber),
		}
	}

	vehicle := Vehicle{
		VehicleNumber: vehicleNumber,
		CarriageNum:   carriageNum,
		Using:         true,
	}
	vehicleJSON, err := json.Marshal(vehicle)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}

	err = ctx.GetStub().PutState(string(vehicleNumber), vehicleJSON)
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

//DeleteVehicle deletes a vehicle by vehicleNumber from the world state.
func (s *SmartContract) DeleteVehicle(ctx contractapi.TransactionContextInterface, vehicleNumber int) Result {
	exists, err := s.VehicleExists(ctx, vehicleNumber)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	if !exists {
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("the vehicle %d does not exist", vehicleNumber),
		}
	}
	return Result{
		Code: 200,
		Msg:  "",
	}
}

// QueryVehicleByvehiclenumber returns the vehicles stored in the world state with given vehicleNumber
func (s *SmartContract) QueryVehicleByvehiclenumber(ctx contractapi.TransactionContextInterface, vehicleNumber int) QueryResult {
	vehicleJSON, err := ctx.GetStub().GetState(string(vehicleNumber))
	if err != nil {
		return QueryResult{
			Code: 402,
			Msg:  fmt.Sprintf("failed to read from world state: %v", err),
			Data: Vehicle{},
		}
	}
	if vehicleJSON == nil {
		return QueryResult{
			Code: 402,
			Msg:  fmt.Sprintf("the vehicle %d does not exist", vehicleNumber),
			Data: Vehicle{},
		}
	}

	var vehicle Vehicle
	err = json.Unmarshal(vehicleJSON, &vehicle)
	if err != nil {
		return QueryResult{
			Code: 402,
			Msg:  err.Error(),
			Data: Vehicle{},
		}
	}
	return QueryResult{
		Code: 200,
		Msg:  "",
		Data: vehicle,
	}
}

// QueryAllVehicles returns all vehicles found in world state
func (s *SmartContract) QueryAllVehicles(ctx contractapi.TransactionContextInterface) QueryResults {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return QueryResults{
			Code: 402,
			Msg:  err.Error(),
			Data: Vehicles{},
		}
	}
	defer resultsIterator.Close()

	var vehicles Vehicles
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return QueryResults{
				Code: 402,
				Msg:  err.Error(),
				Data: Vehicles{},
			}
		}

		var vehicle Vehicle
		err = json.Unmarshal(queryResponse.Value, &vehicle)
		if err != nil {
			return QueryResults{
				Code: 402,
				Msg:  err.Error(),
				Data: Vehicles{},
			}
		}
		vehicles = append(vehicles, vehicle)
	}
	return QueryResults{
		Code: 200,
		Msg:  "",
		Data: vehicles,
	}
}
