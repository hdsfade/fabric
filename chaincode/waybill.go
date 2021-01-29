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

type WayBillQueryResult struct {
	Code int     `json:"code"`
	Msg  string  `json:"msg"`
	Data WayBill `json:"data"`
}

//WayBillExists judges a waybill if exists or not.
func (s *SmartContract) WayBillExists(ctx contractapi.TransactionContextInterface, trainNumber string) (bool, error) {
	waybillIndexKey, err := ctx.GetStub().CreateCompositeKey(waybillIndexName, []string{trainNumber})
	if err != nil {
		return false, fmt.Errorf("failed to read from world state %v", err)
	}

	trainJSON, err := ctx.GetStub().GetState(waybillIndexKey)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state %v", err)
	}
	return trainJSON != nil, nil
}

//type WayBillExistData struct {
//	ISWayBillExist bool `json:"isWayBillExist"`
//}
//
//type WayBillExistResult struct {
//	wayBillExistData WayBillExistData `json:"data"`
//}

func (s *SmartContract) HasWayBill(ctx contractapi.TransactionContextInterface, trainNumber string) Result {
	result, _ := s.WayBillExists(ctx, trainNumber)
	if result {
		return Result{
			Code: 200,
			Msg:  fmt.Sprintf("the waybill %s exists", trainNumber),
		}
	} else {
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("the waybill %s does not exist", trainNumber),
		}
	}
}

////CreateWayBill issues a new line to the world state with given details.
func (s *SmartContract) CreateWayBill(ctx contractapi.TransactionContextInterface, trainNumber string) Result {
	createCargoResult := s.CreateCargo(ctx, trainNumber)
	if createCargoResult.Code != 200 {
		return createCargoResult
	}
	waybillIndexKey, err := ctx.GetStub().CreateCompositeKey(waybillIndexName, []string{trainNumber})
	if err != nil {
		s.DeleteCargo(ctx, trainNumber)
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}

	exists, err := s.WayBillExists(ctx, trainNumber)
	if err != nil {
		s.DeleteCargo(ctx, trainNumber)
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	if exists {
		s.DeleteCargo(ctx, trainNumber)
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("the waybill %s already exists", trainNumber),
		}
	}

	exists, err = s.TrainExists(ctx, trainNumber)
	if err != nil {
		s.DeleteCargo(ctx, trainNumber)
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	if !exists {
		s.DeleteCargo(ctx, trainNumber)
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("the train %s does not exist", trainNumber),
		}
	}

	scheduleNumber, err := strconv.Atoi(trainNumber[8:12])
	if err != nil {
		s.DeleteCargo(ctx, trainNumber)
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("trainNumber error: %v", err),
		}
	}
	scheduleIndexKey, err := ctx.GetStub().CreateCompositeKey(scheduleIndexName, []string{strconv.Itoa(scheduleNumber)})
	if err != nil {
		s.DeleteCargo(ctx, trainNumber)
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("failed to read schedule %d from world state: %v", scheduleNumber, err),
		}
	}
	scheduleJSON, err := ctx.GetStub().GetState(scheduleIndexKey)
	if err != nil {
		s.DeleteCargo(ctx, trainNumber)
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("failed to read schedule %d from world state: %v", scheduleNumber, err),
		}
	}
	if scheduleJSON == nil {
		s.DeleteCargo(ctx, trainNumber)
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("the schedule %d does not exist", scheduleNumber),
		}
	}
	var schedule Schedule
	err = json.Unmarshal(scheduleJSON, &schedule)
	if err != nil {
		s.DeleteCargo(ctx, trainNumber)
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}

	lineIndexKey, err := ctx.GetStub().CreateCompositeKey(lineIndexName, []string{strconv.Itoa(schedule.LineNumber)})
	if err != nil {
		s.DeleteCargo(ctx, trainNumber)
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("failed to read schedule %d's line %d from world state: %v", scheduleNumber, schedule.LineNumber, err),
		}
	}
	lineJSON, err := ctx.GetStub().GetState(lineIndexKey)
	if err != nil {
		s.DeleteCargo(ctx, trainNumber)
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("failed to read schedule %d's line %d from world state: %v", scheduleNumber, schedule.LineNumber, err),
		}
	}
	if lineJSON == nil {
		s.DeleteCargo(ctx, trainNumber)
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("the schedule %d's line %d does not exist", scheduleNumber, schedule.LineNumber),
		}
	}
	var line Line
	err = json.Unmarshal(lineJSON, &line)
	if err != nil {
		s.DeleteCargo(ctx, trainNumber)
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
		s.DeleteCargo(ctx, trainNumber)
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}

	err = ctx.GetStub().PutState(waybillIndexKey, wayBillJSON)
	if err != nil {
		s.DeleteCargo(ctx, trainNumber)
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

//QueryWayBillBytrainnumber returns the waybill in the world state with given trainnumber
func (s *SmartContract) QueryWayBillBytrainnumber(ctx contractapi.TransactionContextInterface, trainNumber string) WayBillQueryResult {
	waybillIndexKey, err := ctx.GetStub().CreateCompositeKey(waybillIndexName, []string{trainNumber})
	if err != nil {
		return WayBillQueryResult{
			Code: 402,
			Msg:  err.Error(),
			Data: WayBill{
				TrainNumber:       " ",
				WayStation:        []string{},
				ArrivalTime:       []string{},
				LeaveTime:         []string{},
				Location:          0,
				StationTrainState: false,
				CheckDescription:  " ",
			},
		}
	}

	waybillJSON, err := ctx.GetStub().GetState(waybillIndexKey)
	if err != nil {
		return WayBillQueryResult{
			Code: 402,
			Msg:  err.Error(),
			Data: WayBill{
				TrainNumber:       " ",
				WayStation:        []string{},
				ArrivalTime:       []string{},
				LeaveTime:         []string{},
				Location:          0,
				StationTrainState: false,
				CheckDescription:  " ",
			},
		}
	}
	if waybillJSON == nil {
		return WayBillQueryResult{
			Code: 402,
			Msg:  fmt.Sprintf("the waybill %s does not exist", trainNumber),
			Data: WayBill{
				TrainNumber:       " ",
				WayStation:        []string{},
				ArrivalTime:       []string{},
				LeaveTime:         []string{},
				Location:          0,
				StationTrainState: false,
				CheckDescription:  " ",
			},
		}
	}

	var waybill WayBill
	err = json.Unmarshal(waybillJSON, &waybill)
	if err != nil {
		return WayBillQueryResult{
			Code: 402,
			Msg:  err.Error(),
			Data: WayBill{
				TrainNumber:       " ",
				WayStation:        []string{},
				ArrivalTime:       []string{},
				LeaveTime:         []string{},
				Location:          0,
				StationTrainState: false,
				CheckDescription:  " ",
			},
		}
	}

	return WayBillQueryResult{
		Code: 200,
		Msg:  "success",
		Data: waybill,
	}
}
