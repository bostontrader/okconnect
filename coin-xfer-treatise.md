- a Treatise -
    ... on the transfer of coin to OKEx specifically ...
    ... as well as any other exchange generally.

Transfering coin from a local wallet to OKEx presents a remarkable quantity of tedious considerations.  It's easy enough to do manually, but it becomes much more difficult if you try to control and monitor the process with software.  Please allow us to fret over as many of these details as we can imagine.  After that we will discuss practical methods to prune the madness.

One important part of this issue is bookkeeping.  The coin that you start with in your local wallet is an asset on your books.  At all times you must be able to verify that your local wallet balance and your bookkeeping records agree 100% with each other.  Likewise, the balance of coins you have on deposit at OKEx must also match your bookkeeping records.  

But consider what happens when you push the button on your coin client to start the transfer.

The local client will immediately report a new balance, but OKEx does not yet know anything about this.  What bookkeeping transaction should you make to record this?  Even after OKEx sees an incoming transaction its API will still not include that in its balance, even as "on hold".  Only when OKEx is fully satisified with whatever makes it happy, does the balance become available in your OKEx account.

One method to handle this situation, admittedly rather tedious, is to create three asset accounts on your books:

Local Wallet
Somewhere in Cyberspace
OKEx

In the begininng the Local Wallet has a DR balance and all the others are zero (or unchanged). Recall that at all times you can use a variety of tools of choice to determine the balances of each of these three accounts.  When you push the button on your local client to initiate the transaction, you:

DR Somewhere in Cyberspace
CR Local Wallet

