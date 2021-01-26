//@author: hdsfade
//@date: 2021-01-26-09:55
package chaincode

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"strconv"
)

var waybillIndexName = "waybill"

//WayBill describes details of a waybill
type WayBill struct { //运单
	TrainNumber       string   `json:"trainNumber"`
	WayStation        []string `json:"wayStation"`
	ArrivalTime       []string `json:"arrivalTime"`
	LeaveTime         []string `json:"leaveTime"`
	Location          int      `json:"location"` //列车当前位置，供实时查询
	StationTrainState bool     `json:"stationTrainState"`
	CheckDescription  string   `json:"checkDescription"`
}

//WayBillExists judges a waybill if exists or not.
func (s *SmartContract) WayBillExists(ctx contractapi.TransactionContextInterface, trainNumber string) (bool, error) {
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

////CreateWayBill issues a new line to the world state with given details.
func (s *SmartContract) CreateWayBill(ctx contractapi.TransactionContextInterface, trainNumber string) Result {
	waybillIndexKey, err := ctx.GetStub().CreateCompositeKey(waybillIndexName, []string{trainNumber})
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}

	exists, err := s.WayBillExists(ctx, trainNumber)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	if exists {
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("the waybill %s already exists", trainNumber),
		}
	}

	exists, err = s.TrainExists(ctx, trainNumber)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	if !exists {
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("the train %s does not exist", trainNumber),
		}
	}

	scheduleNumber, err := strconv.Atoi(trainNumber[8:12])
	if err != nil {
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("trainNumber error: %v", err),
		}
	}
	scheduleIndexKey, err := ctx.GetStub().CreateCompositeKey(scheduleIndexName, []string{strconv.Itoa(scheduleNumber)})
	if err != nil {
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("failed to read schedule %d from world state: %v", scheduleNumber, err),
		}
	}
	scheduleJSON, err := ctx.GetStub().GetState(scheduleIndexKey)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("failed to read schedule %d from world state: %v", scheduleNumber, err),
		}
	}
	if scheduleJSON == nil {
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("the schedule %d does not exist", scheduleNumber),
		}
	}
	var schedule Schedule
	err = json.Unmarshal(scheduleJSON, &schedule)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}

	lineIndexKey, err := ctx.GetStub().CreateCompositeKey(lineIndexName, []string{strconv.Itoa(schedule.LineNumber)})
	if err != nil {
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("failed to read schedule %d's line %d from world state: %v", scheduleNumber, schedule.LineNumber, err),
		}
	}
	lineJSON, err := ctx.GetStub().GetState(lineIndexKey)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("failed to read schedule %d's line %d from world state: %v", scheduleNumber, schedule.LineNumber, err),
		}
	}
	if lineJSON == nil {
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("the schedule %d's line %d does not exist", scheduleNumber, schedule.LineNumber),
		}
	}
	var line Line
	err = json.Unmarshal(lineJSON, &line)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}

	wayBill := WayBill{
		TrainNumber:       trainNumber,
		WayStation:        line.WayStation,
		ArrivalTime:       []string{},
		LeaveTime:         []string{},
		Location:          0,
		StationTrainState: false,
		CheckDescription:  " ",
	}
	wayBillJSON, err := json.Marshal(wayBill)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}

	err = ctx.GetStub().PutState(waybillIndexKey, wayBillJSON)
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

//UpdateWayBill updates an existing waybill in the world state with provided parameters
func (s *SmartContract) UpdateWayBill(ctx contractapi.TransactionContextInterface, trainNumber, arrivalTime, leaveTime string, location int,
	stationTrainState bool, checkDescription string) Result {
	waybillIndexKey, err := ctx.GetStub().CreateCompositeKey(waybillIndexName, []string{trainNumber})
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	wayBillJSON, err := ctx.GetStub().GetState(waybillIndexKey)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	if wayBillJSON == nil {
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("the waybill %s does not exist", trainNumber),
		}
	}

	var waybill WayBill
	err = json.Unmarshal(wayBillJSON, &waybill)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}

	//overwriting original details
	if arrivalTime == "" {
		waybill.ArrivalTime = append(waybill.ArrivalTime, arrivalTime)
	} else if leaveTime == "" {
		waybill.LeaveTime = append(waybill.LeaveTime, leaveTime)
	}
	waybill.Location = location
	waybill.StationTrainState = stationTrainState
	waybill.CheckDescription = checkDescription
	wayBillJSON, err = json.Marshal(waybill)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	err = ctx.GetStub().PutState(waybillIndexKey, wayBillJSON)
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
