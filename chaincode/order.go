//@author: hdsfade
//@date: 2021-01-17-15:42
package chaincode

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"strconv"
	"time"
)

var orderIndexName = "order"
var trainorderIndexName = "train~order"

//Order describes details of a order
type Order struct { //订单
	OrderId            int      `json:"orderId"`
	GenerateTime       string   `json:"generateTime"`
	CustomerId         int      `json:"customerId"`
	TrainNumber        string   `json:"trainNumber"`
	StartingStation    string   `json:"startingStation"`
	DestinationStation string   `json:"destinationStation"`
	CarriageNumber     int      `json:"carriageNumber"`
	Price              int      `json:"price"`        //订单金额
	TotalTypeNum       int      `json:"totalTypeNum"` //订单中也要货物信息
	CargoType          []string `json:"cargoType"`
	GoodsNum           []int    `json:"goodsNum"`
	GoodsName          []string `json:"goodsName"`
	CheckResult        bool     `json:"checkResult"`
	CheckDescription   string   `json:"checkDescription"`
}

type Orders struct {
	OrdersData []Order `json:"orders"`
}

//OrderQueryResult structure used for handing result of query
type OrderQueryResult struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data Order  `json:"data"`
}

//OrderQueryResults structure used for handing result of queryAll
type OrderQueryResults struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data Orders `json:"data"`
}

//OrderExists judges a order if exists or not
func (s *SmartContract) OrderExists(ctx contractapi.TransactionContextInterface, orderId int) (bool, error) {
	orderIndexKey, err := ctx.GetStub().CreateCompositeKey(orderIndexName, []string{strconv.Itoa(orderId)})
	if err != nil {
		return false, fmt.Errorf("failed to read from world state %v", err)
	}

	orderJSON, err := ctx.GetStub().GetState(orderIndexKey)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state %v", err)
	}
	return orderJSON != nil, nil
}

//CreateOrder issues a new order to the world state with given details.
func (s *SmartContract) CreateOrder(ctx contractapi.TransactionContextInterface, orderId, customerId int, trainNumber string,
	startingStation, destinationStation string, carriageNumber, price, totalTypeNum int, cargoType []string, goodsNumber []int,
	goodsName []string) Result {
	orderIndexKey, err := ctx.GetStub().CreateCompositeKey(orderIndexName, []string{strconv.Itoa(orderId)})
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}

	exists, err := s.OrderExists(ctx, orderId)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	if exists {
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("the order %d already exists", orderId),
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

	updateTrainResult := s.UpdateTrain(ctx, trainNumber, carriageNumber)
	if updateTrainResult.Code != 200 {
		return updateTrainResult
	}

	order := Order{
		OrderId:            orderId,
		GenerateTime:       time.Now().Format("2006-01-02"),
		CustomerId:         customerId,
		TrainNumber:        trainNumber,
		StartingStation:    startingStation,
		DestinationStation: destinationStation,
		CarriageNumber:     carriageNumber,
		Price:              price,
		TotalTypeNum:       totalTypeNum,
		CargoType:          cargoType,
		GoodsNum:           goodsNumber,
		GoodsName:          goodsName,
		CheckResult:        false,
		CheckDescription:   " ",
	}
	orderJSON, err := json.Marshal(order)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}

	err = ctx.GetStub().PutState(orderIndexKey, orderJSON)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}

	//create compositekey train~order
	trainorderIndexKey, err := ctx.GetStub().CreateCompositeKey(
		trainorderIndexName, []string{trainNumber, strconv.Itoa(orderId)})
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	err = ctx.GetStub().PutState(trainorderIndexKey, []byte(strconv.Itoa(orderId)))
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

//DeleteOrder deletes a order by orderId from the world state
func (s *SmartContract) DeleteOrder(ctx contractapi.TransactionContextInterface, orderId int) Result {
	orderIndexKey, err := ctx.GetStub().CreateCompositeKey(orderIndexName, []string{strconv.Itoa(orderId)})
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}

	exists, err := s.OrderExists(ctx, orderId)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	if !exists {
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("the order %d does not exist", orderId),
		}
	}

	err = ctx.GetStub().DelState(orderIndexKey)
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

