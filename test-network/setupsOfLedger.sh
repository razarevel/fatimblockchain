./network.sh down
./network.sh up
./network.sh createChannel -c channel1
./network.sh createChannel -c channel2
./network.sh createChannel -c channel3
./network.sh createChannel -c channel4
./network.sh createChannel -c channel5
./network.sh createChannel -c channel6


./network.sh deployCC -ccn basic_channel1 -ccp ../asset-transfer-basic/chaincode-go-channel1 -ccl go -c channel1
./network.sh deployCC -ccn basic_channel2 -ccp ../asset-transfer-basic/chaincode-go-channel2 -ccl go -c channel2
./network.sh deployCC -ccn basic_channel3 -ccp ../asset-transfer-basic/chaincode-go-channel3 -ccl go -c channel3
./network.sh deployCC -ccn basic_channel4 -ccp ../asset-transfer-basic/chaincode-go-channel4 -ccl go -c channel4
./network.sh deployCC -ccn basic_channel5 -ccp ../asset-transfer-basic/chaincode-go-channel5 -ccl go -c channel5
./network.sh deployCC -ccn basic_channel6 -ccp ../asset-transfer-basic/chaincode-go-channel6 -ccl go -c channel6

clear

docker ps -a
