[![Build Status](https://travis-ci.org/bostontrader/okconnect.svg?branch=master)](https://travis-ci.org/bostontrader/okconnect)
[![MIT license](http://img.shields.io/badge/license-MIT-brightgreen.svg)](http://opensource.org/licenses/MIT)

# Welcome to OKConnect

OKEx provides an API for using their service.  That's fine and dandy but in order to make effective use of this API, you will be greatly convenienced by acquiring a handful of additional tools.  More specifically:

1. [OKCatbox](https://github.com/bostontrader/okcatbox). You will need an OKEx API sandbox to play in.  Learning how to use the real OKEx API looks suspiciously close to DOS and general hackery from their point of view.  Perhaps it's better to beat a sandbox to death first, before trying to use the real OKEx API.

2. [Bookwerx](https://github.com/bostontrader/bookwerx-core-rust). You will also need some method of bookkeeping. The only reason anybody cares about an API is so that they can work the service using other software.  What do you do with OKEx?  That's right... you place orders and buy and sell things. Placing, fulfilling, and cancelling orders spawn a remarkably tedious snake nest of bookkeeping tasks.  Ignoring the bookkeeping is the method of chumps.   The bookkeeping and account balance information provided by the OKEx API is, ahem, less than well thought out.  It's also permeated with round-off error and various blind-spots. Dealing with the bookkeeping manually will easily drive you mad. Unless of course you have bookwerx in the arena with you.

OKConnect is the glue that binds the OKEx (or OKCatbox) API and Bookwerx together. With OKConnect, you can focus on higher-level tasks such as placing and cancelling orders, dealing with the consequences of order fulfillment, as well as reconciliation of the Bookwerx and OKEx records,  while letting the OKEx (or OKCatbox) API and Bookwerx handle the low level details.

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

* Access to a [Bookwerx](https://github.com/bostontrader/bookwerx-core-rust) server.

* A YML configuration file that ties these two things together.  

Unfortunately, finding and configuring all this is rather tedious.  Fortunately we have the following tutorial to walk you through an example.

In this example, we're going to figure out how to persuade OKConnect to place an order to sell 1 BTC and buy 25 BSV at the implied price of BSVBTC = 0.04.  Along the way we will also use OKConnect to compare the bookkeeping records in bookwerx with the account balances on OKCatbox.


1. Setup Bookwerx:

A. Our first order of business is to get the bookwerx bookkeeping going.  Although it is moderately elaborate to setup we present a [public demonstration version of the bookwerx UI](http://185.183.96.73:3005/).  You can follow this tutorial manually using the Bookwerx UI and it will show you the actual http requests that it uses to do its job.
  
B. Browse to http://185.183.96.73:3005/apikeys and either request a new API key or use any key that you have created in the past.
```
curl -X POST http://185.183.96.73:3003/apikeys
```
Save this as an env variable named APIKEY

C. Since we are going to use BTC and BSV in our subsequent transactions, we must first define them as currencies in bookwerx.  Do so now.

```
curl -d 'apikey=$APIKEY&rarity=0&symbol=BTC&title=Bitcoin' http://185.183.96.73:3003/currencies
curl -d "apikey=$APIKEY&rarity=0&symbol=BSV&title=Bitcoin SV" http://185.183.96.73:3003/currencies
```

D. Next, let's anticipate some accounts that we might need:  Recall that each account has a description and a currency.  Two accounts can have the same description as long as they have different currencies.

Owner's Equity	    BTC
Local Wallet        BTC

OKEx Funding	    BTC

Fee         BSV

OKEx Spot	    BTC
OKEx Spot-Hold 	BTC
OKEx Spot	    BSV

E. Next, in order to produce balance sheet and income statement style reports, we'll need to define some useful categories that we can use to tag the accounts:

Ex	Expenses
Eq	Equity
A	Assets

F. Finally, let's combine the accounts and categories:

A	OKEx Funding    BTC
A	OKEx Spot	    BTC
A	OKEx Spot-Hold  BTC
A	OKEx Spot	    BSV
A	Local Wallet	BTC

Ex	Fee	            BTC
Eq	Owner's Equity	BTC

The meaning of these accounts is probably mostly self-evident.  But we'll examine them more closely soon.

G. With these accounts let's make an initial transaction to contribute some BTC to our books:

2020-05-01T12:34:55.000Z
Initial Equity

  DR Local Wallet   BTC 2.0
  CR Owner's Equity BTC 2.0

If we produce a balance sheet as of now, you can see that everything looks as expected.  Likewise, with a PNL over any timespan.


2. Setup OKcatbox

As you have learned, pretty soon we're going to need to use the OKEx API.  As mentioned earlier, we don't want to taunt the real API until we're fairly substantially ready to do so.  At this time we're not worthy.  So we'll use the OKCatbox instead.

 We will use a public demonstration of the [OKCatbox](https://github.com/bostontrader/okcatbox) accessible at http://185.183.96.73:8090 to do this.  Let's refer to this URL as OKEXURL even though we know this is only a mimic of the same, not the real deal.

A. As with the real API we'll need access credentials.  

Using a tool of your choice, POST OKEXURL/credentials.  Save the result body in a file of your choice.  Let's call this file OKEXCredentials.  Please realize that the /credentials endpoint is only a convenience provided by OKCatbox.  It is not present in the real OKEx API and the credentials produced will be useless there.
 
3. Setup OKConnect

Establish a YML config file in a location of your choice.  Configure it thus:

bookwerx
    apikey : The API Key you acuired earlier

okex
    credentials : The full path to the file you created to hold your okex credentials.

compare
    okex
        funding
            btc : 5
            
        spot
            btc : 10

The compare key enables us to describe which accounts on OKEx correspond with which bookwerx account ids.

4. Hello compare

One important task of OKConnect is to help us reconcile our Bookwerx and OKEx records.  Using the associated APIs for each we can extract the information we need, make suitable comparisons, and illustrate any differences.  So now that everything is setup (hopefully properly) we can just ask okconnect to compare the balances!

```
okconnect -cmd compare -config myconfig
```

Tada!  Look! No differences.  At this time bookwerx and okex should both agree that there are zero balances anywhere on the OKEx service.

5. Deposit BTC with OKEx

The next step is to transfer BTC from our local wallet to OKEx.  This step involves good news and bad news.  The bad news is that making this transfer can be an insanely tedious thing to do if we contemplate the nuances too closely.  For a good time please examine our treatise on this bold assertion.  However, the good news is that we're going to make some simplifying assumptions to make this a lot easier.

5.1. Manually initiate a coin transfer using your method of choice.  If we're using OKCatbox we just pretend to do this.  If we were using the real OKEx server then we would do this for real.  Either way, we tell this to OKConnect. 
```
okconnect -cmd deposit -currency btc -amount 500 -timestamp=dlksjdlfk -addr aaa -txid tttttt 
```

5.2 Compare
```
okconnect -cmd recon
```
OKCatbox has heard about the alleged deposit but has not yet heard the broadcast transaction so no account balance info has changed. Nothing has changed in bookwerx either so everything still compares.

If you do this quickly for the real OKEx server, you get the same results.

5.3 catbox aux-api hear broadcast

```
okcatboxurl/hear-broadcast?txid=ttttt
```

This forces the catbox to hear the txid.  For the real server we don't need this.

5.4 Compare again
```
okconnect -cmd recon
```
OKCatbox has now heard about the alleged deposit as has observed the broadcast transaction.  Nothing has changed in bookwerx or OKCatbox so everything still reconciles.


If you do this quickly for the real server, you get the same results.