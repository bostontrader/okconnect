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

CAT_EQUITY="$(curl -d "apikey=$APIKEY&symbol=Eq&title=Equity" $BSERVER/categories | jq .data.last_insert_id)"
CAT_EXPENSE="$(curl -d "apikey=$APIKEY&symbol=Ex&title=Expenses" $BSERVER/categories | jq .data.last_insert_id)"
CAT_ASSET="$(curl -d "apikey=$APIKEY&symbol=A&title=Assets" $BSERVER/categories | jq .data.last_insert_id)"

curl -d "apikey=$APIKEY&account_id=$ACCT_LCL_WALLET&category_id=$CAT_ASSET" $BSERVER/acctcats
curl -d "apikey=$APIKEY&account_id=$ACCT_FUNDING&category_id=$CAT_ASSET" $BSERVER/acctcats
curl -d "apikey=$APIKEY&account_id=$ACCT_SPOT_BTC&category_id=$CAT_ASSET" $BSERVER/acctcats
curl -d "apikey=$APIKEY&account_id=$ACCT_SPOT_BSV&category_id=$CAT_ASSET" $BSERVER/acctcats
curl -d "apikey=$APIKEY&account_id=$ACCT_SPOT_HOLD&category_id=$CAT_ASSET" $BSERVER/acctcats
curl -d "apikey=$APIKEY&account_id=$ACCT_FEE&category_id=$CAT_EXPENSE" $BSERVER/acctcats
curl -d "apikey=$APIKEY&account_id=$ACCT_EQUITY&category_id=$CAT_EQUITY" $BSERVER/acctcats