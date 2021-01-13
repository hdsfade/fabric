//@author: hdsfade
//@date: 2021-01-13-21:45
package chaincode

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"strconv"
)

//Vehicle describes details of a vehicle
type Vehicle struct { //车辆
	VehicleNumber int  `json:"vehicleNumber"`
	CarriageNum   int  `json:"carriageNum"`
	Using         bool `json:"using"`
}

type Vehicles []Vehicle

//VehicleQueryResult structure used for handing result of query vehicle
type VehicleQueryResult struct {
	Code int     `json:"code"`
	Msg  string  `json:"msg"`
	Data Vehicle `json:"data"`
}

//VehicleQueryResults structure used for handing result of query all vehicles
type VehicleQueryResults struct {
	Code int      `json:"code"`
	Msg  string   `json:"msg"`
	Data Vehicles `json:"data"`
}

//Vehicle Exists judges a vehicle if exists or not.
func (s *SmartContract) VehicleExists(ctx contractapi.TransactionContextInterface, vehicleNumber int) (bool, error) {
	vehicleIndexKey, err := ctx.GetStub().CreateCompositeKey(vehicleIndexName, []string{strconv.Itoa(vehicleNumber)})
	if err != nil {
		return false, fmt.Errorf("failed to read from world state %v", err)
	}

	vehicleJSON, err := ctx.GetStub().GetState(vehicleIndexKey)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state %v", err)
	}
	return vehicleJSON != nil, nil
}

//CreateVehicle issues a new vehicle to the world state with given details.
func (s *SmartContract) CreateVehicle(ctx contractapi.TransactionContextInterface, vehicleNumber, carriageNum int) Result {
	vehicleIndexKey, err := ctx.GetStub().CreateCompositeKey(vehicleIndexName, []string{strconv.Itoa(vehicleNumber)})

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

	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	err = ctx.GetStub().PutState(vehicleIndexKey, vehicleJSON)
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
	vehicleIndexKey, err := ctx.GetStub().CreateCompositeKey(vehicleIndexName, []string{strconv.Itoa(vehicleNumber)})
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

	err = ctx.GetStub().DelState(vehicleIndexKey)
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

// QueryVehicleByvehiclenumber returns the vehicles stored in the world state with given vehicleNumber
func (s *SmartContract) QueryVehicleByvehiclenumber(ctx contractapi.TransactionContextInterface, vehicleNumber int) VehicleQueryResult {
	vehicleIndexKey, err := ctx.GetStub().CreateCompositeKey(vehicleIndexName, []string{string(vehicleNumber)})
	if err != nil {
		return VehicleQueryResult{
			Code: 402,
			Msg:  fmt.Sprintf("failed to read from world state: %v", err),
			Data: Vehicle{},
		}
	}

	vehicleJSON, err := ctx.GetStub().GetState(vehicleIndexKey)
	if err != nil {
		return VehicleQueryResult{
			Code: 402,
			Msg:  fmt.Sprintf("failed to read from world state: %v", err),
			Data: Vehicle{},
		}
	}
	if vehicleJSON == nil {
		return VehicleQueryResult{
			Code: 402,
			Msg:  fmt.Sprintf("the vehicle %d does not exist", vehicleNumber),
			Data: Vehicle{},
		}
	}

	var vehicle Vehicle
	err = json.Unmarshal(vehicleJSON, &vehicle)
	if err != nil {
		return VehicleQueryResult{
			Code: 402,
			Msg:  err.Error(),
			Data: Vehicle{},
		}
	}
	return VehicleQueryResult{
		Code: 200,
		Msg:  "",
		Data: vehicle,
	}
}

// QueryAllVehicles returns all vehicles found in world state
func (s *SmartContract) QueryAllVehicles(ctx contractapi.TransactionContextInterface) VehicleQueryResults {
	vehicleResultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(vehicleIndexName, []string{})
	if err != nil {
		return VehicleQueryResults{
			Code: 402,
			Msg:  err.Error(),
			Data: Vehicles{},
		}
	}
	defer vehicleResultsIterator.Close()

	var vehicles Vehicles
	for vehicleResultsIterator.HasNext() {
		vehicleQueryResponse, err := vehicleResultsIterator.Next()
		if err != nil {
			return VehicleQueryResults{
				Code: 402,
				Msg:  err.Error(),
				Data: Vehicles{},
			}
		}

		var vehicle Vehicle
		err = json.Unmarshal(vehicleQueryResponse.Value, &vehicle)
		if err != nil {
			return VehicleQueryResults{
				Code: 402,
				Msg:  err.Error(),
				Data: Vehicles{},
			}
		}
		vehicles = append(vehicles, vehicle)
	}
	if vehicles == nil {
		return VehicleQueryResults{
			Code: 402,
			Msg:  "No vehicle",
			Data: Vehicles{},
		}
	}
	return VehicleQueryResults{
		Code: 200,
		Msg:  "",
		Data: vehicles,
	}
}
