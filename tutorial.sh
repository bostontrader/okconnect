BSERVER="http://185.183.96.73:3003"
APIKEY="$(curl -X POST $BSERVER/apikeys | jq -r .apikey)"
CURRENCY_BTC="$(curl -d "apikey=$APIKEY&rarity=0&symbol=BTC&title=Bitcoin" $BSERVER/currencies | jq .data.last_insert_id)"
CURRENCY_BSV="$(curl -d "apikey=$APIKEY&rarity=0&symbol=BSV&title=Bitcoin SV" $BSERVER/currencies | jq .data.last_insert_id)"

ACCT_EQUITY="$(curl -d "apikey=$APIKEY&rarity=0&currency_id=$CURRENCY_BTC&title=Owners Equity" $BSERVER/accounts | jq .data.last_insert_id)"
ACCT_LCL_WALLET="$(curl -d "apikey=$APIKEY&rarity=0&currency_id=$CURRENCY_BTC&title=Local Wallet" $BSERVER/accounts | jq .data.last_insert_id)"
ACCT_FUNDING="$(curl -d "apikey=$APIKEY&rarity=0&currency_id=$CURRENCY_BTC&title=OKEx Funding" $BSERVER/accounts | jq .data.last_insert_id)"
ACCT_FEE="$(curl -d "apikey=$APIKEY&rarity=0&currency_id=$CURRENCY_BTC&title=Fee" $BSERVER/accounts | jq .data.last_insert_id)"
ACCT_SPOT_BTC="$(curl -d "apikey=$APIKEY&rarity=0&currency_id=$CURRENCY_BTC&title=OKEx Spot" $BSERVER/accounts | jq .data.last_insert_id)"
ACCT_SPOT_BSV="$(curl -d "apikey=$APIKEY&rarity=0&currency_id=$CURRENCY_BSV&title=OKEx Spot" $BSERVER/accounts | jq .data.last_insert_id)"
ACCT_SPOT_HOLD="$(curl -d "apikey=$APIKEY&rarity=0&currency_id=$CURRENCY_BTC&title=OKEx Spot-Hold" $BSERVER/accounts | jq .data.last_insert_id)"

CAT_ASSETS="$(curl -d "apikey=$APIKEY&symbol=A&title=Assets" $BSERVER/categories | jq .data.last_insert_id)"
CAT_LIABILITIES="$(curl -d "apikey=$APIKEY&symbol=L&title=Liabilities" $BSERVER/categories | jq .data.last_insert_id)"
CAT_EQUITY="$(curl -d "apikey=$APIKEY&symbol=Eq&title=Equity" $BSERVER/categories | jq .data.last_insert_id)"
CAT_REVENUE="$(curl -d "apikey=$APIKEY&symbol=R&title=Revenue" $BSERVER/categories | jq .data.last_insert_id)"
CAT_EXPENSES="$(curl -d "apikey=$APIKEY&symbol=Ex&title=Expenses" $BSERVER/categories | jq .data.last_insert_id)"

curl -d "apikey=$APIKEY&account_id=$ACCT_LCL_WALLET&category_id=$CAT_ASSETS" $BSERVER/acctcats
curl -d "apikey=$APIKEY&account_id=$ACCT_FUNDING&category_id=$CAT_ASSETS" $BSERVER/acctcats
curl -d "apikey=$APIKEY&account_id=$ACCT_SPOT_BTC&category_id=$CAT_ASSETS" $BSERVER/acctcats
curl -d "apikey=$APIKEY&account_id=$ACCT_SPOT_BSV&category_id=$CAT_ASSETS" $BSERVER/acctcats
curl -d "apikey=$APIKEY&account_id=$ACCT_SPOT_HOLD&category_id=$CAT_ASSETS" $BSERVER/acctcats
curl -d "apikey=$APIKEY&account_id=$ACCT_FEE&category_id=$CAT_EXPENSES" $BSERVER/acctcats
curl -d "apikey=$APIKEY&account_id=$ACCT_EQUITY&category_id=$CAT_EQUITY" $BSERVER/acctcats

TXID1="$(curl -d "apikey=$APIKEY&notes=Initial Equity&time=2020-05-01T12:34:55.000Z" $BSERVER/transactions | jq .data.last_insert_id)"
curl -d "&account_id=$ACCT_LCL_WALLET&apikey=$APIKEY&amount=2&amount_exp=0&transaction_id=$TXID1" $BSERVER/distributions
curl -d "&account_id=$ACCT_EQUITY&apikey=$APIKEY&amount=-2&amount_exp=0&transaction_id=$TXID1" $BSERVER/distributions

OKEXURL="http://185.183.96.73:8090"
OKEX_CREDENTIALS="okcatbox.json"
curl -X POST $OKEXURL/credentials --output $OKEX_CREDENTIALS

echo "bookwerxconfig:" >> okconnect.yaml
echo "  apikey: $APIKEY" >> okconnect.yaml
echo "  server: $BSERVER" >> okconnect.yaml
echo "okexconfig:" >> okconnect.yaml
echo "  credentials: $OKEX_CREDENTIALS" >> okconnect.yaml
echo "  server: $OKEXURL" >> okconnect.yaml

echo "compareconfig:" >> okconnect.yaml
echo "  funding:" >> okconnect.yaml
echo "    - currencyid: btc" >> okconnect.yaml
echo "      available: $ACCT_FUNDING" >> okconnect.yaml
echo "  spot:" >> okconnect.yaml
echo "    - currencyid: btc" >> okconnect.yaml
echo "      available: $ACCT_SPOT_BTC" >> okconnect.yaml
echo "    - currencyid: bsv" >> okconnect.yaml
echo "      hold: $ACCT_SPOT_BSV" >> okconnect.yaml
echo "      available: $ACCT_SPOT_HOLD" >> okconnect.yaml

./okconnect -cmd compare -config okconnect.yaml