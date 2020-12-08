alias lctrld='lctrld --config ~/source/work/LaunchControlD/config.yaml'
alias quicknew='lctrld events new ~/source/work/LaunchControlD/eventsample1.yml --provider virtualbox'
alias quicksetup='lctrld payload setup $EVTID'
alias quickdeploy='lctrld payload deploy $EVTID'
alias quickteardown='lctrld events teardown $EVTID && VBoxManage unregistervm $EVTID-0'
alias quickssh='docker-machine -s /tmp/workspace/evts/$EVTID/.docker/machine ssh $EVTID-0'

function reenv() {
    export EVTID=$1
    export EVTDIR=/tmp/workspace/evts/$EVTID
    export MACHINE_STORAGE_PATH=$EVTDIR/.docker/machine/
}

function testsendworks() {
    export EVT=$EVTDIR/event.json
    export ALICE=`jq -r '.accounts."alice@apeunit.com".address' $EVT`
    export FAUCET=`jq -r '.accounts.dropgiver.address' $EVT`
    launchpayloadcli tx send $FAUCET $ALICE 1drop --keyring-backend test --home /tmp/workspace/evts/$EVTID/nodeconfig/extra_accounts/dropgiver/ --node tcp://192.168.99.100:26657 --chain-id $EVTID
}

function faucetstatus() {
    export EVT=$EVTDIR/event.json
    export FAUCETIP=`jq -r '.state."alice@apeunit.com".Instance.IPAddress' $EVT`
    curl http://$FAUCETIP:8000/status
}

function faucetsend() {
    export EVT=$EVTDIR/event.json
    export ALICE=`jq -r '.accounts."alice@apeunit.com".address' $EVT`
    export FAUCET=`jq -r '.accounts.dropgiver.address' $EVT`
    export FAUCETIP=`jq -r '.state."alice@apeunit.com".Instance.IPAddress' $EVT`
    curl -v -X POST -d 'token=abadjoke' http://$FAUCETIP:8000/send/$ALICE/500drop
}
