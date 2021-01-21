//@author: hdsfade
//@date: 2021-01-17-14:25
package chaincode

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"strconv"
)

var scheduleIndexName = "schedule"
var linescheduleIndexName = "line~schedule"
var vehiclescheduleIndexName = "vehicle~schedule"

//Schedule describes details of a schedule
type Schedule struct {
	ScheduleNumber int  `json:"scheduleNumber"`
	LineNumber     int  `json:"lineNumber"`
	VehicleNumber  int  `json:"vehicleNumber"`
	UnitPrice      int  `json:"unitPrice"`
	Using          bool `json:"using"`
}

type Schedules struct {
	ScheduleData []Schedule `json:"schedules"`
}

//ScheduleQueryResult structure used for handing result of query
type ScheduleQueryResult struct {
	Code int      `json:"code"`
	Msg  string   `json:"msg"`
	Data Schedule `json:"data"`
}

//ScheduleQueryResults structure used for handing result of queryAll
type ScheduleQueryResults struct {
	Code int       `json:"code"`
	Msg  string    `json:"msg"`
	Data Schedules `json:"data"`
}

//ScheduleExists judges a schedule if exists or not
func (s *SmartContract) ScheduleExists(ctx contractapi.TransactionContextInterface, scheduleNumber int) (bool, error) {
	scheduleIndexkey, err := ctx.GetStub().CreateCompositeKey(scheduleIndexName, []string{strconv.Itoa(scheduleNumber)})
	if err != nil {
		return false, fmt.Errorf("failed to read from world state %v", err)
	}

	scheduleJSON, err := ctx.GetStub().GetState(scheduleIndexkey)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state %v", err)
	}
	return scheduleJSON != nil, nil
}

//CreateSchedule issues a new schedule to the world state with given details
func (s *SmartContract) CreateSchedule(ctx contractapi.TransactionContextInterface, scheduleNumber, lineNumber, vehicleNumber, unitPrice int) Result {
	scheduleIndexKey, err := ctx.GetStub().CreateCompositeKey(scheduleIndexName, []string{strconv.Itoa(scheduleNumber)})
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}

	exists, err := s.ScheduleExists(ctx, scheduleNumber)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	if exists {
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("the schedule %d already exists", scheduleNumber),
		}
	}

	//if the line lineNumber does not exist, the schedule couldn't be created
	exists, err = s.LineExists(ctx, lineNumber)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	if exists == false {
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("the line %d does not exist", lineNumber),
		}
	}

	//if the vehicle vehicleNumber does not exist, the schedule couldn't be created
	exists, err = s.LineExists(ctx, vehicleNumber)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	if exists == false {
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("the vehicle %d does not exist", vehicleNumber),
		}
	}

	schedule := Schedule{
		ScheduleNumber: scheduleNumber,
		LineNumber:     lineNumber,
		VehicleNumber:  vehicleNumber,
		UnitPrice:      unitPrice,
		Using:          true,
	}
	scheduleJSON, err := json.Marshal(schedule)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}

	err = ctx.GetStub().PutState(scheduleIndexKey, scheduleJSON)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}

	value := []byte{0x00}
	//create compositekey schedule~line
	linescheduleIndexKey, err := ctx.GetStub().CreateCompositeKey(
		linescheduleIndexName, []string{strconv.Itoa(lineNumber), strconv.Itoa(scheduleNumber)})
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	err = ctx.GetStub().PutState(linescheduleIndexKey, value)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}

	//create compositekey schedule-vehicle
	vehiclescheduleIndexKey, err := ctx.GetStub().CreateCompositeKey(
		vehiclescheduleIndexName, []string{strconv.Itoa(vehicleNumber), strconv.Itoa(scheduleNumber)})
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	err = ctx.GetStub().PutState(vehiclescheduleIndexKey, value)
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

//DeleteSchedule deletes a line by lineNumber from the world state.
func (s *SmartContract) DeleteSchedule(ctx contractapi.TransactionContextInterface, scheduleNumber int) Result {
	scheduleIndexKey, err := ctx.GetStub().CreateCompositeKey(scheduleIndexName, []string{strconv.Itoa(scheduleNumber)})
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}

	exists, err := s.ScheduleExists(ctx, scheduleNumber)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	if !exists {
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("the schedule %d does not exist", scheduleNumber),
		}
	}

	scheduleJSON, err := ctx.GetStub().GetState(strconv.Itoa(scheduleNumber))
	var schedule Schedule
	err = json.Unmarshal(scheduleJSON, &schedule)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}

	//delete compositekey schedule~line
	linescheduleIndexKey, err := ctx.GetStub().CreateCompositeKey(
		linescheduleIndexName, []string{strconv.Itoa(schedule.LineNumber), strconv.Itoa(scheduleNumber)})
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	err = ctx.GetStub().DelState(linescheduleIndexKey)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}

	//delete compositekey schedule~vehicle
	vehiclescheduleIndexKey, err := ctx.GetStub().CreateCompositeKey(
		vehiclescheduleIndexName, []string{strconv.Itoa(schedule.VehicleNumber), strconv.Itoa(scheduleNumber)})
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	err = ctx.GetStub().DelState(vehiclescheduleIndexKey)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}

	err = ctx.GetStub().DelState(scheduleIndexKey)
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

