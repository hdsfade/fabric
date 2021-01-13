//@author: hdsfade
//@date: 2021-01-13-21:45
package chaincode

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

//Result structure used for handing result of create or delete
type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

//vehicle, station, line compositekey prefix
var vehicleIndexName = "vehicle"
var stationIndexName = "station"
var lineIndexName = "line"
var stationlineIndexName = "station~line"

// Init  ledger(can add a default set of assets to the ledger)
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	//Init vehicles
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
			return err
		}
		vehicleIndexKey, err := ctx.GetStub().CreateCompositeKey(vehicleIndexName, []string{string(vehicle.VehicleNumber)})
		if err != nil {
			return err
		}
		err = ctx.GetStub().PutState(vehicleIndexKey, vehicleJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	//Init stations
	stations := []Station{
		{StationName: "上海", Country: "中国", Using: true, Describtion: ""},
		{StationName: "北京", Country: "中国", Using: true, Describtion: ""},
		{StationName: "广州", Country: "中国", Using: true, Describtion: ""},
		{StationName: "杭州", Country: "中国", Using: true, Describtion: ""},
		{StationName: "宁波", Country: "中国", Using: true, Describtion: ""},
		{StationName: "南京", Country: "中国", Using: true, Describtion: ""},
		{StationName: "嘉兴", Country: "中国", Using: true, Describtion: ""},
	}
	for _, station := range stations {
		stationJSON, err := json.Marshal(station)
		if err != nil {
			return err
		}

		vehicleIndexKey, err := ctx.GetStub().CreateCompositeKey(stationIndexName, []string{station.StationName})
		err = ctx.GetStub().PutState(vehicleIndexKey, stationJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	//Init lines, create compositekey line~station
	lines := []Line{
		{LineNumber: 1, WayStation: []string{"宁波", "杭州", "南京"}, WayStationType: []string{"始发站", "途径站", "终点站"}, Using: true},
		{LineNumber: 2, WayStation: []string{"宁波", "杭州", "上海"}, WayStationType: []string{"始发站", "途径站", "终点站"}, Using: true},
		{LineNumber: 3, WayStation: []string{"宁波", "嘉兴", "上海"}, WayStationType: []string{"始发站", "途径站", "终点站"}, Using: true},
	}
	for _, line := range lines {
		lineJSON, err := json.Marshal(line)
		if err != nil {
			return err
		}

		lineIndexKey, err := ctx.GetStub().CreateCompositeKey(lineIndexName, []string{string(line.LineNumber)})
		err = ctx.GetStub().PutState(lineIndexKey, lineJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}

		value := []byte{0x00}
		for _, stationName := range line.WayStation {
			stationLineIndexKey, err := ctx.GetStub().CreateCompositeKey(stationlineIndexName, []string{stationName, string(line.LineNumber)})
			if err != nil {
				return err
			}
			err = ctx.GetStub().PutState(stationLineIndexKey, value)
			if err != nil {
				return fmt.Errorf("failed to put to world state. %v", err)
			}
		}
	}

	return nil
}