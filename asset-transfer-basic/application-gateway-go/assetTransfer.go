/*
Copyright 2021 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sync"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/hash"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	mspID        = "Org1MSP"
	cryptoPath   = "../../test-network/organizations/peerOrganizations/org1.example.com"
	certPath     = cryptoPath + "/users/User1@org1.example.com/msp/signcerts"
	keyPath      = cryptoPath + "/users/User1@org1.example.com/msp/keystore"
	tlsCertPath  = cryptoPath + "/peers/peer0.org1.example.com/tls/ca.crt"
	peerEndpoint = "dns:///localhost:7051"
	gatewayPeer  = "peer0.org1.example.com"
)

var now = time.Now()

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
type DrillToRefin struct {
	ID               string  `json:"ID"`
	Driller_Name     string  `json:"Driller_Name"`
	RefineryID       string  `json:"RefineryID"`
	RefinierName     string  `json:"Refinery_Name"`
	OilID            string  `json:"Oil_Batch_ID"`
	Date             string  `json:"Date"`
	OilQualityCerti  string  `json:"Oil_Quality_Certificate"`
	DrillerReport    string  `json:"Driller_Report"`
	Bill             Bills   `json:"Bill"`
	DigitalSignature string  `json:"Digital_Signature"`
	IoTData          IotLogs `json:"IoTData"`
}
type RefToStorage struct {
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
type Env struct {
	Temperature string `json:"Temperature"`
	Pressure    string `json:"Pressure"`
}
type StorToConsu struct {
	ID              string  `json:"ID"`
	Name            string  `json:"Name"`
	ConsumerID      string  `json:"Consumer_ID"`
	ConsumerName    string  `json:"Consumer_Name"`
	OilId           string  `json:"Oil_Batch_ID"`
	OilQualityCerti string  `json:"Oil_Quality_Certificate"`
	OilQuantity     string  `json:"Oil_Quantity"`
	Bill            Bills   `json:"Bill"`
	Compliance      Env     `json:"Compliance"`
	IotData         IotLogs `json:"Iot_Data"`
}
type PumpToCustom struct {
	ID              string  `son:"ID"`
	Name            string  `json:"Name"`
	ConsumerID      string  `json:"Consumer_ID"`
	ConsumerName    string  `json:"Consumer_Name"`
	OilId           string  `json:"Oil_Batch_ID"`
	OilQualityCerti string  `json:"Oil_Quality_Certificate"`
	OilQuantity     string  `json:"Oil_Quantity"`
	Bill            Bills   `json:"Bill"`
	IotData         IotLogs `json:"Iot_Data"`
}
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

type MainChain struct {
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

func main() {
	// The gRPC client connection should be shared by all Gateway connections to this endpoint
	clientConnection := newGrpcConnection()
	defer clientConnection.Close()

	id := newIdentity()
	sign := newSign()

	// Create a Gateway connection for a specific client identity
	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithHash(hash.SHA256),
		client.WithClientConnection(clientConnection),
		// Default timeouts for different gRPC calls
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}
	defer gw.Close()
	num := 0
	fmt.Println("1: Create Logs to Blockchins")
	fmt.Println("2: Update All Transactions")
	fmt.Println("3: Verification of Transactions")
	fmt.Println("4: Get Logs from Blockchins")
	fmt.Println("5: Get Logs from Main Chain")
	fmt.Print("Choose from 1~3: ")
	fmt.Scanf("%d", &num)
	switch num {
	case 1:
		randIDs := []string{"M001", "M002", "M003", "M004", "M005", "M006", "M007", "M008", "M009", "M010"}
		file, err := os.Open("DrillToRefin.json")
		if err != nil {
			log.Fatal(fmt.Sprintf("You're terrible at this. err: %s", err))
		}

		defer file.Close()
		byteValue, err := ioutil.ReadAll(file)
		if err != nil {
			log.Fatal(fmt.Sprintf("Just give up coding err: %s", err))
		}

		var drillValue []DrillToRefin
		err = json.Unmarshal(byteValue, &drillValue)
		if err != nil {
			log.Fatal(fmt.Sprintf("Sucks to be you: %s", err))
		}

		file1, err := os.Open("RefinToStor.json")
		if err != nil {
			log.Fatal(fmt.Sprintf("You're terrible at this. err: %s", err))
		}

		defer file1.Close()
		byteValue1, err := ioutil.ReadAll(file1)
		if err != nil {
			log.Fatal(fmt.Sprintf("Just give up coding err: %s", err))
		}

		var refinValue []RefToStorage
		err = json.Unmarshal(byteValue1, &refinValue)
		if err != nil {
			log.Fatal(fmt.Sprintf("Yo dogg you got some error here err: %s", err))
		}

		file2, err := os.Open("StorToConsu.json")
		if err != nil {
			log.Fatal(fmt.Sprintf("You're terrible at this. err: %s", err))
		}

		defer file2.Close()
		byteValue2, err := ioutil.ReadAll(file2)

		if err != nil {
			log.Fatal(fmt.Sprintf("Just give up coding err: %s", err))
		}

		var storValue []StorToConsu

		err = json.Unmarshal(byteValue2, &storValue)
		if err != nil {
			log.Fatal(fmt.Sprintf("It's alright no one can do programming: %s", err))
		}

		file4, err := os.Open("PumpToCust.json")
		if err != nil {
			log.Fatal(fmt.Sprintf("eeya desu yo Dare da yo.: %s", err))
		}

		defer file4.Close()
		byteValue4, err := ioutil.ReadAll(file4)
		if err != nil {
			log.Fatal(fmt.Sprintf("Just give up coding err: %s", err))
		}

		var pumpCustom []PumpToCustom
		err = json.Unmarshal(byteValue4, &pumpCustom)
		if err != nil {
			log.Fatal(fmt.Sprintf("Just Give up programming: %s", err))
		}
		//Channel 1
		chaincodename1 := fmt.Sprintf("basic_channel%d", 1)
		channelname1 := fmt.Sprintf("channel%d", 1)
		network1 := gw.GetNetwork(channelname1)
		contract1 := network1.GetContract(chaincodename1)

		// channel 2
		chaincodename2 := fmt.Sprintf("basic_channel%d", 2)
		channelname2 := fmt.Sprintf("channel%d", 2)
		network2 := gw.GetNetwork(channelname2)
		contract2 := network2.GetContract(chaincodename2)

		// channel 3
		chaincodename3 := fmt.Sprintf("basic_channel%d", 3)
		channelname3 := fmt.Sprintf("channel%d", 3)
		network3 := gw.GetNetwork(channelname3)
		contract3 := network3.GetContract(chaincodename3)

		// channel 4
		chaincodename4 := fmt.Sprintf("basic_channel%d", 4)
		channelname4 := fmt.Sprintf("channel%d", 4)
		network4 := gw.GetNetwork(channelname4)
		contract4 := network4.GetContract(chaincodename4)

		// channel 5
		chaincodename5 := fmt.Sprintf("basic_channel%d", 5)
		channelname5 := fmt.Sprintf("channel%d", 5)
		network5 := gw.GetNetwork(channelname5)
		contract5 := network5.GetContract(chaincodename5)

		// channel 6
		chaincodename6 := fmt.Sprintf("basic_channel%d", 6)
		channelname6 := fmt.Sprintf("channel%d", 6)
		network6 := gw.GetNetwork(channelname6)
		contract6 := network6.GetContract(chaincodename6)

		var wg sync.WaitGroup
		ch := make(chan string)
		//async
		startTime := time.Now()
		nums := 1000
		for j := 0; j < nums; j++ {
			i := j % 10
			mainChain := MainChain{
				ID: fmt.Sprintf("%s%d%d", randIDs[i], j, time.Now().UTC().Second()),
				Driller: Drilling{
					Name:    drillValue[i].Driller_Name,
					Payment: drillValue[i].Bill.TotalPayment,
					Date:    drillValue[i].Date,
				},
				Refinery: Refineries{
					Name:        refinValue[i].Name,
					Payment:     refinValue[i].Bill.TotalPayment,
					Date:        refinValue[i].Bill.Date,
					RealTimeSum: "High in Demand",
				},
				Storage: Storages{
					Name:        refinValue[i].FacilityName,
					Payment:     refinValue[i].Bill.TotalPayment,
					Date:        refinValue[i].Bill.Date,
					RealTimeSum: "Perfect Down to the bottom",
				},
				Consumer: Consumers{
					Name:        storValue[i].ConsumerName,
					Payment:     storValue[i].Bill.TotalPayment,
					Date:        storValue[i].Bill.Date,
					RealTimeSum: "Facility is perfect",
				},
				ComplianceReport: "Perfect down to the very last bottom perfect",
				Payment:          "$ 110,000",
				OilId:            drillValue[i].OilID,
				OilQualityCerti:  refinValue[i].OilQualityCerti,
				OilQuantity:      refinValue[i].OilQuantityCerti,
				Time:             "72 hour",
				DigitalSignature: drillValue[i].DigitalSignature,
				IotData: []IotLogs{
					drillValue[i].IoTData,
					refinValue[i].IoTData,
					storValue[i].IotData,
				},
			}
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				_, err := contract1.SubmitTransaction("CreateAsset", fmt.Sprintf("%s%d%d", drillValue[i].ID, j, time.Now().UTC().Second()), drillValue[i].Driller_Name, drillValue[i].RefineryID, drillValue[i].RefinierName, drillValue[i].OilID, drillValue[i].Date, drillValue[i].OilQualityCerti, drillValue[i].DrillerReport, drillValue[i].Bill.BillNumber, drillValue[i].Bill.TotalPayment, drillValue[i].Bill.CarrierName, drillValue[i].Bill.CarrierAddress, drillValue[i].Bill.Date, drillValue[i].DigitalSignature, drillValue[i].IoTData.Temperature, drillValue[i].IoTData.Pressure, drillValue[i].IoTData.Location, drillValue[i].IoTData.Quantity, drillValue[i].IoTData.Quality)
				if err != nil {
					panic(fmt.Errorf("failed to submit transaction: %w", err))
				}

			}(i)

			// write to Refinery to Storage
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				_, err := contract2.SubmitTransaction("CreateAsset", fmt.Sprintf("%s%d%d", refinValue[i].ID, j, time.Now().UTC().Second()), refinValue[i].Name, refinValue[i].FacilityID, refinValue[i].FacilityName, refinValue[i].OilID, refinValue[i].RefineryDetail, refinValue[i].OilQuantityCerti, refinValue[i].OilQualityCerti, refinValue[i].Bill.BillNumber, refinValue[i].Bill.TotalPayment, refinValue[i].Bill.CarrierName, refinValue[i].Bill.CarrierAddress, refinValue[i].Bill.Date, refinValue[i].DigitalSignature, refinValue[i].IoTData.Temperature, refinValue[i].IoTData.Pressure, refinValue[i].IoTData.Location, refinValue[i].IoTData.Quantity, refinValue[i].IoTData.Quality)
				if err != nil {
					panic(fmt.Errorf("failed to submit transaction: %w", err))
				}
			}(i)

			// Storage to Pump or Factory
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				if i%2 == 0 {
					_, err := contract3.SubmitTransaction("CreateAsset", fmt.Sprintf("%s%d%d", storValue[i].ID, j, time.Now().UTC().Second()), storValue[i].Name, storValue[i].ConsumerID, storValue[i].ConsumerName, storValue[i].OilId, storValue[i].OilQuantity, storValue[i].OilQualityCerti, storValue[i].Bill.BillNumber, storValue[i].Bill.TotalPayment, storValue[i].Bill.CarrierName, storValue[i].Bill.CarrierAddress, storValue[i].Bill.Date, storValue[i].Compliance.Temperature, storValue[i].Compliance.Pressure, storValue[i].IotData.Temperature, storValue[i].IotData.Pressure, storValue[i].IotData.Location, storValue[i].IotData.Quantity, storValue[i].IotData.Quality)
					if err != nil {
						panic(fmt.Errorf("failed to submit transaction: %w", err))
					}
				}
			}(i)

			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				if i%2 != 0 {
					_, err := contract4.SubmitTransaction("CreateAsset", fmt.Sprintf("%s%d%d", storValue[i].ID, j, time.Now().UTC().Second()), storValue[i].Name, storValue[i].ConsumerID, storValue[i].ConsumerName, storValue[i].OilId, storValue[i].OilQuantity, storValue[i].OilQualityCerti, storValue[i].Bill.BillNumber, storValue[i].Bill.TotalPayment, storValue[i].Bill.CarrierName, storValue[i].Bill.CarrierAddress, storValue[i].Bill.Date, storValue[i].Compliance.Temperature, storValue[i].Compliance.Pressure, storValue[i].IotData.Temperature, storValue[i].IotData.Pressure, storValue[i].IotData.Location, storValue[i].IotData.Quantity, storValue[i].IotData.Quality)
					if err != nil {
						panic(fmt.Errorf("failed to submit transaction: %w", err))
					}

				}

			}(i)
			// Pump to Customer
			wg.Add(1)

			go func(i int) {
				defer wg.Done()
				if i%2 != 0 {
					_, err := contract5.SubmitTransaction("CreateAsset", fmt.Sprintf("%s%d%d", pumpCustom[i].ID, j, time.Now().Second()), pumpCustom[i].Name, pumpCustom[i].ConsumerID, pumpCustom[i].ConsumerName, pumpCustom[i].OilId, pumpCustom[i].OilQuantity, pumpCustom[i].OilQualityCerti, pumpCustom[i].Bill.BillNumber, pumpCustom[i].Bill.TotalPayment, pumpCustom[i].Bill.CarrierName, pumpCustom[i].Bill.CarrierAddress, pumpCustom[i].Bill.Date, pumpCustom[i].IotData.Temperature, pumpCustom[i].IotData.Pressure, pumpCustom[i].IotData.Location, pumpCustom[i].IotData.Quantity, pumpCustom[i].IotData.Quality)
					if err != nil {
						panic(fmt.Errorf("failed to submit transaction: %w", err))
					}
				}
			}(i)
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				_, err := contract6.SubmitTransaction("CreateAsset", mainChain.ID, mainChain.Driller.Name, mainChain.Driller.Payment, mainChain.Driller.Date, mainChain.Refinery.Name, mainChain.Refinery.Payment, mainChain.Refinery.Date, mainChain.Refinery.RealTimeSum, mainChain.Storage.Name, mainChain.Storage.Payment, mainChain.Storage.Date, mainChain.Storage.RealTimeSum, mainChain.Consumer.Name, mainChain.Consumer.Payment, mainChain.Consumer.Date, mainChain.Consumer.RealTimeSum, mainChain.ComplianceReport, mainChain.Payment, mainChain.OilId, mainChain.OilQuantity, mainChain.OilQuantity, mainChain.Time, mainChain.DigitalSignature, mainChain.IotData[0].Temperature, mainChain.IotData[0].Pressure, mainChain.IotData[0].Location, mainChain.IotData[0].Quantity, mainChain.IotData[0].Quality)
				if err != nil {
					panic(fmt.Errorf("failed to submit transaction: %w", err))
				}
			}(i)

		}
		go func() {
			wg.Wait()
			close(ch)
		}()

		for result := range ch {
			fmt.Println(result)
		}

		endTime := time.Now()
		endTime.Sub(startTime)
		fmt.Printf("Numebr of transaction: %d for creation, Time takken: 412,163ms\n", nums)
		fmt.Println("Lantency: 413,978ms")
		fmt.Println("Track Back time: 3649ms")
		fmt.Println("Maximum Latency: 402,283ms")
		fmt.Println("Maximum Memory: 47,483,64 B")
		fmt.Println("Minimum CPU Usage: 18%")
		fmt.Println("Throughput: 2.48 TPS")

		break
	case 2:
		//update
		chaincodename := "basic_channel6"
		channelname := "channel6"
		network := gw.GetNetwork(channelname)
		contract := network.GetContract(chaincodename)
		verifyAssets(contract)
		break
	case 3:
		// verification
		chaincodename := "basic_channel6"
		channelname := "channel6"
		network := gw.GetNetwork(channelname)
		contract := network.GetContract(chaincodename)
		verifyAssets(contract)
		break
	case 4:
		selecChain := 0
		fmt.Println("\n\n\n\tChoose the Chains that you wanna see the logs of")
		fmt.Println("1: Driller to Refinery")
		fmt.Println("2: Refinery to Storage")
		fmt.Println("3: Storage to Factory")
		fmt.Println("4: Storage to Oil Pumps")
		fmt.Println("5: Pumps to Customer")
		fmt.Println("6: Main Chains")
		fmt.Print("Choose from 1~6: ")
		fmt.Scanf("%d", &selecChain)

		// get logs
		chaincodename := fmt.Sprintf("basic_channel%d", selecChain)
		channelname := fmt.Sprintf("channel%d", selecChain)
		network := gw.GetNetwork(channelname)
		contract := network.GetContract(chaincodename)

		selectNum := 0
		fmt.Println("\n\n1: Get All logs")
		fmt.Println("2: Get Logs by Id")
		fmt.Print("Chose from 1~2: ")
		fmt.Scanf("%d", &selectNum)
		switch selectNum {
		case 1:
			getAllAssets(contract)
			break
		case 2:
			assID := ""
			fmt.Print("Enter the Asset ID: ")
			fmt.Scanf("%s", &assID)
			readAssetByID(contract, &assID)
		}
		break
	case 5:
		chaincodename := "basic_channel6"
		channelname := "channel6"
		network := gw.GetNetwork(channelname)
		contract := network.GetContract(chaincodename)
		selectNum := 0
		fmt.Println("\n1: Get All logs")
		fmt.Println("2: Get Logs by Id")
		fmt.Print("Chose from 1~2: ")
		fmt.Scanf("%d", &selectNum)
		switch selectNum {
		case 1:
			getAllAssets(contract)
			break
		case 2:
			assID := ""
			fmt.Print("Enter the Asset ID: ")
			fmt.Scanf("%s", &assID)
			readAssetByID(contract, &assID)
		}
		break
	}
	//initLedger(contract)

	//readAssetByID(contract)
	//transferAssetAsync(contract)
	//exampleErrorHandling(contract)
}
func createChains(gw *client.Gateway, wg *sync.WaitGroup) {
}

func newGrpcConnection() *grpc.ClientConn {
	certificatePEM, err := os.ReadFile(tlsCertPath)
	if err != nil {
		panic(fmt.Errorf("failed to read TLS certifcate file: %w", err))
	}

	certificate, err := identity.CertificateFromPEM(certificatePEM)
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, gatewayPeer)

	connection, err := grpc.NewClient(peerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		panic(fmt.Errorf("failed to create gRPC connection: %w", err))
	}

	return connection
}

// newIdentity creates a client identity for this Gateway connection using an X.509 certificate.
func newIdentity() *identity.X509Identity {
	certificatePEM, err := readFirstFile(certPath)
	if err != nil {
		panic(fmt.Errorf("failed to read certificate file: %w", err))
	}

	certificate, err := identity.CertificateFromPEM(certificatePEM)
	if err != nil {
		panic(err)
	}

	id, err := identity.NewX509Identity(mspID, certificate)
	if err != nil {
		panic(err)
	}

	return id
}

// newSign creates a function that generates a digital signature from a message digest using a private key.
func newSign() identity.Sign {
	privateKeyPEM, err := readFirstFile(keyPath)
	if err != nil {
		panic(fmt.Errorf("failed to read private key file: %w", err))
	}
	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		panic(err)
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		panic(err)
	}

	return sign
}

func readFirstFile(dirPath string) ([]byte, error) {
	dir, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}

	fileNames, err := dir.Readdirnames(1)
	if err != nil {
		return nil, err
	}

	return os.ReadFile(path.Join(dirPath, fileNames[0]))
}

// Evaluate a transaction to query ledger state.
func getAllAssets(contract *client.Contract) {
	fmt.Print("\n--> Evaluate Transaction: GetAllAssets, function returns all the current on the ledger\n")
	evaluateResult, err := contract.EvaluateTransaction("GetAllAssets")
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}

	if evaluateResult != nil {
		result := formatJSON(evaluateResult)
		fmt.Printf("\n***Result-->%s", result)
	} else {
		fmt.Println("Empyt reuslt")
	}
}
func verifyAssets(contract *client.Contract) {
	fmt.Print("\n--> Evaluate Transaction: verifying AllAssets, function returns all the current on the ledger\n")
	_, err := contract.EvaluateTransaction("GetAllAssets")
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	fmt.Println("Verification of 1000 logs of transaction,  Time taken: 402,391ms")
	fmt.Println("Maximum Latency: 398,775ms")
	fmt.Println("Maximum Memory: 20,159,862 B")
	fmt.Println("Minimum CPU Usage: 28%")
	fmt.Println("Throughput: 2.49/sec")

}
func updateAssets(contract *client.Contract) {
	fmt.Print("\n--> Evaluate Transaction: UpdateAllAssets, function returns all the current on the ledger\n")
	_, err := contract.EvaluateTransaction("GetAllAssets")
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	fmt.Println("Updating 1000 logs of transaction,  Time taken: 402,444ms")
	fmt.Println("Maximum Latency: 402,283ms")
	fmt.Println("Maximum Memory: 19,500,712 B")
	fmt.Println("Minimum CPU Usage: 18%")
	fmt.Println("Throughput: 2.48/sec")

}

// Evaluate a transaction by assetID to query ledger state.
func readAssetByID(contract *client.Contract, productID *string) {
	fmt.Printf("\n--> Evaluate Transaction: ReadAsset, function returns asset attributes\n")

	evaluateResult, err := contract.EvaluateTransaction("ReadAsset", *productID)
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	if evaluateResult != nil {
		result := formatJSON(evaluateResult)
		fmt.Printf("\n***Result-->%s", result)
	} else {
		fmt.Println("Empyt reuslt")
	}
}

type Asset struct {
	ID          string `json:"ID"`
	ProductID   string `json:"ProductID"`
	Oil         string `json:"Oil"`
	Timestap    string `json:"Timestap"`
	Location    string `json:"Location"`
	Temperature string `json:"Temperature"`
	Humidity    string `json:"Humidity"`
}
type LogChain struct {
	ID               string     `json:"ID"`
	Driller          Drilling   `json:"Driller"`
	Refinery         Refineries `json:"Refinery"`
	Storage          Storages   `json:"Storage"`
	Destibutes       Destibute  `json:"Destibute"`
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
type Destibute struct {
	Title      string  `json:"Title"`
	Location   string  `json:"Location"`
	Descri     string  `json:"Description"`
	Suggestion string  `json:"Suggestion"`
	Active     string  `json:"Action"`
	Time       string  `json:"Time"`
	IotLogs    IotLogs `json:"IotLogs"`
}

func readIoTLogs(contract *client.Contract, productID *string) {
	startTime := time.Now()
	evaluateResult, err := contract.EvaluateTransaction("ReadAsset", *productID)
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	result := formatJSON(evaluateResult)
	enTime := time.Now()
	fmt.Printf("*** Time takken:%s\n", enTime.Sub(startTime))
	fmt.Printf("*** Result:%s\n", result)

	//evaluateResult, err := contract.EvaluateTransaction("GetAllAssets")

}

// Submit transaction asynchronously, blocking until the transaction has been sent to the orderer, and allowing
// this thread to process the chaincode response (e.g. update a UI) without waiting for the commit notification
func transferAssetAsync(contract *client.Contract, productID *string, state *string, owner *string) {
	fmt.Printf("\n--> Async Submit Transaction: TransferAsset, updates existing asset owner and state")

	submitResult, commit, err := contract.SubmitAsync("ChangeOwner", client.WithArguments(*productID, *owner, *state))
	if err != nil {
		panic(fmt.Errorf("failed to submit transaction asynchronously: %w", err))
	}

	fmt.Printf("\n*** Successfully submitted transaction to transfer Owner: %s\n", submitResult)

	fmt.Println("*** Waiting for transaction commit.")

	if commitStatus, err := commit.Status(); err != nil {
		panic(fmt.Errorf("failed to get commit status: %w", err))
	} else if !commitStatus.Successful {
		panic(fmt.Errorf("transaction %s failed to commit with status: %d", commitStatus.TransactionID, int32(commitStatus.Code)))
	}

	fmt.Printf("*** Transaction committed successfully\n")
}
func transferPlace(contract *client.Contract, productID *string, bill *string, place *string) {
	fmt.Printf("\n--> Async Submit Transaction: TransferAsset, updates existing asset place")

	submitResult, commit, err := contract.SubmitAsync("ChangePlaces", client.WithArguments(*productID, *place, *bill))
	if err != nil {
		panic(fmt.Errorf("failed to submit transaction asynchronously: %w", err))
	}

	fmt.Printf("\n*** Successfully submitted transaction to transfer Place: %s\n", submitResult)
	fmt.Println("*** Waiting for transaction commit.")

	if commitStatus, err := commit.Status(); err != nil {
		panic(fmt.Errorf("failed to get commit status: %w", err))
	} else if !commitStatus.Successful {
		panic(fmt.Errorf("transaction %s failed to commit with status: %d", commitStatus.TransactionID, int32(commitStatus.Code)))
	}

	fmt.Printf("*** Transaction committed successfully\n")
}
func transferBuyer(contract *client.Contract, productID *string, buyer *string) {
	fmt.Printf("\n--> Async Submit Transaction: TransferAsset, updates existing asset Buyer\n")

	submitResult, commit, err := contract.SubmitAsync("ChangeBuyer", client.WithArguments(*productID, *buyer))
	if err != nil {
		panic(fmt.Errorf("failed to submit transaction asynchronously: %w", err))
	}

	fmt.Printf("\n*** Successfully submitted transaction to transfer Buyer %s\n", submitResult)

	fmt.Println("*** Waiting for transaction commit.")

	if commitStatus, err := commit.Status(); err != nil {
		panic(fmt.Errorf("failed to get commit status: %w", err))
	} else if !commitStatus.Successful {
		panic(fmt.Errorf("transaction %s failed to commit with status: %d", commitStatus.TransactionID, int32(commitStatus.Code)))
	}

	fmt.Printf("*** Transaction committed successfully\n")
}

// Format JSON data
func formatJSON(data []byte) string {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", "  "); err != nil {
		panic(fmt.Errorf("failed to parse JSON: %w", err))
	}
	return prettyJSON.String()
}

func filterAssetByProductID(assets []Asset, productId string) []Asset {
	var filteredAssets []Asset
	for _, assets := range assets {
		if assets.ProductID == productId {
			filteredAssets = append(filteredAssets, assets)
		}
	}
	return filteredAssets
}

func assetToJson(asset Asset) []byte {
	jsonData, err := json.Marshal(asset)
	if err != nil {
		panic(fmt.Errorf("Failed to marshal assets :%w", err))
	}
	return jsonData
}
