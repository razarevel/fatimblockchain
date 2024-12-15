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
type IotLogs struct {
	Temperature string `json:"Temperature"`
	Pressure    string `json:"Pressure"`
	Location    string `json:"Location"`
	Quantity    string `json:"Quantity"`
	Quality     string `json:"Quality"`
}
type Bills struct {
	BillNumber     string `json:"Bill_Number"`
	TotalPayment   string `json:"Total_Payment"`
	CarrierName    string `json:"Carrier_Name"`
	CarrierAddress string `json:"Carrier_Address"`
	Date           string `json:"Date"`
}
type Env struct {
	Temperature string `json:"Temperature"`
	Pressure    string `json:"Pressure"`
}
type Asset struct {
	ID              string  `json:"ID"`
	Name            string  `json:"Name"`
	FacilityID      string  `json:"Facility_ID"`
	FacilityName    string  `json:"Facility_Name"`
	OilId           string  `json:"Oil_Batch_ID"`
	OilQualityCerti string  `json:"Oil_Quality_Certificate"`
	OilQuantity     string  `json:"Oil_Quantity"`
	Bill            Bills   `json:"Bill"`
	Compliance      Env     `json:"Compliance"`
	IotData         IotLogs `json:"Iot_Data"`
}

func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, name string, facilID string, facilName string, oilId string, oilQuanti string, oilQuali string, bilNumber string, totalPay string, carrName string, carrAdd string, date string, temp string, press string, iotTemp string, iotPres string, iotLoc string, iotQuan string, iotQual string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}
	bilss := Bills{
		BillNumber:     bilNumber,
		TotalPayment:   totalPay,
		CarrierName:    carrName,
		CarrierAddress: carrAdd,
		Date:           date,
	}
	iotLog := IotLogs{
		Temperature: iotTemp,
		Pressure:    iotPres,
		Location:    iotLoc,
		Quantity:    iotQuan,
		Quality:     iotQual,
	}
	compliance := Env{
		Temperature: temp,
		Pressure:    press,
	}
	asset := Asset{
		ID:              id,
		Name:            name,
		FacilityID:      facilID,
		FacilityName:    facilName,
		OilId:           oilId,
		OilQualityCerti: oilQuali,
		OilQuantity:     oilQuanti,
		Bill:            bilss,
		Compliance:      compliance,
		IotData:         iotLog,
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
