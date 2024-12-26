cd /home/sample/fabric1/scripts/fabric-samples/test-network
./network.sh down
./network.sh up createChannel -c mychannel -ca
./network.sh deployCC -ccn basic -ccp ../asset-transfer-basic/chaincode-go/ -ccl go
if [ $0 == "1" ]; then
    bash startproject.sh
fi