//QueryScheduleByschedulenumber returns the schedule in the world state with given scheduleNumber
func (s *SmartContract) QueryScheduleByschedulenumber(ctx contractapi.TransactionContextInterface, scheduleNumber int) ScheduleQueryResult {
	scheduleIndexKey, err := ctx.GetStub().CreateCompositeKey(scheduleIndexName, []string{strconv.Itoa(scheduleNumber)})
	if err != nil {
		return ScheduleQueryResult{
			Code: 402,
			Msg:  fmt.Sprintf("failed to read from world state: %v", err),
			Data: Schedule{
				ScheduleNumber: 0,
				LineNumber:     0,
				VehicleNumber:  0,
				UnitPrice:      0,
				Using:          false,
			},
		}
	}

	scheduleJSON, err := ctx.GetStub().GetState(scheduleIndexKey)
	if err != nil {
		return ScheduleQueryResult{
			Code: 402,
			Msg:  fmt.Sprintf("failed to read from world state: %v", err),
			Data: Schedule{
				ScheduleNumber: 0,
				LineNumber:     0,
				VehicleNumber:  0,
				UnitPrice:      0,
				Using:          false,
			},
		}
	}
	if scheduleJSON == nil {
		return ScheduleQueryResult{
			Code: 402,
			Msg:  fmt.Sprintf("the scheduel %d does not exist", scheduleNumber),
			Data: Schedule{
				ScheduleNumber: 0,
				LineNumber:     0,
				VehicleNumber:  0,
				UnitPrice:      0,
				Using:          false,
			},
		}
	}

	var schedule Schedule
	err = json.Unmarshal(scheduleJSON, &schedule)
	if err != nil {
		return ScheduleQueryResult{
			Code: 402,
			Msg:  fmt.Sprintf("the scheduel %d does not exist", scheduleNumber),
			Data: Schedule{
				ScheduleNumber: 0,
				LineNumber:     0,
				VehicleNumber:  0,
				UnitPrice:      0,
				Using:          false,
			},
		}
	}
	return ScheduleQueryResult{
		Code: 200,
		Msg:  "success",
		Data: schedule,
	}
}

//QueryAllSchedules returns all schedules found in world state
func (s *SmartContract) QueryAllSchedules(ctx contractapi.TransactionContextInterface) ScheduleQueryResults {
	var emptyschedules []Schedule
	emptyschedules = append(emptyschedules, Schedule{
		ScheduleNumber: 0,
		LineNumber:     0,
		VehicleNumber:  0,
		UnitPrice:      0,
		Using:          false,
	})

	scheduleResultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(scheduleIndexName, []string{})
	if err != nil {
		return ScheduleQueryResults{
			Code: 402,
			Msg:  err.Error(),
			Data: Schedules{ScheduleData: emptyschedules},
		}
	}
	defer scheduleResultsIterator.Close()

	var schedules []Schedule
	for scheduleResultsIterator.HasNext() {
		scheduleQueryResponse, err := scheduleResultsIterator.Next()
		if err != nil {
			return ScheduleQueryResults{
				Code: 402,
				Msg:  err.Error(),
				Data: Schedules{ScheduleData: emptyschedules},
			}
		}

		var schedule Schedule
		err = json.Unmarshal(scheduleQueryResponse.Value, &schedule)
		if err != nil {
			return ScheduleQueryResults{
				Code: 402,
				Msg:  err.Error(),
				Data: Schedules{ScheduleData: emptyschedules},
			}
		}
		schedules = append(schedules, schedule)
		if schedules == nil {
			return ScheduleQueryResults{
				Code: 402,
				Msg:  "No schedule",
				Data: Schedules{ScheduleData: emptyschedules},
			}
		}
	}

	return ScheduleQueryResults{
		Code: 200,
		Msg:  "success",
		Data: Schedules{ScheduleData: schedules},
	}
}
