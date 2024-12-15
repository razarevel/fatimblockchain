package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
// Insert struct field in alphabetic order => to achieve determinism across languages
// golang keeps the order when marshal to json but doesn't order automatically
type Drilling struct {
	Name    string `json:"Name"`
	Payment string `json:"Payment"`
	Date    string `json:"Date"`
}
type Refineries struct {
	Name        string `json:"Name"`
	Payment     string `json:"Payment"`
	Date        string `json:"Date"`
	RealTimeSum string `json:"Reail_Time_Summary"`
}
type Storages struct {
	Name        string `json:"Name"`
	Payment     string `json:"Payment"`
	Date        string `json:"Date"`
	RealTimeSum string `json:"Reail_Time_Summary"`
}
type Consumers struct {
	Name        string `json:"Name"`
	Payment     string `json:"Payment"`
	Date        string `json:"Date"`
	RealTimeSum string `json:"Reail_Time_Summary"`
}
type IotLogs struct {
	Temperature string `json:"Temperature"`
	Pressure    string `json:"Pressure"`
	Location    string `json:"Location"`
	Quantity    string `json:"Quantity"`
	Quality     string `json:"Quality"`
}
type Asset struct {
	ID               string     `json:"ID"`
	Driller          Drilling   `json:"Driller"`
	Refinery         Refineries `json:"Refinery"`
	Storage          Storages   `json:"Storage"`
	Consumer         Consumers  `json:"Consumer"`
	ComplianceReport string     `json:"Compliance_Report"`
	Payment          string     `json:"Payment"`
	OilId            string     `json:"Oil_Batch_ID"`
	OilQualityCerti  string     `json:"Oil_Quality_Certificate"`
	OilQuantity      string     `json:"Oil_Quantity"`
	Time             string     `json:"Time_To_Complete"`
	DigitalSignature string     `json:"Digital_Signature"`
	IotData          []IotLogs  `json:"IotData"`
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, drillName string, drillPay string, drillDate string, refName string, refPay string, refDate string, refReal string, stName string, stPay string, stDate string, stReal string, conName string, conPay string, conDate string, conReal string, complia string, payment string, oilID string, oilQuali string, oilQuanti string, time string, digSign string, iotTemp string, iotPres string, iotLoc string, iotquanti string, iotquali string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}
	drill := Drilling{
		Name:    drillName,
		Payment: drillPay,
		Date:    drillDate,
	}
	refin := Refineries{
		Name:        refName,
		Payment:     refPay,
		Date:        refDate,
		RealTimeSum: refReal,
	}
	consu := Consumers{
		Name:        conName,
		Payment:     conPay,
		Date:        conDate,
		RealTimeSum: conReal,
	}
	iot := []IotLogs{
		{
			Temperature: iotTemp,
			Pressure:    iotPres,
			Location:    iotLoc,
			Quantity:    iotquanti,
			Quality:     iotquali,
		},
	}
	asset := Asset{
		ID:               id,
		Driller:          drill,
		Refinery:         refin,
		Consumer:         consu,
		ComplianceReport: complia,
		Payment:          payment,
		OilId:            oilID,
		OilQualityCerti:  oilQuali,
		OilQuantity:      oilQuanti,
		Time:             time,
		DigitalSignature: digSign,
		IotData:          iot,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// DeleteAsset deletes an given asset from the world state.
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

// updatre
func (s *SmartContract) UpdateIoTLogs(ctx contractapi.TransactionContextInterface, id string, temp string, press string, loca string, quanti string, quali string) (string, error) {
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return "", err
	}
	iotLogs := IotLogs{
		Temperature: temp,
		Pressure:    press,
		Location:    loca,
		Quantity:    quanti,
		Quality:     quali,
	}
	asset.IotData = append(asset.IotData, iotLogs)
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return "", fmt.Errorf("Just giveup coding. err: %s", err)
	}

	err = ctx.GetStub().PutState(id, assetJSON)
	if err != nil {
		return "", fmt.Errorf("You're terrible at this. err: %s", err)
	}
	return "It's all good", nil
}
