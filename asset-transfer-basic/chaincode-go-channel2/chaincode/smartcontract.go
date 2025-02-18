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
type Bills struct {
	BillNumber     string `json:"Bill_Number"`
	TotalPayment   string `json:"Total_Payment"`
	CarrierName    string `json:"Carrier_Name"`
	CarrierAddress string `json:"Carrier_Address"`
	Date           string `json:"Date"`
}
type IotLogs struct {
	Temperature string `json:"Temperature"`
	Pressure    string `json:"Pressure"`
	Location    string `json:"Location"`
	Quantity    string `json:"Quantity"`
	Quality     string `json:"Quality"`
}
type Asset struct {
	ID               string  `json:"ID"`
	Name             string  `json:"Name"`
	FacilityID       string  `json:"Facility_ID"`
	FacilityName     string  `json:"Facility_Name"`
	OilID            string  `json:"Oil_Batch_ID"`
	RefineryDetail   string  `json:"Refinery_Detail"`
	OilQuantityCerti string  `json:"Oil_Quantity_Certificate"`
	OilQualityCerti  string  `json:"Oil_Quality_Certificate"`
	Bill             Bills   `json:"Bill"`
	DigitalSignature string  `json:"Digital_Signature"`
	IoTData          IotLogs `json:"Iot_Data"`
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, name string, facilityID string, facilityName string, oilId string, refineryDetail string, quantiCerti string, ouanliCerti string, billNumber string, totalPay string, carrName string, carrAdd string, date string, digitalSign string, iotTemp string, iotPres string, iotLoc string, iotQuan string, iotQual string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}
	bilss := Bills{
		BillNumber:     billNumber,
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
	asset := Asset{
		ID:               id,
		Name:             name,
		FacilityID:       facilityID,
		FacilityName:     facilityName,
		OilID:            oilId,
		RefineryDetail:   refineryDetail,
		OilQuantityCerti: quantiCerti,
		OilQualityCerti:  ouanliCerti,
		Bill:             bilss,
		DigitalSignature: digitalSign,
		IoTData:          iotLog,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, productID string) ([]*Asset, error) {
	selector := `{"selector": {"ProductID": ` + productID + `}}`
	resultsIterator, err := ctx.GetStub().GetQueryResult(selector)
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
	if len(assets) == 0 {
		return nil, fmt.Errorf("no assets found for ProductID: %s", productID)
	}
	return assets, nil
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

func (s *SmartContract) ChangeIotData(ctx contractapi.TransactionContextInterface, id string, iotTemp string, iotPres string, iotLocat string, iotQuanti string, iotQuali string) (string, error) {
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return "", err
	}
	iotLog := IotLogs{
		Temperature: iotTemp,
		Pressure:    iotPres,
		Location:    iotLocat,
		Quantity:    iotQuanti,
		Quality:     iotQuali,
	}
	asset[0].IoTData = iotLog
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
