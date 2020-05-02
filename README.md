[![Build Status](https://travis-ci.org/bostontrader/okconnect.svg?branch=master)](https://travis-ci.org/bostontrader/okcatbox)
[![MIT license](http://img.shields.io/badge/license-MIT-brightgreen.svg)](http://opensource.org/licenses/MIT)

# Welcome to OKConnect

OKEx provides an API for using their service.  That's fine and dandy but in order to make effective use of this API, you will be greatly convenienced by acquiring a handful of additional tools.  More specifically:

1. [OKCatbox](https://github.com/bostontrader/okcatbox). You will need an OKEx API sandbox to play in.  Learning how to use the real OKEx API looks suspiciously close to DOS and general hackery from their point of view.  Perhaps it's better to beat a sandbox to death first, before trying to use the real OKEx API.

2. Bookwerx. You will also need some method of bookkeeping. The only reason anybody cares about an API is so that they can work the service using other software.  What do you do with OKEx?  That's right... you place orders and buy and sell things. Placing, fulfilling, and cancelling orders spawn a remarkably tedious snake nest of bookkeeping tasks.  Ignoring the bookkeeping is the method of chumps.   The information provided by the OKEx API is, ahem, less than well thought out.  It's also permeated with round-off error and various blind-spots. Dealing with the bookkeeping manually will easily drive you mad. Unless of course you have bookwerx in the arena with you.

OKConnect is the glue that binds the OKEx (or OKCatbox) API and Bookwerx together. With OKConnect, you can focus on higher-level tasks such as placing and cancelling orders, dealing with the consequences of order fulfillment, as well as reconciliation of the Bookwerx and OKEx records,  while letting the OKEx (or OKCatbox) API and Bookwerx handle the low level details.