//UpdateOrder updates an existing order in the world state with provided parameters
func (s *SmartContract) UpdateOrder(ctx contractapi.TransactionContextInterface, orderId int, checkRsult bool, checkDescription string) Result {
	orderIndexkey, err := ctx.GetStub().CreateCompositeKey(orderIndexName, []string{strconv.Itoa(orderId)})
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	orderJSON, err := ctx.GetStub().GetState(orderIndexkey)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	if orderJSON == nil {
		return Result{
			Code: 402,
			Msg:  fmt.Sprintf("the order %d does not exist", orderId),
		}
	}

	var order Order
	err = json.Unmarshal(orderJSON, &order)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	//overwriting original checkResult and checkDescription
	order.CheckResult = checkRsult
	order.CheckDescription = checkDescription
	orderJSON, err = json.Marshal(order)
	if err != nil {
		return Result{
			Code: 402,
			Msg:  err.Error(),
		}
	}
	err = ctx.GetStub().PutState(orderIndexkey, orderJSON)
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

// CheckOrder updates order's checkResult and checkDescription
func (s *SmartContract) CheckOrder(ctx contractapi.TransactionContextInterface, orderId int, checkResult bool, checkDescription string) Result {
	return s.UpdateOrder(ctx, orderId, checkResult, checkDescription)
}

//QueryOrderByorderid returns the order in the world state with given orderId
func (s *SmartContract) QueryOrderByorderid(ctx contractapi.TransactionContextInterface, orderId int) OrderQueryResult {
	orderIndexKey, err := ctx.GetStub().CreateCompositeKey(orderIndexName, []string{strconv.Itoa(orderId)})
	if err != nil {
		return OrderQueryResult{
			Code: 402,
			Msg:  err.Error(),
			Data: Order{
				OrderId:            0,
				GenerateTime:       "",
				CustomerId:         0,
				TrainNumber:        "0",
				StartingStation:    "",
				DestinationStation: "",
				CarriageNumber:     0,
				Price:              0,
				TotalTypeNum:       0,
				CargoType:          []string{},
				GoodsNum:           []int{},
				GoodsName:          []string{},
				CheckResult:        false,
				CheckDescription:   "",
			},
		}
	}

	orderJSON, err := ctx.GetStub().GetState(orderIndexKey)
	if err != nil {
		return OrderQueryResult{
			Code: 402,
			Msg:  fmt.Sprintf("failed to read from world state: %v", err),
			Data: Order{
				OrderId:            0,
				GenerateTime:       "",
				CustomerId:         0,
				TrainNumber:        "0",
				StartingStation:    "",
				DestinationStation: "",
				CarriageNumber:     0,
				Price:              0,
				TotalTypeNum:       0,
				CargoType:          []string{},
				GoodsNum:           []int{},
				GoodsName:          []string{},
				CheckResult:        false,
				CheckDescription:   "",
			},
		}
	}
	if orderJSON == nil {
		return OrderQueryResult{
			Code: 402,
			Msg:  fmt.Sprintf("the order %d does not exist", orderId),
			Data: Order{
				OrderId:            0,
				GenerateTime:       "",
				CustomerId:         0,
				TrainNumber:        "0",
				StartingStation:    "",
				DestinationStation: "",
				CarriageNumber:     0,
				Price:              0,
				TotalTypeNum:       0,
				CargoType:          []string{},
				GoodsNum:           []int{},
				GoodsName:          []string{},
				CheckResult:        false,
				CheckDescription:   "",
			},
		}
	}

	var order Order
	err = json.Unmarshal(orderJSON, &order)
	if err != nil {
		return OrderQueryResult{
			Code: 402,
			Msg:  err.Error(),
			Data: Order{
				OrderId:            0,
				GenerateTime:       "",
				CustomerId:         0,
				TrainNumber:        "0",
				StartingStation:    "",
				DestinationStation: "",
				CarriageNumber:     0,
				Price:              0,
				TotalTypeNum:       0,
				CargoType:          []string{},
				GoodsNum:           []int{},
				GoodsName:          []string{},
				CheckResult:        false,
				CheckDescription:   "",
			},
		}
	}

	return OrderQueryResult{
		Code: 200,
		Msg:  "success",
		Data: order,
	}
}

//QueryAllOrders returns all orders found in world state
func (s *SmartContract) QueryAllOrders(ctx contractapi.TransactionContextInterface) OrderQueryResults {
	var emptyorders []Order
	emptyorders = append(emptyorders, Order{
		OrderId:            0,
		GenerateTime:       " ",
		CustomerId:         0,
		TrainNumber:        "0",
		StartingStation:    " ",
		DestinationStation: " ",
		CarriageNumber:     0,
		Price:              0,
		TotalTypeNum:       0,
		CargoType:          []string{},
		GoodsNum:           []int{},
		GoodsName:          []string{},
		CheckResult:        false,
		CheckDescription:   " ",
	})

	orderResultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(orderIndexName, []string{})
	if err != nil {
		return OrderQueryResults{
			Code: 402,
			Msg:  err.Error(),
			Data: Orders{OrdersData: emptyorders},
		}
	}
	defer orderResultsIterator.Close()

	var orders []Order
	for orderResultsIterator.HasNext() {
		orderQueryResponse, err := orderResultsIterator.Next()
		if err != nil {
			return OrderQueryResults{
				Code: 402,
				Msg:  err.Error(),
				Data: Orders{OrdersData: emptyorders},
			}
		}

		var order Order
		err = json.Unmarshal(orderQueryResponse.Value, &order)
		if err != nil {
			return OrderQueryResults{
				Code: 402,
				Msg:  err.Error(),
				Data: Orders{OrdersData: emptyorders},
			}
		}
		orders = append(orders, order)
	}
	if orders == nil {
		return OrderQueryResults{
			Code: 402,
			Msg:  "No order",
			Data: Orders{OrdersData: emptyorders},
		}
	}

	return OrderQueryResults{
		Code: 200,
		Msg:  "success",
		Data: Orders{OrdersData: orders},
	}
}
