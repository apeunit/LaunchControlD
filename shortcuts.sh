alias lctrld='lctrld --config ~/source/work/LaunchControlD/config.yaml'
alias quicknew='lctrld events new ~/source/work/LaunchControlD/eventsample1.yml --provider virtualbox'
alias quicksetup='lctrld payload setup $EVTID'
alias quickdeploy='lctrld payload deploy $EVTID'
alias quickteardown='lctrld events teardown $EVTID && VBoxManage unregistervm $EVTID-0'
alias quickssh='docker-machine -s /tmp/workspace/evts/$EVTID/.docker/machine ssh $EVTID-0'

function reenv() {
    export EVTID=$1
    export EVTDIR=/tmp/workspace/evts/$EVTID
    export EVT=$EVTDIR/event.json
    export EVTIP=`jq -r '.state."alice@apeunit.com".Instance.IPAddress' $EVT`
    export MACHINE_STORAGE_PATH=$EVTDIR/.docker/machine/
}

function testsendworks() {
    export EVT=$EVTDIR/event.json
    export ALICE=`jq -r '.accounts."alice@apeunit.com".address' $EVT`
    export FAUCET=`jq -r '.accounts.dropgiver.address' $EVT`
    launchpayloadcli tx send $FAUCET $ALICE 1drop --keyring-backend test --home /tmp/workspace/evts/$EVTID/nodeconfig/extra_accounts/dropgiver/ --node tcp://192.168.99.100:26657 --chain-id $EVTID
}

function testeverything {
    echo "FAUCET: GET http://$EVTIP:8000/status"
    curl http://$EVTIP:8000/status > /dev/null
    echo ""

    ALICE=`jq -r '.accounts."alice@apeunit.com".address' $EVT`
    FAUCET=`jq -r '.accounts.dropgiver.address' $EVT`
    echo "FAUCET: POST http://$EVTIP:8000/send/$ALICE/500drop"
    TXHASH=`curl -X POST -d 'token=abadjoke' http://$EVTIP:8000/send/$ALICE/1drops |jq -r '.txhash'`
    echo ""

    echo "GET $EVTIP:1317/auth/accounts/$ALICE"
    curl http://$EVTIP:1317/auth/accounts/$ALICE
    echo ""

    echo "GET $EVTIP:1317/bank/balances/$ALICE"
    curl http://$EVTIP:1317/bank/balances/$ALICE
    echo ""

    sleep 0.5

    echo "GET $EVTIP:1317/txs"
    curl http://$EVTIP:1317/txs
    echo ""

    echo "GET $EVTIP:1317/txs/$TXHASH"
    curl http://$EVTIP:1317/txs/$TXHASH
}