This represents the initial broadcast.  Your local client's balance has declined, but OKEx doesn't know anything about this yet.  [Where's the money Lebowski?] (https://www.youtube.com/watch?v=r9twTtXkQNA&t=17)  We don't have better terminology for this so let's just say that the money is "somewhere in cyberspace".


Eventually OKEx hears about this incoming coin. Their webpage will display the incoming transaction and confirmations and the API will see a deposit.  But this amount is not yet reflected anywhere else in the OKEx API.

Finally OKEx is happy about this transfer.  The coins are no longer lost in space.

DR OKEx
CR Somewhere in Cyberspace


Another wrinkle with doing these transfers is that you must determine timestamps for your bookkeeping transaction(s). Here are some choices:

* What time does your local client say?

* What is the timestamp for the block that includes the transaction?

* What time does OKEx say?

While you're doing this, please make sure you have any timezone issues figured out as well.

If you try to record this transfer using a single bookkeeping transaction, that's a big ["does not compute"](https://www.youtube.com/watch?v=ZBAijg5Betw) because somebody's information is not going to match the single timestamp you have.  But if you've recorded the transfer using more than one transaction, as described earlier, then it's easy to record all of the correct times.

One final wrinkle with dealing these transfers is that OKEx only reports account balances upon request.  There's no push notification of deposits or changes in your balances available in their API so you have to poll it to observe the balances.


So how does OKConnect fit into all this?  That's an excellent question and we're glad you asked.

Recall that the basic purpose of OKConnect is to keep the records of OKEx and bookwerx in sync.  OKConnect can order OKEx to do something to your account or it can detect that something has happened to your account, all via the API.  It will then make suitable transactions with bookwerx. Anything that involves okex _and_ bookwerx is a good task for OKConnect.  But anything that involves okex _xor_ bookwerx is a task for other tools.

Dealing with these transfers gives us a messy boundary that we have to tread carefully.

In order to send coin to OKEx we must first determine a receiving address at OKEx.  But determining the address only, is a task for other tools (such as OKProbe) because bookwerx doesn't care about just an address.

After we get an address we must then coax whatever coin client we're using to send the coins.  There's no way OKConnect can deal with the bazillion different coin clients available and we're not even going to try.

So... [Drumroll please...](https://www.youtube.com/watch?v=-R81ugVBLlw&t=9)

After a suitable amount of wringing-of-hands and gnashing-of-teeth, we've come to the following reasonable practical policy:

1. Using a method of your choosing, determine a suitable address.  Perhaps use OKProbe.

2. Using a method of your choosing, such as your coin client, initiate the transfer.

3. When you see this transaction in a block, use the timestamp of the block and:

okconnect -cmd deposit -currency btc -crlocal 500 -drok 499 -drfee 1 -timestamp=blocktime -poll 60

This will assert that a transaction has been made to deposit btc with okex, at the given time.  You will have to manually figure out how much the fee is and who pays it.  Notice the "dr" and "cr" in the option names.  These are clues to their meaning.  OKConnect will then poll the server every 60 seconds looking for the balance to increase by the amount specified by drok.

4. After okconnect sees this, it will:

DR- OKEx
CR- Local wallet

Notice that we are skipping the "somewhere in cyberspace" phase. That's just too tedious.

5. If okconnect is interrupted before the polling detects the increase in balance, then you just have to make the bookwerx transaction by hand.






Any attempt to reconcile okex and bookwerx will fail 
3.  be done 


Dealing with a suspense account is too tedious for the vast majority of real people doing this.




  okconnect and okcatbox draw the line here.  Fetching addresses and doing whatever you do with BTC clients is outside the scope of OKConnect. OKConnect focuses on making OKEX and bookwerx work in sync.  
Another twisted tale to contemplate:  We start with coins that are safely in our local wallet.  Bookwerx records this as an asset.  When you transfer coins to okex, where should this asset be recorded?  That depends upon how tedious you wish to get.  Some factors to consider:

1. When you first push the button to send, the transaction is broadcast.  Your local wallet shows an immediate reduction in balance, but okex does not show any increase.  [Where did the money go?](https://www.youtube.com/watch?v=fWESB56wcxY&t=18)  You could create a "transiting through cyberspace" asset account if you wish and debit that.  That strikes me as a bit much, but each to his own.

2. Eventually OKEx hears about this and it will report said transfer as on hold.  It knows the money is on the way but it doesn't have enough confirmation yet.  IMHO, this is a more reasonable time to actually make the DR/CR transaction.  Because you've initiated the transaction w/o using OKConnect.





 Here's how to simulate this:
,
5.1 Get okex address via the catbox using okprobe.

5.2 Invoke the pseudo API deposit on the catbox, with status = pending.  Poof! Deposit made, but pending.  Problem solved :-)





	C. Eventually OKEx verifies the transfer.  

	At that point 
		DR OKEx Deposit BTC
		CR OKEx Deposit-Pending BTC

Then we can then successfully reconcile. 


6. Xfer BTC from Deposit to Spot

In order to do any wheelin' 'n' dealin' we'll have to transfer some BTC from the deposit to the spot account.
okconnect -xferfrom=deposit -xferto=spot -currency=BTC -quan=all|nnn

If the API call fails, nothing is transfered and nothing happens.

If the API call succeeds then:

	DR Spot BTC
	CR Deposit BTC

Be sure to run the auto reconciler to ensure that all is still well.


7. Place maker order to trade BTC/BSV

okconnect -sell=BTC -buy=BSV -sellquan=lll -buyquan=lkjlkj -
Now let's submit an order to sell some of this BTC.

If the API call fails, then no order is placed and nothing happens.

If the API calls succeeds then we have a new order

	DR Spot-Hold BTC
	CR Spot      BTC

Because this is a maker order the order should appear in the order book but it won't execute immediately.

We should see this new order on the websocket channel.

We should see this new order in any other place where we might see this order.

reconcile still works


8. Fulfill the Order



 
2. The OKEx sandbox is open oure and available on github.  It's also rather elaborate to setup so we present a demonstration version here.

OkEx requires credentials.  For our sandbox, just push the btton here to get them.

the sandbox needs some test data to get started with.  We provide a template.  You can download, modify, and upload if you like.



1. init bookwerx with:

1a. currency BTC and BSV

1b. spot, BTC and spot BSV, begining equity BTC

1c. d equity 1 BTC


1. Init sandbox with test data because we need a market that supports this.

2. configure sandbox credential.

3. execute the API call to set the order.

4. look at the pnl and bs.


