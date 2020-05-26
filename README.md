[![Build Status](https://travis-ci.org/bostontrader/okconnect.svg?branch=master)](https://travis-ci.org/bostontrader/okconnect)
[![MIT license](http://img.shields.io/badge/license-MIT-brightgreen.svg)](http://opensource.org/licenses/MIT)

# Welcome to OKConnect

OKEx provides an API for using their service.  That's fine and dandy but in order to make effective use of this API, you will be greatly convenienced by acquiring a handful of additional tools.  More specifically:

1. [Bookwerx](https://github.com/bostontrader/bookwerx-core-rust). You will need some method of bookkeeping. The only reason anybody cares about an API is so that they can work the service using other software.  What do you do with the OKEx API?  That's right... you place orders and buy and sell things. Placing, fulfilling, and cancelling orders spawn a remarkably tedious snake nest of bookkeeping tasks.  Ignoring the bookkeeping is the method of chumps.   The bookkeeping and account balance information provided by the OKEx API is, ahem, less than divinely inspired.  It's a disorganized jumble permeated with documentation errors, round-off errors, and various blind-spots. Dealing with the bookkeeping manually will easily drive you mad. Unless of course you have bookwerx in the arena with you.

2. [OKCatbox](https://github.com/bostontrader/okcatbox). You might also want an OKEx API sandbox to play in.  Learning how to use the real OKEx API looks suspiciously close to DOS and general hackery from their point of view.  Perhaps it's better to taunt a sandbox first, before trying to use the real OKEx API.

OKConnect is the glue that binds the OKEx (or OKCatbox) API and Bookwerx together. With OKConnect, you can focus on higher-level tasks such as placing and cancelling orders, dealing with the consequences of order fulfillment, as well as reconciliation of the Bookwerx and OKEx records,  while letting the OKEx (or OKCatbox) API and Bookwerx figure out the bookkeeping.

## Getting Started

OKConnect is a command-line tool.  It takes a handful of runtime args, one of which is a command that specifies OKConnect's specific operation.

For example,
```
./okconnect
```
or
```
./okconnect -cmd help
```
Would both produce usage information.

In order for OKConnect to work it's going to need:

* Access to the OKEx API or a functioning mimic such as [OKCatbox](https://github.com/bostontrader/okcatbox).

* Access to a [Bookwerx Core](https://github.com/bostontrader/bookwerx-core-rust) server.

* A YML configuration file that ties these two things together.  

Unfortunately, finding, installing, and configuring all this is rather tedious.  Fortunately we have the following tutorial to walk you through an example.

In this example, we're going to figure out how to persuade OKConnect to place an order to sell 1 BTC and buy 25 BSV at the implied price of BSVBTC = 0.04. We will use a public demonstration of OKCatbox to do this.  Along the way we will also use OKConnect to compare the bookkeeping records in Bookwerx with the account balances on OKCatbox.


1. Setup Bookwerx:

A. Our first order of business is to get the Bookwerx bookkeeping going.  Although it is moderately elaborate to install from scratch, we present a [public demonstration version of the Bookwerx UI](http://185.183.96.73:3005/).  You can follow this tutorial manually using the Bookwerx UI. As an extra bonus, the Bookwerx UI will show you the actual http requests that it uses to do its job.

B. Using the Bookwerx UI go to the BServer tab, and enter http://185.183.96.73:3003.  The UI that you are using communicates with a public demonstration Bookwerx Core server via a RESTful API. Be sure to test the connection in order to proceed.

We're going to use the URL of the Bookwerx Core server in subsequent requests, so let's save it as a shell variable:
```
BSERVER="http://185.183.96.73:3003"
```

Also notice that this URL is not https.  That's ok for a public demonstration tutorial but if you were using real data you would probably want to use https.
  
C. Using the Bookwerx UI go to the newly visible API key tab and either request a new API key or use any key you have created in the past.  So for example, if you had earlier started this tutorial with a given API key, enter that key now to continue.  

D. Notice that the Bookwerx UI shows the http request that it will use in order to get a new key.  You could submit this request manually, using curl:
```
curl -X POST $BSERVER/apikeys
```

The response looks intuitively obvious to our human eyeballs.  However, we're going to need to use this value repeatedly in the future, so it would be really useful to parse this response, pick out just the value of the apikey, and save just that value as a shell variable.  We can easily parse this using [jq](https://stedolan.github.io/jq/). Assuming jq is properly installed and combining all this new-found learning into one command yields:

```
APIKEY="$(curl -X POST $BSERVER/apikeys | jq -r .apikey)"
```

E. Since we are going to use BTC and BSV in our subsequent transactions, we must first define them as currencies in Bookwerx. We can do this manually using the Bookwerx UI by going to the Currencies tab.

As with the APIkey, the Bookwerx UI shows us the http request that it will submit to the Bookwerx Core server to create these new currency records.  If the request succeeds then we will receive the ID of the newly created currency.  We want to be able to use this ID subsequently, so we again want to parse the response using our new friend jq and save the value as a shell variable.

```
CURRENCY_BTC="$(curl -d "apikey=$APIKEY&rarity=0&symbol=BTC&title=Bitcoin" $BSERVER/currencies | jq .data.last_insert_id)"
CURRENCY_BSV="$(curl -d "apikey=$APIKEY&rarity=0&symbol=BSV&title=Bitcoin SV" $BSERVER/currencies | jq .data.last_insert_id)"
```
Upon close inspection you can see a parameter named "rarity".  It's harmless but not relevant for this tutorial so just ignore it.

Another wrinkle is that we have double quotes inside double quotes.  Oddly enough this seems to work for us, but this looks like something that might go wrong for somebody else, so be wary of this.


F. Next, let's anticipate some accounts that we might need. We can do this manually using the Bookwerx UI by going to the Accounts tab.   Realize that each account has a description and a currency.  Two accounts can have the same description as long as they have different currencies.

|Description    |Currency  |
|---------------|----------|
|Owner's Equity	|BTC       | 
|Local Wallet   |BTC       |
|OKEx Funding	|BTC       |
|Fee            |BSV       |
|OKEx Spot	    |BTC       |
|OKEx Spot-Hold |BTC       |
|OKEx Spot	    |BSV       |

As with the currencies you can see a parameter named "rarity" and it's still harmless and irrelevant for this tutorial so just ignore it.

In order to save the IDs of the newly created accounts:
```
ACCT_EQUITY="$(curl -d "apikey=$APIKEY&rarity=0&currency_id=$CURRENCY_BTC&title=Owners Equity" $BSERVER/accounts | jq .data.last_insert_id)"
ACCT_LCL_WALLET="$(curl -d "apikey=$APIKEY&rarity=0&currency_id=$CURRENCY_BTC&title=Local Wallet" $BSERVER/accounts | jq .data.last_insert_id)"
ACCT_FUNDING="$(curl -d "apikey=$APIKEY&rarity=0&currency_id=$CURRENCY_BTC&title=OKEx Funding" $BSERVER/accounts | jq .data.last_insert_id)"
ACCT_FEE="$(curl -d "apikey=$APIKEY&rarity=0&currency_id=$CURRENCY_BTC&title=Fee" $BSERVER/accounts | jq .data.last_insert_id)"
ACCT_SPOT_BTC="$(curl -d "apikey=$APIKEY&rarity=0&currency_id=$CURRENCY_BTC&title=OKEx Spot" $BSERVER/accounts | jq .data.last_insert_id)"
ACCT_SPOT_BSV="$(curl -d "apikey=$APIKEY&rarity=0&currency_id=$CURRENCY_BSV&title=OKEx Spot" $BSERVER/accounts | jq .data.last_insert_id)"
ACCT_SPOT_HOLD="$(curl -d "apikey=$APIKEY&rarity=0&currency_id=$CURRENCY_BTC&title=OKEx Spot-Hold" $BSERVER/accounts | jq .data.last_insert_id)"
```
G. Next, in order to produce balance sheet and income statement style reports, we'll need to define some useful categories that we can use to tag the accounts: We can do this manually using the Bookwerx UI by going to the Categories tab.

|Symbol |Title       |
|-------|------------|
|A	    |Assets      | 
|L      |Liabilities |
|Eq	    |Equity      |
|R      |Revenue     |
|Ex	    |Expenses    |

Please notice that we have defined categories for Liabilities and Revenue, even though we don't yet have any accounts of these types.  Oddly enough we'll need the categories soon, before we need specific accounts of those types.


In order to save the IDs of the newly created categories:
```
CAT_ASSETS="$(curl -d "apikey=$APIKEY&symbol=Eq&title=Assets" $BSERVER/categories | jq .data.last_insert_id)"
CAT_LIABILITIES="$(curl -d "apikey=$APIKEY&symbol=Eq&title=Liabilities" $BSERVER/categories | jq .data.last_insert_id)"
CAT_EQUITY="$(curl -d "apikey=$APIKEY&symbol=Eq&title=Equity" $BSERVER/categories | jq .data.last_insert_id)"
CAT_REVENUE="$(curl -d "apikey=$APIKEY&symbol=Eq&title=Revenue" $BSERVER/categories | jq .data.last_insert_id)"
CAT_EXPENSES="$(curl -d "apikey=$APIKEY&symbol=Eq&title=Expenses" $BSERVER/categories | jq .data.last_insert_id)"
```

H. Finally, let's tag the accounts with suitable categories. We can do this manually using the Bookwerx UI by going to the Categories tab and then click on the "Accounts" button for a particular category and then select the accounts that should be tagged with the given category.

| Category | Account       | Currency|
|----------|---------------|---------|
|A	       |OKEx Funding   | BTC     |
|A	       |OKEx Spot	   | BTC     |
|A	       |OKEx Spot-Hold | BTC     |
|A	       |OKEx Spot	   | BSV     |
|A	       |Local Wallet   | BTC     |
|Ex	       |Fee	           | BTC     |
|Eq	       |Owner's Equity | BTC     |

In this case, even though we still make http requests to do this, we don't care about saving any information from the responses.

```
curl -d "apikey=$APIKEY&account_id=$ACCT_LCL_WALLET&category_id=$CAT_ASSET" $BSERVER/acctcats
curl -d "apikey=$APIKEY&account_id=$ACCT_FUNDING&category_id=$CAT_ASSET" $BSERVER/acctcats
curl -d "apikey=$APIKEY&account_id=$ACCT_SPOT_BTC&category_id=$CAT_ASSET" $BSERVER/acctcats
curl -d "apikey=$APIKEY&account_id=$ACCT_SPOT_BSV&category_id=$CAT_ASSET" $BSERVER/acctcats
curl -d "apikey=$APIKEY&account_id=$ACCT_SPOT_HOLD&category_id=$CAT_ASSET" $BSERVER/acctcats
curl -d "apikey=$APIKEY&account_id=$ACCT_FEE&category_id=$CAT_EXPENSE" $BSERVER/acctcats
curl -d "apikey=$APIKEY&account_id=$ACCT_EQUITY&category_id=$CAT_EQUITY" $BSERVER/acctcats
```
I. With these accounts let's make an initial transaction to contribute some BTC to our books:

2020-05-01T12:34:55.000Z
Initial Equity

DR Local Wallet   BTC 2.0
CR Owner's Equity BTC 2.0

We can do this manually using the Bookwerx UI and going to the Transactions tab.  First we create the transaction, then we edit the transaction to add the two distributions (the dr and cr bits).

```
TXID1="$(curl -d "apikey=$APIKEY&notes=Initial Equity&time=2020-05-01T12:34:55.000Z" $BSERVER/transactions | jq .data.last_insert_id)"
curl -d "&account_id=$ACCT_LCL_WALLET&apikey=$APIKEY&amount=2&amount_exp=0&transaction_id=$TXID1" $BSERVER/distributions
curl -d "&account_id=$ACCT_EQUITY&apikey=$APIKEY&amount=-2&amount_exp=0&transaction_id=$TXID1" $BSERVER/distributions
```

Notice that in the above commands, we save the transaction ID because we needed that to create the distributions.  But we didn't save the IDs for the distributions because we just don't care.

Also note the method that Bookwerx uses to record numbers.  It uses a decimal floating-point system that enables it to _exactly_ record the numbers we see when dealing with crypto coins.

J. If we produce a balance sheet as of a suitable time, such as now, it looks as expected.  Likewise, a PNL over any timespan, such as all of time, also looks as expected.

2. Setup OKcatbox

As you have probably surmised, pretty soon we're going to need to use the OKEx API.  However, as mentioned earlier, we don't want to taunt the real API until we're fairly substantially ready to do so.  At this time we're not worthy.  So we'll use an OKCatbox instead.

We will use a public demonstration of the [OKCatbox](https://github.com/bostontrader/okcatbox) accessible at http://185.183.96.73:8090 to do this.  Let's refer to this URL as OKEXURL even though we know this is only a mimic of the same, not the real deal. In fact, let's save this as a shell variable also.

OKEXURL="http://185.183.96.73:8090"

A. As with the real API we'll need access credentials.  

```
OKEX_CREDENTIALS=okcatbox.json
curl -X POST $OKEXURL/credentials --output $OKEX_CREDENTIALS
```

Save the result body in a file of your choice.  Let's call this file OKEX_CREDENTIALS.  Please realize that the /credentials endpoint is only a convenience provided by OKCatbox.  It is not present in the real OKEx API and the credentials produced will be useless there.
